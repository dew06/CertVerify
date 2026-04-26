package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BulkDownloadCertificatesHandler allows universities to download all their certificates
// @Summary Bulk download certificates
// @Description Download all certificates for a university as a ZIP file
// @Tags Certificates
// @Accept json
// @Produce application/zip
// @Param university_id query string true "University ID"
// @Param password query string true "Password"
// @Success 200 {file} application/zip
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /certificates/bulk-download [get]
func BulkDownloadCertificatesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		universityID := c.Query("university_id")
		password := c.Query("password")

		if universityID == "" || password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "university_id and password required"})
			return
		}

		// Get university
		var university models.University
		if err := db.Where("id = ?", universityID).First(&university).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
			return
		}

		// Verify password
		if !university.CheckPassword(password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}

		// Get all certificates for this university
		var certificates []models.Certificate
		if err := db.Where("university_id = ?", universityID).Find(&certificates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch certificates"})
			return
		}

		if len(certificates) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No certificates found"})
			return
		}

		log.Printf("📦 Creating ZIP with %d certificates for %s", len(certificates), university.Name)

		// Create temporary ZIP file
		zipFilename := fmt.Sprintf("certificates_%s_%d.zip", universityID, len(certificates))
		zipPath := filepath.Join(os.TempDir(), zipFilename)

		zipFile, err := os.Create(zipPath)
		if err != nil {
			log.Printf("❌ Failed to create ZIP: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ZIP file"})
			return
		}
		defer os.Remove(zipPath) // Clean up after sending

		// Create ZIP writer
		zipWriter := zip.NewWriter(zipFile)

		successCount := 0
		for _, cert := range certificates {
			// Check if PDF exists
			if cert.PDFPath == "" {
				log.Printf("⚠️  Certificate %s has no PDF", cert.ID)
				continue
			}

			pdfPath := filepath.Join(".", cert.PDFPath)

			// Check if file exists
			if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
				log.Printf("⚠️  PDF not found: %s", pdfPath)
				continue
			}

			// Open PDF file
			pdfFile, err := os.Open(pdfPath)
			if err != nil {
				log.Printf("⚠️  Failed to open PDF %s: %v", pdfPath, err)
				continue
			}

			// Create entry in ZIP
			filename := fmt.Sprintf("%s_%s_%s.pdf",
				cert.StudentName,
				cert.Degree,
				cert.ID[:8])

			zipEntry, err := zipWriter.Create(filename)
			if err != nil {
				pdfFile.Close()
				log.Printf("⚠️  Failed to create ZIP entry: %v", err)
				continue
			}

			// Copy PDF to ZIP
			if _, err := io.Copy(zipEntry, pdfFile); err != nil {
				pdfFile.Close()
				log.Printf("⚠️  Failed to copy PDF to ZIP: %v", err)
				continue
			}

			pdfFile.Close()
			successCount++
		}

		// Add summary file
		summary := fmt.Sprintf(`Certificate Archive Summary
==============================
University: %s
Total Certificates: %d
Included in ZIP: %d
Generated: %s

This archive contains all certificates issued by %s.
Each certificate is blockchain-verified and can be authenticated at:
https://your-domain.com/verify
`, university.Name, len(certificates), successCount,
			fmt.Sprintf("%v", os.Getenv("GENERATED_DATE")), university.Name)

		summaryEntry, _ := zipWriter.Create("README.txt")
		summaryEntry.Write([]byte(summary))

		// Close ZIP
		if err := zipWriter.Close(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize ZIP"})
			return
		}

		zipFile.Close()

		log.Printf("✅ ZIP created with %d certificates", successCount)

		// Send ZIP file
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilename))
		c.Header("Content-Type", "application/zip")
		c.File(zipPath)
	}
}

// DownloadSingleCertificateHandler downloads a single certificate PDF
// @Summary Download certificate PDF
// @Description Download a single certificate PDF by ID
// @Tags Certificates
// @Produce application/pdf
// @Param id path string true "Certificate ID"
// @Success 200 {file} application/pdf
// @Failure 404 {object} map[string]interface{}
// @Router /certificates/{id}/download [get]
func DownloadSingleCertificateHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		certID := c.Param("id")

		var certificate models.Certificate
		if err := db.Where("id = ?", certID).First(&certificate).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
			return
		}

		if certificate.PDFPath == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "PDF not available"})
			return
		}

		pdfPath := filepath.Join(".", certificate.PDFPath)

		if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "PDF file not found"})
			return
		}

		filename := fmt.Sprintf("certificate_%s_%s.pdf", certificate.StudentName, certificate.ID[:8])

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", "application/pdf")
		c.File(pdfPath)
	}
}
