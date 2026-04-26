package handlers

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"cert-system/server/internal/database"
	"cert-system/server/internal/ipfs"
	"cert-system/server/internal/models"
	"cert-system/server/internal/services"

	"github.com/gin-gonic/gin"
)

// BatchIssueCSV godoc
// @Summary Bulk issue certificates from CSV
// @Description Upload CSV file to issue multiple certificates at once
// @Tags certificates
// @Accept multipart/form-data
// @Produce json
// @Param csv_file formData file true "CSV file with student data"
// @Param university_id formData string true "University ID"
// @Param password formData string true "University password"
// @Success 200 {object} map[string]interface{}
// @Router /certificates/batch-csv [post]
func (h *CertificateHandler) BatchIssueCSV(c *gin.Context) {
	// Get uploaded file
	file, err := c.FormFile("csv_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No CSV file uploaded"})
		return
	}

	// Get form data
	universityID := c.PostForm("university_id")
	password := c.PostForm("password")

	if universityID == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing university_id or password"})
		return
	}

	// Verify university
	var university models.University
	result := database.DB.Where("id = ?", universityID).First(&university)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
		return
	}

	if !university.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Domain not verified"})
		return
	}

	// Open CSV file
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot read CSV file"})
		return
	}
	defer fileContent.Close()

	// Parse CSV
	reader := csv.NewReader(fileContent)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format"})
		return
	}

	// Validate headers
	// expectedHeaders := []string{"student_name", "degree", "gpa", "email"}
	if len(headers) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "CSV must have at least student_name and degree columns",
		})
		return
	}

	// Process rows
	var students []struct {
		StudentName string
		Degree      string
		GPA         float64
		Email       string
	}

	lineNumber := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Error reading CSV at line %d: %s", lineNumber, err.Error()),
			})
			return
		}
		lineNumber++

		if len(record) < 2 {
			continue // Skip empty rows
		}

		studentName := record[0]
		degree := record[1]

		var gpa float64
		if len(record) > 2 && record[2] != "" {
			gpa, _ = strconv.ParseFloat(record[2], 64)
		}

		var email string
		if len(record) > 3 {
			email = record[3]
		}

		students = append(students, struct {
			StudentName string
			Degree      string
			GPA         float64
			Email       string
		}{
			StudentName: studentName,
			Degree:      degree,
			GPA:         gpa,
			Email:       email,
		})
	}

	if len(students) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid student records found in CSV"})
		return
	}

	// Limit check
	if len(students) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Maximum 1,000 certificates per batch. Please split into multiple files.",
		})
		return
	}

	// Process certificates in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	issued := []string{}
	failed := []string{}

	// Use worker pool (10 concurrent workers)
	semaphore := make(chan struct{}, 10)

	for i, student := range students {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire

		go func(index int, s struct {
			StudentName string
			Degree      string
			GPA         float64
			Email       string
		}) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release

			// Generate certificate ID
			certIDBytes := make([]byte, 16)
			rand.Read(certIDBytes)
			certID := hex.EncodeToString(certIDBytes)

			// Generate salt
			saltBytes := make([]byte, 32)
			rand.Read(saltBytes)
			salt := hex.EncodeToString(saltBytes)

			// Generate PDF
			pdfData := services.CertificateData{
				CertID:           certID,
				StudentName:      s.StudentName,
				Degree:           s.Degree,
				University:       university.Name,
				UniversityDomain: university.Domain,
				IssueDate:        time.Now(),
				GPA:              s.GPA,
			}

			pdfBytes, err := h.pdfGenerator.GenerateCertificatePDF(pdfData)
			if err != nil {
				mu.Lock()
				failed = append(failed, fmt.Sprintf("%s (PDF generation failed)", s.StudentName))
				mu.Unlock()
				return
			}

			// Compute hashes
			pdfHash := h.pdfGenerator.ComputePDFHash(pdfBytes)
			certHash := h.cardanoService.HashCertificate(certID, s.StudentName, s.Degree, salt)

			// Upload PDF to IPFS
			ipfsPDFHash, err := h.ipfsClient.UploadPDF(pdfBytes)
			if err != nil {
				mu.Lock()
				failed = append(failed, fmt.Sprintf("%s (IPFS upload failed)", s.StudentName))
				mu.Unlock()
				return
			}

			// Save to database
			certificate := models.Certificate{
				CertID:          certID,
				UniversityID:    university.ID,
				StudentName:     s.StudentName,
				Degree:          s.Degree,
				GPA:             s.GPA,
				IssueDate:       time.Now(),
				CertificateHash: certHash,
				PDFHash:         pdfHash,
				Salt:            salt,
				IPFSPDFHash:     ipfsPDFHash,
				CardanoTxID:     "",
				StudentEmail:    s.Email,
			}

			mu.Lock()
			database.DB.Create(&certificate)
			issued = append(issued, certID)
			mu.Unlock()

		}(i, student)
	}

	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"total":       len(students),
		"issued":      len(issued),
		"failed":      len(failed),
		"cert_ids":    issued,
		"failed_list": failed,
		"message":     fmt.Sprintf("✅ %d certificates created. Use batch-anchor endpoint to write to blockchain.", len(issued)),
		"next_step":   "POST /api/certificates/batch-anchor",
	})
}

type CertificateHandler struct {
	cardanoService *services.CardanoService
	ipfsClient     *ipfs.Client
	pdfGenerator   *services.PDFGeneratorService
	qrService      *services.QRCodeService
}

