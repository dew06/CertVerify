package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"cert-system/server/internal/database"
	"cert-system/server/internal/ipfs"
	"cert-system/server/internal/merkle"
	"cert-system/server/internal/models"
	"cert-system/server/internal/services"

	"github.com/gin-gonic/gin"
)

type BatchAnchorHandler struct {
	cardanoService *services.CardanoProductionService
}

func NewBatchAnchorHandler() *BatchAnchorHandler {
	network := os.Getenv("CARDANO_NETWORK")
	if network == "" {
		network = "preview"
	}
	apiKey := os.Getenv("BLOCKFROST_API_KEY")

	return &BatchAnchorHandler{
		cardanoService: services.NewCardanoProductionService(network, apiKey),
	}
}

type BatchAnchorRequest struct {
	UniversityID string `json:"university_id" binding:"required"`
	Password     string `json:"password" binding:"required"`
	MaxCerts     int    `json:"max_certs"` // Optional, default 1000
	SendEmails   bool   `json:"send_emails"`
}

type BatchAnchorResponse struct {
	Success      bool     `json:"success" example:"true"`
	Certificates int      `json:"certificates" example:"10"`
	MerkleRoot   string   `json:"merkle_root" example:"abc123def456..."`
	CardanoTxID  string   `json:"cardano_txid" example:"f75eac8ac7b3..."`
	ExplorerURL  string   `json:"explorer_url" example:"https://preview.cardanoscan.io/transaction/..."`
	Cost         string   `json:"cost" example:"0.17 ADA (~$0.08)"`
	CostPerCert  string   `json:"cost_per_cert" example:"$0.008000"`
	Message      string   `json:"message" example:"✅ 10 certificates anchored and 10 emails sent!"`
	Wallet       string   `json:"wallet" example:"addr_test1..."`
	Network      string   `json:"network" example:"preview"`
	EmailsSent   int      `json:"emails_sent" example:"10"`
	EmailErrors  []string `json:"email_errors,omitempty"`
	EmailWarning string   `json:"email_warning,omitempty" example:"2 emails failed to send"`
}

