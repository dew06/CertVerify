package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cert-system/server/internal/database"
	"cert-system/server/internal/middleware"
	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
)

// UploadMyCertificate godoc
// @Summary Upload a certificate PDF
// @Description Student uploads their certificate PDF to link it to their profile and auto-create an education entry
// @Tags student
// @Accept multipart/form-data
// @Produce json
// @Param certificate formData file true "Certificate PDF file"
// @Success 200 {object} map[string]interface{}
// @Router /student/upload-certificate [post]
func UploadMyCertificate(c *gin.Context) {
	// ── 1. Authenticate student ───────────────────────────────────────────
	_, email, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// ── 2. Validate file upload ───────────────────────────────────────────
	fileHeader, err := c.FormFile("certificate")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded. Please attach a PDF file using the 'certificate' field.",
		})
		return
	}

	if filepath.Ext(fileHeader.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are accepted."})
		return
	}

	const maxSize = 10 << 20 // 10 MB
	if fileHeader.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Maximum allowed size is 10 MB."})
		return
	}

	// ── 3. Open file ──────────────────────────────────────────────────────
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file."})
		return
	}
	defer file.Close()

	// ── 4. Extract cert ID from PDF text ──────────────────────────────────
	// readPDFText and extractField are defined in verify.go (same package)
	content, err := readPDFText(file, fileHeader.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": "Could not read PDF content. Make sure you are uploading a valid CertChain-issued certificate.",
		})
		return
	}

	// PDF template uses "Verification ID:" label with raw hex value
	certID := extractField(content, `Verification ID:\s*([a-f0-9]+)`)
	if certID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": "Could not find a Verification ID in this PDF. Make sure you are uploading a CertChain-issued certificate.",
		})
		return
	}

	log.Printf("📄 Student %s attempting to upload certificate %s", email, certID)

	// ── 5. Look up certificate in database ────────────────────────────────
	var cert models.Certificate
	if err := database.DB.Where("cert_id = ?", certID).First(&cert).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"valid": false,
			"error": "Certificate ID not recognised. This certificate may not have been issued through CertChain.",
		})
		return
	}

	// ── 6. Verify ownership — cert must belong to the logged-in student ───
	if cert.StudentEmail != email {
		log.Printf("⚠️  Ownership mismatch: cert %s belongs to %s, uploaded by %s", certID, cert.StudentEmail, email)
		c.JSON(http.StatusForbidden, gin.H{
			"valid": false,
			"error": "This certificate was not issued to your account.",
		})
		return
	}

	// ── 7. Save PDF to disk ───────────────────────────────────────────────
	uploadDir := "./uploads/student-certificates"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory."})
		return
	}

	filename := fmt.Sprintf("%s_%d.pdf", certID, time.Now().UnixMilli())
	savePath := filepath.Join(uploadDir, filename)

	// Seek back to start before saving (readPDFText consumed the reader)
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file."})
		return
	}
	defer out.Close()
	io.Copy(out, file)

	// ── 8. Update pdf_path on the certificate record if not already set ───
	if cert.PDFPath == "" {
		database.DB.Model(&cert).Update("pdf_path", savePath)
	}

	// ── 9. Fetch issuing university ───────────────────────────────────────
	var university models.University
	database.DB.Where("id = ?", cert.UniversityID).First(&university)

	// ── 10. Look up student record to get their primary key ───────────────
	var student models.Student
	if err := database.DB.Where("email = ?", email).First(&student).Error; err != nil {
		log.Printf("⚠️  Could not find student record for %s: %v", email, err)
	}

	// ── 11. Auto-create education entry if not already present ────────────
	//
	// Prevents duplicates if the student re-uploads the same certificate.
	// Matches on student_id + degree + university_id.
	var existingEdu models.StudentEducation
	educationCreated := false

	notFound := database.DB.Where(
		"student_id = ? AND degree = ? AND university_id = ?",
		student.ID, cert.Degree, cert.UniversityID,
	).First(&existingEdu).Error

	if notFound != nil {
		issueDate := cert.IssueDate
		newEdu := models.StudentEducation{
			StudentID:    student.ID,
			Degree:       cert.Degree,
			FieldOfStudy: cert.Degree, // cert stores degree only; use as field_of_study fallback
			GPA:          cert.GPA,
			EndDate:      &issueDate,
			IsCurrent:    false,
			UniversityID: &cert.UniversityID,
		}
		if err := database.DB.Create(&newEdu).Error; err != nil {
			log.Printf("⚠️  Failed to auto-create education for cert %s: %v", certID, err)
		} else {
			educationCreated = true
			log.Printf("✅ Auto-created education entry for %s (cert %s)", email, certID)
		}
	} else {
		log.Printf("ℹ️  Education entry already exists for %s / %s", email, cert.Degree)
	}

	log.Printf("✅ Student %s successfully uploaded certificate %s", email, certID)

	// ── 12. Build blockchain info ─────────────────────────────────────────
	isAnchored := cert.CardanoTxID != "" && cert.BlockchainStatus == "anchored"
	explorerURL := ""
	if isAnchored {
		network := os.Getenv("CARDANO_NETWORK")
		if network == "" {
			network = "preview"
		}
		explorerURL = fmt.Sprintf("https://%s.cardanoscan.io/transaction/%s", network, cert.CardanoTxID)
	}

	// ── 13. Respond ───────────────────────────────────────────────────────
	c.JSON(http.StatusOK, gin.H{
		"valid":               true,
		"message":             "✅ Certificate verified and linked to your account.",
		"cert_id":             cert.CertID,
		"student_name":        cert.StudentName,
		"degree":              cert.Degree,
		"gpa":                 cert.GPA,
		"issue_date":          cert.IssueDate.Format("January 2, 2006"),
		"university":          university.Name,
		"university_domain":   university.Domain,
		"university_verified": university.IsVerified,
		"blockchain": gin.H{
			"status":       cert.BlockchainStatus,
			"anchored":     isAnchored,
			"cardano_txid": cert.CardanoTxID,
			"explorer_url": explorerURL,
		},
		"ipfs": gin.H{
			"pdf_hash":    cert.IPFSPDFHash,
			"gateway_url": fmt.Sprintf("https://ipfs.io/ipfs/%s", cert.IPFSPDFHash),
		},
		"email_sent":        cert.EmailSent,
		"education_created": educationCreated,
	})
}
