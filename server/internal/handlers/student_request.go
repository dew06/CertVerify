package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"cert-system/server/internal/database"
	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetMyProfileRequests gets all requests for student's profile
// @Summary Get profile requests
// @Description Get all company requests to view your profile (student endpoint)
// @Tags Student
// @Accept json
// @Produce json
// @Param email query string true "Student email"
// @Success 200 {object} map[string]interface{}
// @Router /student/profile-requests [get]
func GetMyProfileRequests(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email required"})
		return
	}

	// Get all requests for this student
	var requests []models.ProfileRequest
	if err := database.DB.
		Preload("Company").
		Where("student_email = ?", email).
		Order("requested_at DESC").
		Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
		return
	}

	// Format response with company details
	type RequestWithCompany struct {
		ID              uuid.UUID `json:"id"`
		CompanyName     string    `json:"company_name"`
		CompanyIndustry string    `json:"company_industry"`
		CompanyLocation string    `json:"company_location"`
		CompanyWebsite  string    `json:"company_website"`
		Message         string    `json:"message"`
		Status          string    `json:"status"`
		RequestedAt     time.Time `json:"requested_at"`
		ExpiresAt       time.Time `json:"expires_at"`
		IsExpired       bool      `json:"is_expired"`
	}

	var result []RequestWithCompany
	for _, req := range requests {
		result = append(result, RequestWithCompany{
			ID:              req.ID,
			CompanyName:     req.Company.Name,
			CompanyIndustry: req.Company.Industry,
			CompanyLocation: req.Company.Location,
			CompanyWebsite:  req.Company.Website,
			Message:         req.Message,
			Status:          req.Status,
			RequestedAt:     req.RequestedAt,
			ExpiresAt:       req.ExpiresAt,
			IsExpired:       time.Now().After(req.ExpiresAt) && req.Status == "pending",
		})
	}

	// Count by status
	var pending, accepted, rejected int64
	database.DB.Model(&models.ProfileRequest{}).Where("student_email = ? AND status = ?", email, "pending").Count(&pending)
	database.DB.Model(&models.ProfileRequest{}).Where("student_email = ? AND status = ?", email, "accepted").Count(&accepted)
	database.DB.Model(&models.ProfileRequest{}).Where("student_email = ? AND status = ?", email, "rejected").Count(&rejected)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"requests": result,
		"total":    len(result),
		"counts": gin.H{
			"pending":  pending,
			"accepted": accepted,
			"rejected": rejected,
		},
	})
}

// RespondToRequest allows student to accept/reject profile request
// @Summary Respond to profile request
// @Description Accept or reject a company's request to view your profile
// @Tags Student
// @Accept json
// @Produce json
// @Param body body map[string]string true "Response data"
// @Success 200 {object} map[string]interface{}
// @Router /student/respond-request [post]
func RespondToRequest(c *gin.Context) {
	var req struct {
		RequestID string `json:"request_id" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		Action    string `json:"action" binding:"required"` // accept or reject
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate action
	if req.Action != "accept" && req.Action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action must be 'accept' or 'reject'"})
		return
	}

	requestID, err := uuid.Parse(req.RequestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	// Get request
	var profileReq models.ProfileRequest
	if err := database.DB.Where("id = ? AND student_email = ?", requestID, req.Email).
		First(&profileReq).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	// Check if already responded
	if profileReq.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "Request already responded to",
			"status": profileReq.Status,
		})
		return
	}

	// Check if expired
	if time.Now().After(profileReq.ExpiresAt) {
		database.DB.Model(&profileReq).Updates(map[string]interface{}{
			"status": "expired",
		})
		c.JSON(http.StatusGone, gin.H{"error": "Request has expired"})
		return
	}

	// Update request
	status := "rejected"
	if req.Action == "accept" {
		status = "accepted"
	}

	now := time.Now()
	if err := database.DB.Model(&profileReq).Updates(map[string]interface{}{
		"status":       status,
		"responded_at": now,
	}).Error; err != nil {
		log.Printf("Error updating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update request"})
		return
	}

	log.Printf("✅ Student %s %s profile request from company", req.Email, status)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Request %s successfully", status),
		"status":  status,
	})
}

// UpdateProfileVisibility allows student to hide/show profile in searches
// @Summary Update profile visibility
// @Description Toggle whether your profile appears in company searches
// @Tags Student
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Visibility data"
// @Success 200 {object} map[string]interface{}
// @Router /student/profile-visibility [post]
func UpdateProfileVisibility(c *gin.Context) {
	var req struct {
		Email        string `json:"email" binding:"required,email"`
		IsSearchable bool   `json:"is_searchable"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update all certificates for this student
	result := database.DB.Model(&models.Certificate{}).
		Where("student_email = ?", req.Email).
		Update("is_searchable", req.IsSearchable)

	if result.Error != nil {
		log.Printf("Error updating visibility: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update visibility"})
		return
	}

	log.Printf("✅ Student %s set profile visibility to: %v", req.Email, req.IsSearchable)

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Profile visibility updated",
		"is_searchable": req.IsSearchable,
		"updated_count": result.RowsAffected,
	})
}

// GetMyProfileSettings gets student's current profile settings
// @Summary Get profile settings
// @Description Get your current profile visibility and information
// @Tags Student
// @Accept json
// @Produce json
// @Param email query string true "Student email"
// @Success 200 {object} map[string]interface{}
// @Router /student/profile-settings [get]
func GetMyProfileSettings(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email required"})
		return
	}

	// Get one certificate to check settings (all should have same is_searchable)
	var cert models.Certificate
	if err := database.DB.Where("student_email = ?", email).First(&cert).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No certificates found"})
		return
	}

	// Count certificates
	var totalCerts int64
	database.DB.Model(&models.Certificate{}).Where("student_email = ?", email).Count(&totalCerts)

	// Count pending requests
	var pendingRequests int64
	database.DB.Model(&models.ProfileRequest{}).
		Where("student_email = ? AND status = ?", email, "pending").
		Count(&pendingRequests)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"profile": gin.H{
			"email":              cert.StudentEmail,
			"name":               cert.StudentName,
			"is_searchable":      cert.IsSearchable,
			"total_certificates": totalCerts,
		},
		"requests": gin.H{
			"pending": pendingRequests,
		},
	})
}
