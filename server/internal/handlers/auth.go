package handlers

import (
	"log"
	"net/http"
	"time"

	"cert-system/server/internal/database"
	"cert-system/server/internal/middleware"
	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// =============================================================================
// SHARED TYPES
// =============================================================================

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required,min=8"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// =============================================================================
// UNIFIED LOGIN
// =============================================================================

func UnifiedLoginHandler(db *gorm.DB) gin.HandlerFunc {
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

		log.Printf("🔐 Login attempt: %s", req.Email)

		// ── 1. Try university ─────────────────────────────────────────────
		var university models.University
		if err := db.Where("email = ?", req.Email).First(&university).Error; err == nil {
			if !university.CheckPassword(req.Password) {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid email or password"})
				return
			}

			token, err := middleware.GenerateToken(university.ID.String(), university.Email, university.Name, university.TokenVersion)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
				return
			}

			log.Printf("✅ University login: %s", university.Name)
			c.JSON(http.StatusOK, gin.H{
				"success":  true,
				"message":  "Login successful",
				"role":     "university",
				"token":    token,
				"redirect": "/dashboard/university",
				"user": gin.H{
					"id":          university.ID,
					"name":        university.Name,
					"email":       university.Email,
					"domain":      university.Domain,
					"is_verified": university.IsVerified,
				},
			})
			return
		}

		// ── 2. Try company ────────────────────────────────────────────────
		var company models.Company
		if err := db.Where("email = ?", req.Email).First(&company).Error; err == nil {
			if !company.CheckPassword(req.Password) {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid email or password"})
				return
			}

			token, err := middleware.GenerateCompanyToken(company.ID.String(), company.Email, company.Name, company.TokenVersion)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
				return
			}

			log.Printf("✅ Company login: %s", company.Name)
			c.JSON(http.StatusOK, gin.H{
				"success":  true,
				"message":  "Login successful",
				"role":     "company",
				"token":    token,
				"redirect": "/dashboard/company",
				"user": gin.H{
					"id":          company.ID,
					"name":        company.Name,
					"email":       company.Email,
					"industry":    company.Industry,
					"is_verified": company.IsVerified,
				},
			})
			return
		}

		// ── 3. Try student ────────────────────────────────────────────────
		var student models.Student
		if err := db.Where("email = ? AND deleted_at IS NULL", req.Email).First(&student).Error; err == nil {
			if !student.CheckPassword(req.Password) {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid email or password"})
				return
			}

			now := time.Now()
			student.LastLogin = &now
			if err := db.Save(&student).Error; err != nil {
				log.Printf("Warning: failed to update last_login for %s: %v", student.Email, err)
			}

			token, err := middleware.GenerateStudentToken(student.ID.String(), student.Email, student.Name, student.TokenVersion)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
				return
			}

			log.Printf("✅ Student login: %s", student.Name)
			c.JSON(http.StatusOK, gin.H{
				"success":  true,
				"message":  "Login successful",
				"role":     "student",
				"token":    token,
				"redirect": "/dashboard/student",
				"user": gin.H{
					"id":            student.ID,
					"name":          student.Name,
					"email":         student.Email,
					"is_searchable": student.IsSearchable,
				},
			})
			return
		}

		// ── 4. Not found in any table ─────────────────────────────────────
		log.Printf("❌ Login failed — email not found: %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid email or password"})
	}
}

// =============================================================================
// UNIVERSITY AUTH
// =============================================================================

func GetCurrentUserHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		universityID, email, name, exists := middleware.GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		var university models.University
		if err := db.Where("id = ?", universityID).First(&university).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
			return
		}

		var totalCerts, anchoredCerts, pendingCerts int64
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

func RefreshTokenHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		universityID, email, name, exists := middleware.GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}
		var university models.University
		if err := db.Where("id = ?", universityID).First(&university).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
			return
		}
		token, err := middleware.GenerateToken(universityID, email, name, university.TokenVersion)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "token": token, "message": "Token refreshed successfully"})
	}
}

func RefreshCompanyTokenHandler(c *gin.Context) {
	companyID, email, name, exists := middleware.GetCompanyFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	var company models.Company
	if err := database.DB.Where("id = ?", companyID).First(&company).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}
	token, err := middleware.GenerateCompanyToken(companyID, email, name, company.TokenVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "token": token, "message": "Token refreshed successfully"})
}

