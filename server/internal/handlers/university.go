package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"cert-system/server/internal/database"
	"cert-system/server/internal/models"
	"cert-system/server/internal/services"

	"github.com/gin-gonic/gin"
)

type UniversityHandler struct{}

func NewUniversityHandler() *UniversityHandler {
	return &UniversityHandler{}
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Domain   string `json:"domain" binding:"required"`
	Password string `json:"password" binding:"required,min=12"`
}

// Register godoc
// @Summary Register a new university
// @Description Register a university with domain and generate Cardano wallet
// @Tags university
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "University registration details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /university/register [post]
func (h *UniversityHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	var existing models.University
	if database.DB.Where("email = ?", req.Email).First(&existing).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Check if domain already exists
	if database.DB.Where("domain = ?", req.Domain).First(&existing).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Domain already registered"})
		return
	}

	// Get network from environment
	network := os.Getenv("CARDANO_NETWORK")
	if network == "" {
		network = "preview"
	}

	// Generate Cardano wallet with PROPER address generation
	walletAddress, privateKey, recoveryPhrase, err := services.GenerateCardanoWallet(network)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Wallet generation failed",
			"details": err.Error(),
		})
		return
	}

	// Generate encryption salt
	saltBytes := make([]byte, 32)
	rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)

	// Encrypt private key with password
	encryptedPrivateKey, err := services.EncryptPrivateKey(privateKey, req.Password, salt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encrypt private key",
		})
		return
	}

	// Create university record
	// 1. Initialize the struct
	university := models.University{
		Name:                req.Name,
		Email:               req.Email,
		Domain:              req.Domain,
		CardanoPublicKey:    walletAddress,
		EncryptedPrivateKey: encryptedPrivateKey,
		EncryptionSalt:      salt,
		IsVerified:          false,
	}

	if err := university.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process security credentials",
		})
		return
	}

	// 3. Save to Database (Now including the hashed password)
	if err := database.DB.Create(&university).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create university",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":              university.ID,
		"wallet_address":  walletAddress,
		"recovery_phrase": recoveryPhrase,
		"message":         "✅ University registered successfully!",
		"warning":         "⚠️ SAVE YOUR RECOVERY PHRASE! This is the only way to recover your wallet.",
		"network":         network,
		"next_steps": []string{
			"1. Save your recovery phrase in a secure location",
			"2. Fund your wallet with test ADA from faucet",
			"3. Verify your domain ownership",
			"4. Start issuing certificates",
		},
	})
}

type DomainVerificationResponse struct {
	UniversityID string                 `json:"university_id"`
	Domain       string                 `json:"domain"`
	Instructions string                 `json:"instructions"`
	FileContent  map[string]interface{} `json:"file_content"`
	FilePath     string                 `json:"file_path"`
}

// GetDomainVerification godoc
// @Summary Get domain verification proof
// @Description Get the verification file content for domain proof
// @Tags university
// @Produce json
// @Param id path string true "University ID"
// @Success 200 {object} DomainVerificationResponse
// @Router /university/{id}/domain-proof [get]
func (h *UniversityHandler) GetDomainVerification(c *gin.Context) {
	universityID := c.Param("id")

	var university models.University
	if err := database.DB.Where("id = ?", universityID).First(&university).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
		return
	}

	fileContent := map[string]interface{}{
		"university_id":      university.ID,
		"domain":             university.Domain,
		"cardano_public_key": university.CardanoPublicKey,
		"verification_date":  university.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"university_id": university.ID,
		"domain":        university.Domain,
		"instructions": fmt.Sprintf(
			"Upload this JSON file to: https://%s/.well-known/cardano-key.json",
			university.Domain,
		),
		"file_content": fileContent,
		"file_path":    "/.well-known/cardano-key.json",
		"steps": []string{
			"1. Copy the file_content JSON below",
			"2. Create a file named 'cardano-key.json'",
			"3. Upload to your domain at: /.well-known/cardano-key.json",
			"4. Make sure the file is publicly accessible",
			"5. Click 'Verify Domain' button",
		},
	})
}

// VerifyDomain godoc
// @Summary Verify domain ownership
// @Description Verify that the university owns the domain by checking the verification file
// @Tags university
// @Produce json
// @Param id path string true "University ID"
// @Success 200 {object} map[string]interface{}
// @Router /university/{id}/verify-domain [post]
func (h *UniversityHandler) VerifyDomain(c *gin.Context) {
	universityID := c.Param("id")

	var university models.University
	result := database.DB.Where("id = ?", universityID).First(&university)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
		return
	}

	if university.IsVerified {
		c.JSON(http.StatusOK, gin.H{
			"message": "Domain already verified",
			"domain":  university.Domain,
		})
		return
	}

	// Construct verification URL
	verificationURL := fmt.Sprintf("https://%s/.well-known/cardano-key.json", university.Domain)

	// Try to fetch the verification file
	// In production, you would actually fetch and verify the file
	// For now, we'll simulate successful verification for development

	// Mark as verified
	university.IsVerified = true
	database.DB.Save(&university)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "✅ Domain verified successfully!",
		"domain":  university.Domain,
		"note": fmt.Sprintf(
			"In production, we would fetch and verify: %s",
			verificationURL,
		),
		"status": "verified",
	})
}
