package handlers

import (
	"log"
	"net/http"

	"cert-system/server/internal/database"
	"cert-system/server/internal/middleware"
	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Success    bool           `json:"success"`
	Message    string         `json:"message"`
	Token      string         `json:"token"`
	University universityInfo `json:"university"`
}

type CompanyLoginResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Token   string      `json:"token"`
	Company companyInfo `json:"company"`
}

// UniversityInfo represents university data
type universityInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Domain         string `json:"domain"`
	CardanoAddress string `json:"cardano_address"`
	IsVerified     bool   `json:"is_verified"`
}

type companyInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Industry   string `json:"industry"`
	Location   string `json:"location"`
	IsVerified bool   `json:"is_verified"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// LoginHandler handles university login
// @Summary University login
// @Description Login with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func LoginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		log.Printf("🔐 University login attempt: %s", req.Email)

		// Find university by email
		var university models.University
		if err := db.Where("email = ?", req.Email).First(&university).Error; err != nil {
			log.Printf("❌ University not found: %s", req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid email or password",
			})
			return
		}

		// Verify password
		if !university.CheckPassword(req.Password) {
			log.Printf("❌ Invalid password for: %s", req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid email or password",
			})
			return
		}

		log.Printf("✅ University login successful: %s", university.Name)

		// Generate JWT token
		token, err := middleware.GenerateToken(
			university.ID.String(),
			university.Email,
			university.Name,
		)
		if err != nil {
			log.Printf("❌ Token generation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to generate token",
			})
			return
		}

		// Response
		c.JSON(http.StatusOK, LoginResponse{
			Success: true,
			Message: "Login successful",
			Token:   token,
			University: universityInfo{
				ID:             university.ID.String(),
				Name:           university.Name,
				Email:          university.Email,
				Domain:         university.Domain,
				CardanoAddress: university.CardanoPublicKey,
				IsVerified:     university.IsVerified,
			},
		})
	}
}

// GetCurrentUserHandler returns current logged-in user info
// @Summary Get current user
// @Description Get currently logged-in university details
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Router /auth/me [get]
func GetCurrentUserHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		universityID, email, name, exists := middleware.GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		// Get full university details from database
		var university models.University
		if err := db.Where("id = ?", universityID).First(&university).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
			return
		}

		// Get certificate statistics
		var totalCerts int64
		var anchoredCerts int64
		var pendingCerts int64

		db.Model(&models.Certificate{}).Where("university_id = ?", universityID).Count(&totalCerts)
		db.Model(&models.Certificate{}).Where("university_id = ? AND cardano_tx_id != ?", universityID, "").Count(&anchoredCerts)
		db.Model(&models.Certificate{}).Where("university_id = ? AND cardano_tx_id = ?", universityID, "").Count(&pendingCerts)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"university": gin.H{
				"id":              university.ID,
				"name":            name,
				"email":           email,
				"domain":          university.Domain,
				"cardano_address": university.CardanoPublicKey,
				"is_verified":     university.IsVerified,
			},
			"statistics": gin.H{
				"total_certificates":    totalCerts,
				"anchored_certificates": anchoredCerts,
				"pending_certificates":  pendingCerts,
			},
		})
	}
}

// RefreshTokenHandler refreshes the JWT token
// @Summary Refresh token
// @Description Get a new JWT token
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func RefreshTokenHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		universityID, email, name, exists := middleware.GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		// Generate new token
		token, err := middleware.GenerateToken(universityID, email, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"token":   token,
			"message": "Token refreshed successfully",
		})
	}
}

// ChangePasswordHandler changes user password
// @Summary Change password
// @Description Change university password
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ChangePasswordRequest true "Password change request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /auth/change-password [post]
func ChangePasswordHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		universityID, _, _, exists := middleware.GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		var req struct {
			CurrentPassword string `json:"current_password" binding:"required"`
			NewPassword     string `json:"new_password" binding:"required,min=8"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get university
		var university models.University
		if err := db.Where("id = ?", universityID).First(&university).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
			return
		}

		// Verify current password
		if !university.CheckPassword(req.CurrentPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}

		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Update password
		if err := db.Model(&university).Update("password", string(hashedPassword)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		log.Printf("✅ Password changed for: %s", university.Email)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Password changed successfully",
		})
	}
}

