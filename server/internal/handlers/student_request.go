package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"cert-system/server/internal/database"
	"cert-system/server/internal/middleware"
	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GET /api/student/profile-requests
func GetMyProfileRequests(c *gin.Context) {
	studentID, email, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	_ = studentID

	var requests []models.ProfileRequest
	if err := database.DB.
		Preload("Company").
		Where("student_email = ?", email).
		Order("requested_at DESC").
		Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
		return
	}

	// Nested company object — matches frontend: req.company?.name, req.company?.website
	type RequestWithCompany struct {
		ID          uuid.UUID  `json:"id"`
		Status      string     `json:"status"`
		Message     string     `json:"message"`
		RequestedAt time.Time  `json:"requested_at"`
		ExpiresAt   time.Time  `json:"expires_at"`
		RespondedAt *time.Time `json:"responded_at,omitempty"`
		IsExpired   bool       `json:"is_expired"`
		Company     gin.H      `json:"company"`
	}

	result := make([]RequestWithCompany, 0, len(requests))
	for _, req := range requests {
		result = append(result, RequestWithCompany{
			ID:          req.ID,
			Status:      req.Status,
			Message:     req.Message,
			RequestedAt: req.RequestedAt,
			ExpiresAt:   req.ExpiresAt,
			RespondedAt: req.RespondedAt,
			IsExpired:   time.Now().After(req.ExpiresAt) && req.Status == "pending",
			Company: gin.H{
				"id":       req.Company.ID,
				"name":     req.Company.Name,
				"industry": req.Company.Industry,
				"location": req.Company.Location,
				"website":  req.Company.Website,
			},
		})
	}

	var pending, accepted, rejected int64
	database.DB.Model(&models.ProfileRequest{}).Where("student_email = ? AND status = ?", email, "pending").Count(&pending)
	database.DB.Model(&models.ProfileRequest{}).Where("student_email = ? AND status = ?", email, "accepted").Count(&accepted)
	database.DB.Model(&models.ProfileRequest{}).Where("student_email = ? AND status = ?", email, "rejected").Count(&rejected)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"requests": result,
		"total":    len(result),
		"counts":   gin.H{"pending": pending, "accepted": accepted, "rejected": rejected},
	})
}

// POST /api/student/profile-requests/:id/respond
// Frontend calls: api.respondToRequest(id, { action: "accept" | "reject" })
// ID comes from the URL param :id — NOT from the request body
func RespondToRequest(c *gin.Context) {
	_, email, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Read request ID from URL param :id
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	// Accept both { action: "accept"/"reject" } and { status: "accepted"/"rejected" }
	var body struct {
		Action string `json:"action"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalise to a single action string
	action := body.Action
	if action == "" {
		action = body.Status
	}
	if action != "accept" && action != "reject" &&
		action != "accepted" && action != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action must be 'accept' or 'reject'"})
		return
	}

	var profileReq models.ProfileRequest
	if err := database.DB.
		Where("id = ? AND student_email = ?", requestID, email).
		First(&profileReq).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	if profileReq.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "Request already responded to",
			"status": profileReq.Status,
		})
		return
	}
	if time.Now().After(profileReq.ExpiresAt) {
		database.DB.Model(&profileReq).Update("status", "expired")
		c.JSON(http.StatusGone, gin.H{"error": "Request has expired"})
		return
	}

	// Normalise to stored form: "accepted" / "rejected"
	status := "rejected"
	if action == "accept" || action == "accepted" {
		status = "accepted"
	}

	now := time.Now()
	if err := database.DB.Model(&profileReq).Updates(map[string]interface{}{
		"status":       status,
		"responded_at": now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update request"})
		return
	}

	log.Printf("✅ Student %s %s profile request %s", email, status, requestID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Request %s successfully", status),
		"status":  status,
	})
}

// GET /api/student/profile-settings
func GetMyProfileSettings(c *gin.Context) {
	_, email, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var student models.Student
	if err := database.DB.Where("email = ? AND deleted_at IS NULL", email).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	// Only count certificates the student has actually uploaded (pdf_path != '')
	var totalCerts int64
	database.DB.Model(&models.Certificate{}).
		Where("student_email = ? AND pdf_path != ''", email).
		Count(&totalCerts)

	var pendingRequests int64
	database.DB.Model(&models.ProfileRequest{}).
		Where("student_email = ? AND status = ?", email, "pending").
		Count(&pendingRequests)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"profile": gin.H{
			"email":              student.Email,
			"name":               student.Name,
			"is_searchable":      student.IsSearchable,
			"show_name":          student.ShowName,
			"show_email":         student.ShowEmail,
			"show_phone":         student.ShowPhone,
			"show_age":           student.ShowAge,
			"show_gender":        student.ShowGender,
			"show_nationality":   student.ShowNationality,
			"show_linkedin":      student.ShowLinkedIn,
			"total_certificates": totalCerts,
		},
		"requests": gin.H{"pending": pendingRequests},
	})
}

// POST /api/student/privacy-settings
func UpdatePrivacySettings(c *gin.Context) {
	_, email, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		IsSearchable    *bool `json:"is_searchable"`
		ShowName        *bool `json:"show_name"`
		ShowEmail       *bool `json:"show_email"`
		ShowPhone       *bool `json:"show_phone"`
		ShowAge         *bool `json:"show_age"`
		ShowGender      *bool `json:"show_gender"`
		ShowNationality *bool `json:"show_nationality"`
		ShowLinkedIn    *bool `json:"show_linkedin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.IsSearchable != nil {
		updates["is_searchable"] = *req.IsSearchable
	}
	if req.ShowName != nil {
		updates["show_name"] = *req.ShowName
	}
	if req.ShowEmail != nil {
		updates["show_email"] = *req.ShowEmail
	}
	if req.ShowPhone != nil {
		updates["show_phone"] = *req.ShowPhone
	}
	if req.ShowAge != nil {
		updates["show_age"] = *req.ShowAge
	}
	if req.ShowGender != nil {
		updates["show_gender"] = *req.ShowGender
	}
	if req.ShowNationality != nil {
		updates["show_nationality"] = *req.ShowNationality
	}
	if req.ShowLinkedIn != nil {
		updates["show_linkedin"] = *req.ShowLinkedIn
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	if err := database.DB.Model(&models.Student{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update privacy settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Privacy settings updated"})
}
