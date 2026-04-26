package handlers

import (
	"log"
	"net/http"
	"time"

	"cert-system/server/internal/database"
	"cert-system/server/internal/middleware"
	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// SearchStudents searches for students with hidden profiles
// @Summary Search for students
// @Description Search for students based on requirements (shows hidden profiles)
// @Tags Company
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param filters body models.SearchFilters true "Search filters"
// @Success 200 {object} map[string]interface{}
// @Router /company/search [post]
func SearchStudents(c *gin.Context) {
	companyID, _, _, exists := middleware.GetCompanyFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var filters models.SearchFilters
	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if filters.Limit == 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	// Build query
	query := database.DB.Table("certificates c").
		Select(`
			c.id,
			c.cert_id,
			'***' as student_name,
			'***' as student_email,
			c.degree,
			c.gpa,
			c.skills,
			c.experience_years,
			c.issue_date,
			c.university_id,
			u.name as university_name,
			u.domain as university_domain,
			c.blockchain_status,
			c.cardano_tx_id
		`).
		Joins("LEFT JOIN universities u ON c.university_id = u.id").
		Where("c.is_searchable = ? AND c.blockchain_status = ?", true, "anchored")

	// Apply filters
	if filters.Degree != "" {
		query = query.Where("c.degree ILIKE ?", "%"+filters.Degree+"%")
	}

	if filters.MinGPA > 0 {
		query = query.Where("c.gpa >= ?", filters.MinGPA)
	}

	if filters.MaxGPA > 0 {
		query = query.Where("c.gpa <= ?", filters.MaxGPA)
	}

	if len(filters.Skills) > 0 {
		query = query.Where("c.skills && ?", pq.Array(filters.Skills))
	}

	if filters.MinExperience > 0 {
		query = query.Where("c.experience_years >= ?", filters.MinExperience)
	}

	if filters.UniversityID != "" {
		query = query.Where("c.university_id = ?", filters.UniversityID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get results
	var profiles []models.StudentProfileHidden
	if err := query.Limit(filters.Limit).Offset(filters.Offset).Scan(&profiles).Error; err != nil {
		log.Printf("Error searching students: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	// Check which profiles have pending/accepted requests
	profileIDs := make([]uuid.UUID, len(profiles))
	for i, p := range profiles {
		profileIDs[i] = p.ID
	}

	var requests []models.ProfileRequest
	database.DB.Where("company_id = ? AND certificate_id IN ?", companyID, profileIDs).Find(&requests)

	// Create map of request status
	requestStatus := make(map[uuid.UUID]string)
	for _, req := range requests {
		requestStatus[req.CertificateID] = req.Status
	}

	// Add request status to profiles
	type ProfileWithStatus struct {
		models.StudentProfileHidden
		RequestStatus string `json:"request_status"`
		CanRequest    bool   `json:"can_request"`
	}

	var result []ProfileWithStatus
	for _, profile := range profiles {
		status := requestStatus[profile.ID]
		canRequest := status == "" || status == "rejected" || status == "expired"

		result = append(result, ProfileWithStatus{
			StudentProfileHidden: profile,
			RequestStatus:        status,
			CanRequest:           canRequest,
		})
	}

	log.Printf("🔍 Company %s searched: found %d profiles", companyID, len(result))

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"profiles": result,
		"total":    total,
		"limit":    filters.Limit,
		"offset":   filters.Offset,
	})
}

// RequestProfileAccess requests access to student profile
// @Summary Request profile access
// @Description Request to view full student profile details
// @Tags Company
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body map[string]string true "Request data"
// @Success 200 {object} map[string]interface{}
// @Router /company/request-profile [post]
func RequestProfileAccess(c *gin.Context) {
	companyID, _, companyName, exists := middleware.GetCompanyFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		CertificateID string `json:"certificate_id" binding:"required"`
		Message       string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	certID, err := uuid.Parse(req.CertificateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate ID"})
		return
	}

	// Check if certificate exists and is searchable
	var cert models.Certificate
	if err := database.DB.Where("id = ? AND is_searchable = ?", certID, true).First(&cert).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found or not searchable"})
		return
	}

	// Check if request already exists
	var existing models.ProfileRequest
	err = database.DB.Where("company_id = ? AND certificate_id = ? AND status IN ?",
		companyID, certID, []string{"pending", "accepted"}).First(&existing).Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "Request already exists",
			"status": existing.Status,
		})
		return
	}

	// Create request
	profileReq := models.ProfileRequest{
		CompanyID:     uuid.MustParse(companyID),
		CertificateID: certID,
		StudentEmail:  cert.StudentEmail,
		Status:        "pending",
		Message:       req.Message,
		RequestedAt:   time.Now(),
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour), // 30 days
	}

	if err := database.DB.Create(&profileReq).Error; err != nil {
		log.Printf("Error creating profile request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	log.Printf("📬 Company %s requested access to profile %s", companyName, cert.StudentName)

	// TODO: Send email notification to student

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Profile access requested",
		"request_id": profileReq.ID,
		"expires_at": profileReq.ExpiresAt,
	})
}

// GetMyRequests gets all requests made by company
// @Summary Get my profile requests
// @Description Get all profile access requests made by this company
// @Tags Company
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /company/my-requests [get]
func GetMyRequests(c *gin.Context) {
	companyID, _, _, exists := middleware.GetCompanyFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var requests []models.ProfileRequest
	if err := database.DB.Where("company_id = ?", companyID).
		Order("requested_at DESC").
		Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"requests": requests,
		"total":    len(requests),
	})
}

// GetAcceptedProfiles gets full profile details for accepted requests
// @Summary Get accepted profiles
// @Description Get full details of students who accepted profile requests
// @Tags Company
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /company/accepted-profiles [get]
func GetAcceptedProfiles(c *gin.Context) {
	companyID, _, _, exists := middleware.GetCompanyFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get accepted requests
	var requests []models.ProfileRequest
	if err := database.DB.Where("company_id = ? AND status = ?", companyID, "accepted").
		Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
		return
	}

	// Get certificate IDs
	certIDs := make([]uuid.UUID, len(requests))
	for i, req := range requests {
		certIDs[i] = req.CertificateID
	}

	// Get full profile details
	var profiles []models.StudentProfileFull
	if err := database.DB.Table("certificates c").
		Select(`
			c.id, c.cert_id, c.student_name, c.student_email,
			c.degree, c.gpa, c.age, c.gender, c.nationality,
			c.phone, c.linkedin_url, c.skills, c.experience_years,
			c.issue_date, c.university_id,
			u.name as university_name, u.domain as university_domain,
			u.is_verified as university_verified,
			c.blockchain_status, c.cardano_tx_id, c.ipfs_pdf_hash
		`).
		Joins("LEFT JOIN universities u ON c.university_id = u.id").
		Where("c.id IN ?", certIDs).
		Scan(&profiles).Error; err != nil {
		log.Printf("Error fetching profiles: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profiles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"profiles": profiles,
		"total":    len(profiles),
	})
}
