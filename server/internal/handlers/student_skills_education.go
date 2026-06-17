package handlers

import (
	"net/http"
	"time"

	"cert-system/server/internal/database"
	"cert-system/server/internal/middleware"
	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// =============================================================================
// SKILLS
// =============================================================================

// allowedProficiencyLevels is the canonical set of accepted skill levels.
// Stored as a package-level var so it is initialised once and shared across calls.
var allowedProficiencyLevels = map[string]bool{
	"beginner":     true,
	"intermediate": true,
	"advanced":     true,
	"expert":       true,
}

// AddSkill adds a skill entry to the authenticated student's profile.
// @Summary Add skill
// @Description Add a skill to student profile
// @Tags Student
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /student/skills [post]
func AddSkill(c *gin.Context) {
	studentID, _, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		SkillName         string `json:"skill_name"          binding:"required"`
		ProficiencyLevel  string `json:"proficiency_level"`
		YearsOfExperience int    `json:"years_of_experience"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ProficiencyLevel != "" && !allowedProficiencyLevels[req.ProficiencyLevel] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "proficiency_level must be one of: beginner, intermediate, advanced, expert",
		})
		return
	}

	// uuid.Parse instead of uuid.MustParse — bad input returns 400, not a panic
	parsedStudentID, err := uuid.Parse(studentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	skill := models.StudentSkill{
		StudentID:         parsedStudentID,
		SkillName:         req.SkillName,
		ProficiencyLevel:  req.ProficiencyLevel,
		YearsOfExperience: req.YearsOfExperience,
	}

	if err := database.DB.Create(&skill).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add skill"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Skill added successfully",
		"skill":   skill,
	})
}

// DeleteSkill removes a skill entry from the authenticated student's profile.
// The student_id guard in the WHERE clause prevents one student deleting another's skill.
// @Summary Delete skill
// @Description Delete a skill from student profile
// @Tags Student
// @Security BearerAuth
// @Param id path string true "Skill ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /student/skills/{id} [delete]
func DeleteSkill(c *gin.Context) {
	studentID, _, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	skillID := c.Param("id")
	result := database.DB.
		Where("id = ? AND student_id = ?", skillID, studentID).
		Delete(&models.StudentSkill{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete skill"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Skill not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Skill deleted successfully",
	})
}

// =============================================================================
// EDUCATION
// =============================================================================

// AddEducation adds an education record to the authenticated student's profile.
// start_date and end_date are accepted in YYYY-MM-DD format and stored as *time.Time.
// @Summary Add education
// @Description Add an education record to student profile
// @Tags Student
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /student/education [post]
func AddEducation(c *gin.Context) {
	studentID, _, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		UniversityID string  `json:"university_id"`
		Degree       string  `json:"degree"         binding:"required"`
		FieldOfStudy string  `json:"field_of_study"`
		GPA          float64 `json:"gpa"`
		StartDate    string  `json:"start_date"` // expected: YYYY-MM-DD
		EndDate      string  `json:"end_date"`   // expected: YYYY-MM-DD
		IsCurrent    bool    `json:"is_current"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.GPA < 0 || req.GPA > 4.0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GPA must be between 0.0 and 4.0"})
		return
	}

	const dateLayout = "2006-01-02"
	var startDate, endDate *time.Time

	if req.StartDate != "" {
		t, err := time.Parse(dateLayout, req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, expected YYYY-MM-DD"})
			return
		}
		startDate = &t
	}

	if req.EndDate != "" {
		t, err := time.Parse(dateLayout, req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, expected YYYY-MM-DD"})
			return
		}
		endDate = &t
	}

	parsedStudentID, err := uuid.Parse(studentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	education := models.StudentEducation{
		StudentID:    parsedStudentID,
		Degree:       req.Degree,
		FieldOfStudy: req.FieldOfStudy,
		GPA:          req.GPA,
		StartDate:    startDate,
		EndDate:      endDate,
		IsCurrent:    req.IsCurrent,
	}

	if req.UniversityID != "" {
		uniID, err := uuid.Parse(req.UniversityID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid university_id"})
			return
		}
		education.UniversityID = &uniID
	}

	if err := database.DB.Create(&education).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add education"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"message":   "Education added successfully",
		"education": education,
	})
}

// DeleteEducation removes an education record from the authenticated student's profile.
// The student_id guard in the WHERE clause prevents cross-student deletions.
// @Summary Delete education
// @Description Delete an education record from student profile
// @Tags Student
// @Security BearerAuth
// @Param id path string true "Education ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /student/education/{id} [delete]
func DeleteEducation(c *gin.Context) {
	studentID, _, _, exists := middleware.GetStudentFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	educationID := c.Param("id")
	result := database.DB.
		Where("id = ? AND student_id = ?", educationID, studentID).
		Delete(&models.StudentEducation{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete education"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Education record not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Education deleted successfully",
	})
}