// AnchorBatch anchors certificates to blockchain and optionally sends emails
// @Summary Anchor certificates to blockchain
// @Description Anchor multiple certificates to blockchain using Merkle tree and optionally send emails to students
// @Tags Certificates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchAnchorRequest true "Batch anchor request"
// @Success 200 {object} BatchAnchorResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /certificates/batch-anchor [post]
func (h *BatchAnchorHandler) AnchorBatch(c *gin.Context) {
	var req BatchAnchorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MaxCerts == 0 {
		req.MaxCerts = 1000 // Default
	}

	// Get university
	var university models.University
	result := database.DB.Where("id = ?", req.UniversityID).First(&university)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
		return
	}

	// Check if university is verified
	if !university.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Domain not verified",
			"message": "Please verify your domain before anchoring certificates",
		})
		return
	}

	// Find unanchored certificates for this university
	var certificates []models.Certificate
	database.DB.Where("university_id = ? AND cardano_tx_id = ?", req.UniversityID, "").
		Limit(req.MaxCerts).
		Find(&certificates)

	if len(certificates) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No certificates need anchoring",
			"info":    "All certificates are already anchored to blockchain",
		})
		return
	}

	log.Printf("📋 Found %d certificates to anchor for %s", len(certificates), university.Name)

	// Collect certificate hashes
	var hashes []string
	for _, cert := range certificates {
		hashes = append(hashes, cert.CertificateHash)
	}

	// Build Merkle tree
	tree, err := merkle.BuildTree(hashes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Merkle tree creation failed",
		})
		return
	}

	merkleRoot := tree.Root.Hash

	// Generate Merkle proofs for each certificate
	for i, cert := range certificates {
		proof, err := merkle.GenerateProof(tree, i)
		if err != nil {
			continue
		}

		proofJSON, _ := merkle.ProofToJSON(proof)

		// Update certificate with proof
		database.DB.Model(&cert).Updates(map[string]interface{}{
			"merkle_root_hash": merkleRoot,
			"merkle_proof":     proofJSON,
		})
	}

	// Test password by trying to decrypt (this is the only verification needed)
	_, err = services.DecryptPrivateKey(
		university.EncryptedPrivateKey,
		req.Password,
		university.EncryptionSalt,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid password",
			"message": "Failed to decrypt wallet keys. Please check your password.",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Password verified, wallet decrypted")

	// Submit to blockchain
	log.Printf("🚀 Submitting to Cardano blockchain...")
	txID, err := h.cardanoService.SubmitTransaction(
		university.CardanoPublicKey,
		university.EncryptedPrivateKey,
		req.Password,
		university.EncryptionSalt,
		merkleRoot,
		len(certificates),
		university.Name,
	)
	if err != nil {
		log.Printf("❌ Blockchain submission failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Blockchain submission failed",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Transaction submitted: %s", txID)

	// Update all certificates with transaction ID
	database.DB.Model(&models.Certificate{}).
		Where("merkle_root_hash = ?", merkleRoot).
		Updates(map[string]interface{}{
			"cardano_tx_id":     txID,
			"blockchain_status": "anchored",
			"merkle_root":       merkleRoot,
		})

	// Create batch anchor record
	batchAnchor := models.BatchAnchor{
		UniversityID:     university.ID,
		MerkleRootHash:   merkleRoot,
		CardanoTxID:      txID,
		CertificateCount: len(certificates),
	}
	database.DB.Create(&batchAnchor)

	// Determine network for explorer URL
	network := os.Getenv("CARDANO_NETWORK")
	if network == "" {
		network = "preview"
	}

	explorerURL := fmt.Sprintf("https://%s.cardanoscan.io/transaction/%s", network, txID)

	// ========== EMAIL SENDING ==========

	emailsSent := 0
	var emailErrors []string

	// Check if email service is enabled and user wants emails
	emailEnabled := os.Getenv("SMTP_HOST") != ""

	if emailEnabled {
		log.Printf("📧 Starting email distribution to %d students...", len(certificates))

		emailService := services.NewEmailService()
		ipfsClient := ipfs.NewClient(os.Getenv("IPFS_API_URL"))

		// Reload certificates with updated data
		var updatedCerts []models.Certificate
		database.DB.Where("merkle_root_hash = ?", merkleRoot).Find(&updatedCerts)

		for _, cert := range updatedCerts {
			// Check if student has email
			if cert.StudentEmail == "" {
				log.Printf("⚠️  No email for student: %s", cert.StudentName)
				continue
			}

			// Check if IPFS hash exists
			if cert.IPFSPDFHash == "" {
				log.Printf("⚠️  No IPFS hash for certificate: %s", cert.ID)
				continue
			}

			// Download PDF from IPFS
			log.Printf("📥 Downloading PDF from IPFS: %s", cert.IPFSPDFHash)
			pdfBytes, err := ipfsClient.DownloadPDF(cert.IPFSPDFHash)
			if err != nil {
				errMsg := fmt.Sprintf("%s: Failed to download from IPFS - %v", cert.StudentName, err)
				emailErrors = append(emailErrors, errMsg)
				log.Printf("❌ %s", errMsg)
				continue
			}

			// Save temporarily for email attachment
			tempPDFPath := fmt.Sprintf("/tmp/cert_%s.pdf", cert.CertID)
			if err := os.WriteFile(tempPDFPath, pdfBytes, 0644); err != nil {
				errMsg := fmt.Sprintf("%s: Failed to save temp PDF - %v", cert.StudentName, err)
				emailErrors = append(emailErrors, errMsg)
				log.Printf("❌ %s", errMsg)
				continue
			}
			defer os.Remove(tempPDFPath) // Clean up after sending

			// Prepare email data
			frontendURL := os.Getenv("FRONTEND_URL")
			if frontendURL == "" {
				frontendURL = "http://localhost:5173"
			}

			verifyURL := fmt.Sprintf("%s/verify/%s", frontendURL, cert.CertID)

			emailData := services.CertificateEmailData{
				StudentName:    cert.StudentName,
				UniversityName: university.Name,
				CourseName:     cert.Degree,
				IssueDate:      cert.IssueDate.Format("January 2, 2006"),
				CertificateID:  cert.CertID,
				VerifyURL:      verifyURL,
				TransactionURL: explorerURL,
				TransactionID:  txID,
			}

			// Send email
			err = emailService.SendCertificateEmail(
				cert.StudentEmail,
				cert.StudentName,
				emailData,
				tempPDFPath, // Use temp file path
			)

			if err != nil {
				errMsg := fmt.Sprintf("%s (%s): %v", cert.StudentName, cert.StudentEmail, err)
				emailErrors = append(emailErrors, errMsg)
				log.Printf("❌ Failed to send email to %s: %v", cert.StudentEmail, err)
			} else {
				emailsSent++

				// Mark email as sent in database
				database.DB.Model(&cert).Update("email_sent", true)
				log.Printf("✅ Email sent to %s", cert.StudentEmail)
			}
		}

		log.Printf("✅ Email distribution complete: %d/%d sent", emailsSent, len(certificates))
	} else if !emailEnabled {
		log.Printf("ℹ️  Email service not configured (SMTP_HOST not set)")
	}

	// Calculate costs
	costPerCert := 0.08 / float64(len(certificates))

	// Prepare response message
	message := fmt.Sprintf("✅ %d certificates anchored to blockchain!", len(certificates))
	if emailsSent > 0 {
		message = fmt.Sprintf("✅ %d certificates anchored and %d emails sent to students!", len(certificates), emailsSent)
	} else if req.SendEmails && !emailEnabled {
		message = fmt.Sprintf("✅ %d certificates anchored (email not configured)", len(certificates))
	}

	response := gin.H{
		"success":       true,
		"certificates":  len(certificates),
		"merkle_root":   merkleRoot,
		"cardano_txid":  txID,
		"explorer_url":  explorerURL,
		"cost":          "0.17 ADA (~$0.08)",
		"cost_per_cert": fmt.Sprintf("$%.6f", costPerCert),
		"message":       message,
		"wallet":        university.CardanoPublicKey,
		"network":       network,
		"emails_sent":   emailsSent,
	}

	// Add email errors if any
	if len(emailErrors) > 0 {
		response["email_errors"] = emailErrors
		response["email_warning"] = fmt.Sprintf("%d emails failed to send", len(emailErrors))
	}

	c.JSON(http.StatusOK, response)

}