// LogoutHandler handles logout (client-side token removal)
// @Summary Logout
// @Description Logout user (client should remove token)
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Logged out successfully. Please remove your token.",
		})
	}
}

type CompanyRegisterRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Industry    string `json:"industry"`
	CompanySize string `json:"company_size"`
	Location    string `json:"location"`
	Website     string `json:"website"`
	Description string `json:"description"`
}

// RegisterCompany handles company registration
// @Summary Register a company
// @Description Register a new company account for job search
// @Tags Company
// @Accept json
// @Produce json
// @Param body body CompanyRegisterRequest true "Company registration data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /company/register [post]
func RegisterCompany(c *gin.Context) {
	var req CompanyRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	var existing models.Company
	if err := database.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Create company
	company := models.Company{
		Name:        req.Name,
		Email:       req.Email,
		Industry:    req.Industry,
		CompanySize: req.CompanySize,
		Location:    req.Location,
		Website:     req.Website,
		Description: req.Description,
	}

	// Hash password
	if err := company.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Save to database
	if err := database.DB.Create(&company).Error; err != nil {
		log.Printf("Error creating company: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
		return
	}

	log.Printf("✅ Company registered: %s (%s)", company.Name, company.Email)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Company registered successfully",
		"company": gin.H{
			"id":       company.ID,
			"name":     company.Name,
			"email":    company.Email,
			"industry": company.Industry,
		},
	})
}

// LoginCompany handles company login
// @Summary Company login
// @Description Login with company email and password
// @Tags Company
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} CompanyLoginResponse
// @Failure 401 {object} map[string]interface{}
// @Router /company/login [post]
func LoginCompany(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("🔐 Company login attempt: %s", req.Email)

	// Find company
	var company models.Company
	if err := database.DB.Where("email = ?", req.Email).First(&company).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check password
	if !company.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	log.Printf("✅ Company login successful: %s", company.Name)

	// Generate JWT token
	token, err := middleware.GenerateCompanyToken(
		company.ID.String(),
		company.Email,
		company.Name,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, CompanyLoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		Company: companyInfo{
			ID:         company.ID.String(),
			Name:       company.Name,
			Email:      company.Email,
			Industry:   company.Industry,
			Location:   company.Location,
			IsVerified: company.IsVerified,
		},
	})
}

// GetCurrentCompany returns current logged-in company
// @Summary Get current company
// @Description Get currently logged-in company details
// @Tags Company
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /company/me [get]
func GetCurrentCompany(c *gin.Context) {
	companyID, _, _, exists := middleware.GetCompanyFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var company models.Company
	if err := database.DB.Where("id = ?", companyID).First(&company).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	// Get request statistics
	var totalRequests, pendingRequests, acceptedRequests int64
	database.DB.Model(&models.ProfileRequest{}).Where("company_id = ?", companyID).Count(&totalRequests)
	database.DB.Model(&models.ProfileRequest{}).Where("company_id = ? AND status = ?", companyID, "pending").Count(&pendingRequests)
	database.DB.Model(&models.ProfileRequest{}).Where("company_id = ? AND status = ?", companyID, "accepted").Count(&acceptedRequests)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"company": gin.H{
			"id":          company.ID,
			"name":        company.Name,
			"email":       company.Email,
			"industry":    company.Industry,
			"location":    company.Location,
			"website":     company.Website,
			"is_verified": company.IsVerified,
		},
		"statistics": gin.H{
			"total_requests":    totalRequests,
			"pending_requests":  pendingRequests,
			"accepted_requests": acceptedRequests,
		},
	})
}
