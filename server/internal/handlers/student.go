package handlers

import (
	"log"
	"net/http"

	"cert-system/server/internal/database"
	"cert-system/server/internal/middleware"
	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
)

// GetCurrentStudent returns full profile of the authenticated student.
// @Summary Get current student
// @Description Get currently logged-in student profile, skills, education, and statistics
// @Tags Student
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /student/me [get]
func GetCurrentStudent(c *gin.Context) {
	studentID, _, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var student models.Student
	if err := database.DB.Where("id = ?", studentID).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	var skills []models.StudentSkill
	database.DB.Where("student_id = ?", studentID).Find(&skills)

	var education []models.StudentEducation
	database.DB.Preload("University").Where("student_id = ?", studentID).Find(&education)

	// Fetch full certificate records (not just count) so the frontend
	// can display them in the "My Certificates" list
	var certificates []models.Certificate
	database.DB.Where("student_email = ? AND pdf_path != ''", student.Email).
		Order("issue_date DESC").
		Find(&certificates)

	var pendingRequests int64
	database.DB.Model(&models.ProfileRequest{}).
		Where("student_email = ? AND status = ?", student.Email, "pending").Count(&pendingRequests)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"student": gin.H{
			"id":            student.ID,
			"name":          student.Name,
			"email":         student.Email,
			"phone":         student.Phone,
			"age":           student.Age,
			"gender":        student.Gender,
			"nationality":   student.Nationality,
			"linkedin_url":  student.LinkedInURL,
			"is_searchable": student.IsSearchable,
			"privacy_settings": gin.H{
				"show_name":        student.ShowName,
				"show_email":       student.ShowEmail,
				"show_phone":       student.ShowPhone,
				"show_age":         student.ShowAge,
				"show_gender":      student.ShowGender,
				"show_nationality": student.ShowNationality,
				"show_linkedin":    student.ShowLinkedIn,
			},
		},
		"skills":       skills,
		"education":    education,
		"certificates": certificates, // ← full records, not just a count
		"statistics": gin.H{
			"certificates":     len(certificates), // kept for backwards compat
			"pending_requests": pendingRequests,
		},
	})
}

// UpdateStudentProfile updates editable profile fields for the authenticated student.
// Pointer-typed fields mean omitted JSON keys are treated as "no change" rather
// than overwriting existing data with zero values.
// @Summary Update student profile
// @Description Update student profile information
// @Tags Student
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /student/profile [put]
func UpdateStudentProfile(c *gin.Context) {
	studentID, _, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// All fields are optional pointers. A missing field stays nil and is excluded
	// from the UPDATE, preventing zero-value overwrites on partial payloads.
	var req struct {
		Name        *string `json:"name"`
		Phone       *string `json:"phone"`
		Age         *int    `json:"age"`
		Gender      *string `json:"gender"`
		Nationality *string `json:"nationality"`
		LinkedInURL *string `json:"linkedin_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var student models.Student
	if err := database.DB.Where("id = ?", studentID).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Age != nil {
		updates["age"] = *req.Age
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Nationality != nil {
		updates["nationality"] = *req.Nationality
	}
	if req.LinkedInURL != nil {
		updates["linkedin_url"] = *req.LinkedInURL
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields provided to update"})
		return
	}

	if err := database.DB.Model(&student).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	log.Printf("✅ Profile updated: %s", student.Email)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
	})
}
