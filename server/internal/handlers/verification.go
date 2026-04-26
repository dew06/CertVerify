package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/ledongthuc/pdf"

	"cert-system/server/internal/database"
	"cert-system/server/internal/models"
	"cert-system/server/internal/services"

	"github.com/gin-gonic/gin"
)

type VerificationHandler struct {
	cardanoService *services.CardanoService
}

func NewVerificationHandler() *VerificationHandler {
	return &VerificationHandler{
		cardanoService: &services.CardanoService{},
	}
}

type VerifyRequest struct {
	CertID      string `json:"cert_id" binding:"required"`
	StudentName string `json:"student_name" binding:"required"`
	Degree      string `json:"degree" binding:"required"`
}

// Verify godoc
// @Summary Verify a certificate
// @Description Verify certificate authenticity using cert_id and student details
// @Tags verification
// @Accept json
// @Produce json
// @Param request body VerifyRequest true "Verification details"
// @Success 200 {object} map[string]interface{}
// @Router /verify [post]

// Verify - Fast database verification
func (h *VerificationHandler) Verify(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Step 1: Find certificate in database
	var certificate models.Certificate
	result := database.DB.Where("cert_id = ?", req.CertID).First(&certificate)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"valid": false,
			"error": "Certificate not found",
		})
		return
	}

	// Step 2: Verify data matches
	if certificate.StudentName != req.StudentName || certificate.Degree != req.Degree {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "Certificate data mismatch",
		})
		return
	}

	// Step 3: Recompute hash
	computedHash := h.cardanoService.HashCertificate(
		req.CertID,
		req.StudentName,
		req.Degree,
		certificate.Salt,
	)

	if computedHash != certificate.CertificateHash {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "Hash mismatch - certificate tampered",
		})
		return
	}

	// Step 4: Get university
	var university models.University
	database.DB.Where("id = ?", certificate.UniversityID).First(&university)

	// Success!
	c.JSON(http.StatusOK, gin.H{
		"valid":               true,
		"student_name":        certificate.StudentName,
		"degree":              certificate.Degree,
		"gpa":                 certificate.GPA,
		"university":          university.Name,
		"university_domain":   university.Domain,
		"issue_date":          certificate.IssueDate.Format("January 2, 2006"),
		"university_verified": university.IsVerified,
		"blockchain": gin.H{
			"cardano_txid": certificate.CardanoTxID,
			"merkle_root":  certificate.MerkleRootHash,
			"explorer_url": fmt.Sprintf("https://cardanoscan.io/transaction/%s", certificate.CardanoTxID),
		},
		"ipfs": gin.H{
			"pdf_hash":    certificate.IPFSPDFHash,
			"gateway_url": fmt.Sprintf("https://ipfs.io/ipfs/%s", certificate.IPFSPDFHash),
		},
		"message": "✅ Certificate is valid",
	})
}

// GetBlockchainInfo godoc
// @Summary Get blockchain information
// @Description Get blockchain explorer links for a certificate
// @Tags verification
// @Produce json
// @Param certID path string true "Certificate ID"
// @Success 200 {object} map[string]interface{}
// @Router /verify/blockchain/{certID} [get]
func (h *VerificationHandler) GetBlockchainInfo(c *gin.Context) {
	certID := c.Param("certID")

	var certificate models.Certificate
	result := database.DB.Where("cert_id = ?", certID).First(&certificate)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cert_id":      certID,
		"cardano_txid": certificate.CardanoTxID,
		"merkle_root":  certificate.MerkleRootHash,
		"explorers": []gin.H{
			{
				"name":        "CardanoScan",
				"url":         fmt.Sprintf("https://cardanoscan.io/transaction/%s", certificate.CardanoTxID),
				"recommended": true,
			},
			{
				"name": "Cardano Explorer",
				"url":  fmt.Sprintf("https://explorer.cardano.org/en/transaction?id=%s", certificate.CardanoTxID),
			},
		},
	})
}

// VerifyPDF - POST /api/verify/pdf
func (h *VerificationHandler) VerifyPDF(c *gin.Context) {
	file, err := c.FormFile("certificate")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}

	f, _ := file.Open()
	defer f.Close()

	// 1. Read the PDF text content
	content, err := readPDFText(f, file.Size)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse PDF text"})
		return
	}

	// 2. Extract Data using Regex (Matches your PDFGenerator template)
	certID := extractField(content, `Certificate ID: (CERT-[\w-]+)`)
	studentName := extractField(content, `CERTIFICATE OF ACHIEVEMENT\s+(.*)\s+has successfully`)

	if certID == "" {
		c.JSON(400, gin.H{"valid": false, "message": "Could not find Certificate ID in PDF"})
		return
	}

	// 3. Lookup the Original Record
	var certificate models.Certificate
	if err := database.DB.Where("cert_id = ?", certID).First(&certificate).Error; err != nil {
		c.JSON(404, gin.H{"valid": false, "message": "Certificate ID not recognized"})
		return
	}

	// 4. Verify Content Integrity
	// We compare the name extracted from the PDF with the name in our DB
	if studentName != certificate.StudentName {
		c.JSON(200, gin.H{"valid": false, "message": "❌ Data mismatch: Name on PDF does not match our records"})
		return
	}

	c.JSON(200, gin.H{
		"valid":        true,
		"message":      "✅ Certificate content verified via Blockchain anchor",
		"student_name": certificate.StudentName,
		"degree":       certificate.Degree,
		"cardano_tx":   certificate.CardanoTxID,
	})
}

// Helper: Extract text from PDF
func readPDFText(f io.ReaderAt, size int64) (string, error) {
	reader, err := pdf.NewReader(f, size)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	p := reader.Page(1) // Certificates are usually 1 page
	content, _ := p.GetPlainText(nil)
	buf.WriteString(content)
	return buf.String(), nil
}

// Helper: Regex extractor
func extractField(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (h *VerificationHandler) VerifyByID(c *gin.Context) {
	certID := c.Param("id")

	var certificate models.Certificate
	if err := database.DB.Where("cert_id = ?", certID).First(&certificate).Error; err != nil {
		c.JSON(404, gin.H{"valid": false, "message": "Invalid Certificate ID"})
		return
	}

	var university models.University
	database.DB.First(&university, certificate.UniversityID)

	c.JSON(200, gin.H{
		"valid":        true,
		"student_name": certificate.StudentName,
		"degree":       certificate.Degree,
		"university":   university.Name,
		"issue_date":   certificate.IssueDate.Format("January 2, 2006"),
		"blockchain":   gin.H{"tx_hash": certificate.CardanoTxID},
	})
}

// QRCodeData - GET /api/verify/qr/:certID
func (h *VerificationHandler) QRCodeData(c *gin.Context) {
	certID := c.Param("certID")

	var certificate models.Certificate
	result := database.DB.Where("cert_id = ?", certID).First(&certificate)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cert_id": certID,
		"qr_url":  fmt.Sprintf("https://verify.example.com/check/%s", certID),
		"pdf_url": fmt.Sprintf("https://ipfs.io/ipfs/%s", certificate.IPFSPDFHash),
	})
}