func NewCertificateHandler(ipfsURL string) *CertificateHandler {
	baseVerificationURL := os.Getenv("VERIFICATION_URL")
	if baseVerificationURL == "" {
		baseVerificationURL = "https://localhost:5173"
	}

	qrService := services.NewQRCodeService(baseVerificationURL)
	pdfGenerator := services.NewPDFGeneratorService(qrService)

	return &CertificateHandler{
		cardanoService: &services.CardanoService{},
		ipfsClient:     ipfs.NewClient(ipfsURL),
		pdfGenerator:   pdfGenerator,
		qrService:      qrService,
	}
}

// IssueCertificate - POST /api/certificates/issue
type IssueCertificateRequest struct {
	UniversityID string  `json:"university_id" binding:"required"`
	StudentName  string  `json:"student_name" binding:"required"`
	StudentEmail string  `json:"student_email" binding:"required,email"`
	Degree       string  `json:"degree" binding:"required"`
	GPA          float64 `json:"gpa"`
	Password     string  `json:"password" binding:"required"`
}

// Issue godoc
// @Summary Issue a single certificate
// @Description Issue a certificate for a student with PDF and IPFS storage
// @Tags certificates
// @Accept json
// @Produce json
// @Param request body IssueCertificateRequest true "Certificate details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /certificates/issue [post]
func (h *CertificateHandler) Issue(c *gin.Context) {
	var req IssueCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Validate university
	var university models.University
	result := database.DB.Where("id = ?", req.UniversityID).First(&university)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
		return
	}

	if !university.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Domain not verified"})
		return
	}

	// 2. Generate IDs
	certIDBytes := make([]byte, 16)
	rand.Read(certIDBytes)
	certID := hex.EncodeToString(certIDBytes)

	saltBytes := make([]byte, 32)
	rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)

	// 3. Generate PDF with QR code
	pdfData := services.CertificateData{
		CertID:           certID,
		StudentName:      req.StudentName,
		Degree:           req.Degree,
		University:       university.Name,
		UniversityDomain: university.Domain,
		IssueDate:        time.Now(),
		GPA:              req.GPA,
	}

	pdfBytes, err := h.pdfGenerator.GenerateCertificatePDF(pdfData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed"})
		return
	}

	// 4. Compute hashes
	pdfHash := h.pdfGenerator.ComputePDFHash(pdfBytes)
	certHash := h.cardanoService.HashCertificate(certID, req.StudentName, req.Degree, salt)

	// 5. Upload PDF to IPFS (ONLY PDF, no metadata!)
	ipfsPDFHash, err := h.ipfsClient.UploadPDF(pdfBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "IPFS upload failed. Is IPFS Desktop running?",
		})
		return
	}

	// 6. Save to PostgreSQL (ALL metadata here!)
	certificate := models.Certificate{
		CertID:          certID,
		UniversityID:    university.ID,
		StudentName:     req.StudentName,
		Degree:          req.Degree,
		GPA:             req.GPA,
		IssueDate:       time.Now(),
		CertificateHash: certHash,
		PDFHash:         pdfHash,
		Salt:            salt,
		IPFSPDFHash:     ipfsPDFHash,
		StudentEmail:    req.StudentEmail,
	}

	database.DB.Create(&certificate)

	// 7. Return response
	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"cert_id":       certID,
		"ipfs_pdf_hash": ipfsPDFHash,
		"download_url":  fmt.Sprintf("%s/api/certificates/%s/download", os.Getenv("FRONTEND_URL"), certID),
		"verify_url":    fmt.Sprintf("%s/check/%s", os.Getenv("VERIFICATION_URL"), certID),
		"message":       "✅ Certificate created! PDF stored on IPFS.",
	})
}

// Get godoc
// @Summary Get certificate details
// @Description Retrieve certificate information by ID
// @Tags certificates
// @Produce json
// @Param certID path string true "Certificate ID"
// @Success 200 {object} map[string]interface{}
// @Router /certificates/{certID} [get]
func (h *CertificateHandler) Get(c *gin.Context) {
	certID := c.Param("certID")

	var certificate models.Certificate
	result := database.DB.Where("cert_id = ?", certID).First(&certificate)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
		return
	}

	var university models.University
	database.DB.Where("id = ?", certificate.UniversityID).First(&university)

	// Return everything from PostgreSQL (no IPFS fetch needed!)
	c.JSON(http.StatusOK, gin.H{
		"cert_id":             certificate.CertID,
		"student_name":        certificate.StudentName,
		"degree":              certificate.Degree,
		"gpa":                 certificate.GPA,
		"university":          university.Name,
		"university_domain":   university.Domain,
		"issue_date":          certificate.IssueDate,
		"cardano_txid":        certificate.CardanoTxID,
		"ipfs_pdf_hash":       certificate.IPFSPDFHash,
		"university_verified": university.IsVerified,
	})
}

// DownloadPDF godoc
// @Summary Download certificate PDF
// @Description Download the PDF certificate from IPFS
// @Tags certificates
// @Produce application/pdf
// @Param certID path string true "Certificate ID"
// @Success 200 {file} []byte
// @Router /certificates/{certID}/download [get]
func (h *CertificateHandler) DownloadPDF(c *gin.Context) {
	certID := c.Param("certID")

	var certificate models.Certificate
	result := database.DB.Where("cert_id = ?", certID).First(&certificate)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
		return
	}

	// Fetch PDF from IPFS
	pdfBytes, err := h.ipfsClient.DownloadPDF(certificate.IPFSPDFHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to fetch from IPFS",
			"ipfs_gateway": fmt.Sprintf("https://ipfs.io/ipfs/%s", certificate.IPFSPDFHash),
		})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(
		"attachment; filename=%s_%s.pdf",
		certificate.StudentName,
		certificate.CertID[:8],
	))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