func RefreshStudentTokenHandler(c *gin.Context) {
	studentID, email, name, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	var student models.Student
	if err := database.DB.Where("id = ? AND deleted_at IS NULL", studentID).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}
	token, err := middleware.GenerateStudentToken(studentID, email, name, student.TokenVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "token": token, "message": "Token refreshed successfully"})
}

func ChangePasswordHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		universityID, _, _, exists := middleware.GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		var req ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var university models.University
		if err := db.Where("id = ?", universityID).First(&university).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "University not found"})
			return
		}

		if !university.CheckPassword(req.CurrentPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		if err := db.Model(&university).Update("password", string(hashed)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		log.Printf("✅ Password changed: %s", university.Email)
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password changed successfully"})
	}
}

// auth.go — replace the old LogoutHandler with this
func LogoutHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if id, _, _, ok := middleware.GetStudentFromContext(c); ok {
			db.Model(&models.Student{}).Where("id = ?", id).
				UpdateColumn("token_version", gorm.Expr("token_version + 1"))

		} else if id, _, _, ok := middleware.GetCompanyFromContext(c); ok {
			db.Model(&models.Company{}).Where("id = ?", id).
				UpdateColumn("token_version", gorm.Expr("token_version + 1"))

		} else if id, _, _, ok := middleware.GetUserFromContext(c); ok {
			db.Model(&models.University{}).Where("id = ?", id).
				UpdateColumn("token_version", gorm.Expr("token_version + 1"))
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Logged out successfully. Please remove your token.",
		})
	}
}

// =============================================================================
// COMPANY AUTH
// =============================================================================

type CompanyRegisterRequest struct {
	Name        string `json:"name"         binding:"required"`
	Email       string `json:"email"        binding:"required,email"`
	Password    string `json:"password"     binding:"required,min=8"`
	Industry    string `json:"industry"`
	CompanySize string `json:"company_size"`
	Location    string `json:"location"`
	Website     string `json:"website"`
	Description string `json:"description"`
}

func RegisterCompany(c *gin.Context) {
	var req CompanyRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Company
	if err := database.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	company := models.Company{
		Name:        req.Name,
		Email:       req.Email,
		Industry:    req.Industry,
		CompanySize: req.CompanySize,
		Location:    req.Location,
		Website:     req.Website,
		Description: req.Description,
	}

	if err := company.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := database.DB.Create(&company).Error; err != nil {
		log.Printf("Error creating company: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
		return
	}

	log.Printf("✅ Company registered: %s (%s)", company.Name, company.Email)
	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"message":  "Company registered successfully",
		"role":     "company",
		"redirect": "/login",
		"company": gin.H{
			"id":       company.ID,
			"name":     company.Name,
			"email":    company.Email,
			"industry": company.Industry,
		},
	})
}

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

// =============================================================================
// STUDENT AUTH
// =============================================================================

type StudentRegisterRequest struct {
	Email       string `json:"email"        binding:"required,email"`
	Password    string `json:"password"     binding:"required,min=8"`
	Name        string `json:"name"         binding:"required"`
	Phone       string `json:"phone"`
	Age         *int   `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"nationality"`
	LinkedInURL string `json:"linkedin_url"`
}

func RegisterStudent(c *gin.Context) {
	var req StudentRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Student
	if err := database.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	student := models.Student{
		Email:           req.Email,
		Name:            req.Name,
		Phone:           req.Phone,
		Age:             req.Age,
		Gender:          req.Gender,
		Nationality:     req.Nationality,
		LinkedInURL:     req.LinkedInURL,
		IsSearchable:    true,
		ShowName:        false,
		ShowEmail:       false,
		ShowPhone:       false,
		ShowAge:         false,
		ShowGender:      false,
		ShowNationality: false,
		ShowLinkedIn:    false,
	}

	if err := student.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := database.DB.Create(&student).Error; err != nil {
		log.Printf("Error creating student: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student account"})
		return
	}

	log.Printf("✅ Student registered: %s (%s)", student.Name, student.Email)
	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"message":  "Registration successful. You can now log in.",
		"role":     "student",
		"redirect": "/login",
		"student": gin.H{
			"id":    student.ID,
			"name":  student.Name,
			"email": student.Email,
		},
	})
}
