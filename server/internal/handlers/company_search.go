package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"cert-system/server/internal/database"
	"cert-system/server/internal/middleware"
	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// =============================================================================
// SEARCH STUDENTS
// =============================================================================

// studentSearchResult is the privacy-filtered shape returned by SearchStudents.
// Hidden fields are omitempty — absent from JSON rather than "***".
type studentSearchResult struct {
	ID            uuid.UUID                 `json:"id"`
	Name          string                    `json:"name,omitempty"`
	Email         string                    `json:"email,omitempty"`
	Phone         string                    `json:"phone,omitempty"`
	Age           *int                      `json:"age,omitempty"`
	Gender        string                    `json:"gender,omitempty"`
	Nationality   string                    `json:"nationality,omitempty"`
	LinkedInURL   string                    `json:"linkedin_url,omitempty"`
	CreatedAt     time.Time                 `json:"created_at"`
	Skills        []models.StudentSkill     `json:"skills"`
	Education     []models.StudentEducation `json:"education"`
	CertCount     int64                     `json:"certificates_count"`
	RequestStatus string                    `json:"request_status"`
	CanRequest    bool                      `json:"can_request"`
}

// rawStudentRow is the intermediate DB scan target — keeps the real email
// private (needed for cert count + request lookup) while building the response.
type rawStudentRow struct {
	ID          uuid.UUID `gorm:"column:id"`
	RealEmail   string    `gorm:"column:real_email"` // never included in JSON output
	Name        string    `gorm:"column:name"`
	Email       string    `gorm:"column:email"` // visible_email (masked)
	Phone       string    `gorm:"column:phone"`
	Age         *int      `gorm:"column:age"`
	Gender      string    `gorm:"column:gender"`
	Nationality string    `gorm:"column:nationality"`
	LinkedInURL string    `gorm:"column:linkedin_url"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

// SearchStudents searches for students with privacy-controlled profiles.
func SearchStudents(c *gin.Context) {
	companyID, _, _, exists := middleware.GetCompanyFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid company identity"})
		return
	}

	// Accept both POST (JSON body) and GET (query params)
	var filters models.SearchFilters
	if err := c.ShouldBindJSON(&filters); err != nil {
		if err := c.ShouldBindQuery(&filters); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Clamp pagination
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	// Base query — only searchable, non-deleted students
	base := database.DB.Table("students s").
		Where("s.is_searchable = TRUE AND s.deleted_at IS NULL")

	// Whether any cert-related filter was provided
	certFilterActive := filters.MinGPA > 0 || filters.MaxGPA > 0 ||
		filters.Degree != "" || len(filters.Skills) > 0 ||
		filters.MinExperience > 0 || filters.UniversityID != ""

	if certFilterActive {
		// Join on pdf_path != '' — the student has uploaded/claimed the cert.
		// Do NOT filter by blockchain_status here; anchoring is asynchronous.
		base = base.Joins(`
			JOIN certificates c ON c.student_email = s.email
			AND c.pdf_path != ''
		`)

		if filters.Degree != "" {
			base = base.Where("c.degree ILIKE ?", "%"+filters.Degree+"%")
		}
		if filters.MinGPA > 0 {
			base = base.Where("c.gpa >= ?", filters.MinGPA)
		}
		if filters.MaxGPA > 0 {
			base = base.Where("c.gpa <= ?", filters.MaxGPA)
		}
		if filters.MinExperience > 0 {
			base = base.Where("c.experience_years >= ?", filters.MinExperience)
		}
		if len(filters.Skills) > 0 {
			base = base.Where("c.skills && ?", pq.Array(filters.Skills))
		}
		if filters.UniversityID != "" {
			base = base.Where("c.university_id = ?", filters.UniversityID)
		}
	}

	// Count using a subquery to avoid Distinct+Count issues in GORM.
	var total int64
	countSub := base.Select("s.id")
	if certFilterActive {
		countSub = countSub.Distinct("s.id")
	}
	if err := database.DB.Table("(?) AS count_sub", countSub).Count(&total).Error; err != nil {
		log.Printf("Error counting students: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	// Apply Distinct only on the fetch query so one student doesn't appear twice
	// when multiple certificates match.
	if certFilterActive {
		base = base.Distinct("s.*")
	}

	// Fetch the page — real_email is carried privately to support cert/request lookup
	var rows []rawStudentRow
	err = base.Select(`
		s.id,
		s.email                                           AS real_email,
		CASE WHEN s.show_name        THEN s.name         ELSE NULL END AS name,
		CASE WHEN s.show_email       THEN s.email        ELSE NULL END AS email,
		CASE WHEN s.show_phone       THEN s.phone        ELSE NULL END AS phone,
		CASE WHEN s.show_age         THEN s.age          ELSE NULL END AS age,
		CASE WHEN s.show_gender      THEN s.gender       ELSE NULL END AS gender,
		CASE WHEN s.show_nationality THEN s.nationality  ELSE NULL END AS nationality,
		CASE WHEN s.show_linkedin    THEN s.linkedin_url ELSE NULL END AS linkedin_url,
		s.created_at
	`).
		Limit(filters.Limit).
		Offset(filters.Offset).
		Scan(&rows).Error

	if err != nil {
		log.Printf("Error scanning students: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	if len(rows) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"students": []studentSearchResult{},
			"total":    total,
			"limit":    filters.Limit,
			"offset":   filters.Offset,
		})
		return
	}

	// Collect IDs and emails once — used for all batch lookups below
	studentIDs := make([]uuid.UUID, len(rows))
	emails := make([]string, len(rows))
	for i, r := range rows {
		studentIDs[i] = r.ID
		emails[i] = r.RealEmail
	}

	// --- Batch load skills (1 query, not N) ---
	var allSkills []models.StudentSkill
	database.DB.Where("student_id IN ?", studentIDs).Find(&allSkills)

	skillMap := make(map[uuid.UUID][]models.StudentSkill)
	for _, sk := range allSkills {
		skillMap[sk.StudentID] = append(skillMap[sk.StudentID], sk)
	}

	// --- Batch load education (1 query, not N) ---
	var allEducation []models.StudentEducation
	database.DB.Preload("University").Where("student_id IN ?", studentIDs).Find(&allEducation)

	eduMap := make(map[uuid.UUID][]models.StudentEducation)
	for _, ed := range allEducation {
		eduMap[ed.StudentID] = append(eduMap[ed.StudentID], ed)
	}

	// --- Batch load certificate counts (1 query, not N) ---
	type certCount struct {
		StudentEmail string
		Count        int64
	}
	var certCounts []certCount
	database.DB.Model(&models.Certificate{}).
		Select("student_email, COUNT(*) as count").
		Where("student_email IN ? AND pdf_path != ''", emails).
		Group("student_email").
		Scan(&certCounts)

	certMap := make(map[string]int64)
	for _, cc := range certCounts {
		certMap[cc.StudentEmail] = cc.Count
	}

	// --- Batch load request statuses (1 query, not N) ---
	type requestRow struct {
		StudentEmail string
		Status       string
	}
	var requestRows []requestRow
	database.DB.Raw(`
		SELECT DISTINCT ON (student_email)
			student_email, status
		FROM profile_requests
		WHERE company_id = ? AND student_email = ANY(?)
		ORDER BY student_email, requested_at DESC
	`, companyUUID, pq.Array(emails)).Scan(&requestRows)

	requestMap := make(map[string]string)
	for _, rr := range requestRows {
		requestMap[rr.StudentEmail] = rr.Status
	}

	// --- Assemble results ---
	results := make([]studentSearchResult, 0, len(rows))
	for _, row := range rows {
		status := requestMap[row.RealEmail]
		canRequest := status == "" || status == "rejected" || status == "expired"

		results = append(results, studentSearchResult{
			ID:            row.ID,
			Name:          row.Name,
			Email:         row.Email,
			Phone:         row.Phone,
			Age:           row.Age,
			Gender:        row.Gender,
			Nationality:   row.Nationality,
			LinkedInURL:   row.LinkedInURL,
			CreatedAt:     row.CreatedAt,
			Skills:        skillMap[row.ID],
			Education:     eduMap[row.ID],
			CertCount:     certMap[row.RealEmail],
			RequestStatus: status,
			CanRequest:    canRequest,
		})
	}

	log.Printf("🔍 Company %s searched: %d/%d results", companyID, len(results), total)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"students": results,
		"total":    total,
		"limit":    filters.Limit,
		"offset":   filters.Offset,
	})
}

// =============================================================================
// REQUEST PROFILE ACCESS
// =============================================================================

// RequestProfileAccess creates a request to view a student's full profile.
func RequestProfileAccess(c *gin.Context) {
	companyID, _, companyName, exists := middleware.GetCompanyFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Safe parse — MustParse would panic on a malformed JWT claim
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid company identity"})
		return
	}

	var req struct {
		StudentID string `json:"student_id" binding:"required"`
		Message   string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	// Student must be active, searchable, and not soft-deleted
	var student models.Student
	if err := database.DB.
		Where("id = ? AND is_searchable = TRUE AND 	deleted_at IS NULL", studentID).
		First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Student not found or not searchable"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to look up student"})
		}
		return
	}

	// Block duplicate pending / accepted requests
	var existing models.ProfileRequest
	err = database.DB.
		Where("company_id = ? AND student_email = ? AND status IN ?",
			companyUUID, student.Email, []string{"pending", "accepted"}).
		First(&existing).Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "A request for this student is already active",
			"status": existing.Status,
		})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking existing request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing requests"})
		return
	}

	profileReq := models.ProfileRequest{
		CompanyID:    companyUUID,
		StudentID:    &studentID,
		StudentEmail: student.Email,
		Status:       "pending",
		Message:      req.Message,
		RequestedAt:  time.Now(),
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
	}

	if err := database.DB.Create(&profileReq).Error; err != nil {
		log.Printf("Error creating profile request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	log.Printf("📬 Company %s requested access to student %s", companyName, student.Name)

	// TODO: Send email notification to student

	// 201 Created — a new resource was created
	c.JSON(http.StatusCreated, gin.H{
		"success":    true,
		"message":    "Profile access requested",
		"request_id": profileReq.ID,
		"expires_at": profileReq.ExpiresAt,
	})
}

// =============================================================================
// GET MY REQUESTS
// =============================================================================

// GetMyRequests returns all profile-access requests made by the authenticated company.
func GetMyRequests(c *gin.Context) {
	companyID, _, _, exists := middleware.GetCompanyFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid company identity"})
		return
	}

	var requests []models.ProfileRequest
	if err := database.DB.
		Where("company_id = ?", companyUUID).
		Order("requested_at DESC").
		Find(&requests).Error; err != nil {
		log.Printf("Error fetching requests for company %s: %v", companyID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
		return
	}

	// Build privacy-safe response — student email and identity are hidden
	// until they accept. Before acceptance the company only sees the request
	// metadata (status, timestamp, expiry).
	type safeRequest struct {
		ID          interface{} `json:"id"`
		Status      string      `json:"status"`
		Message     string      `json:"message"`
		RequestedAt time.Time   `json:"requested_at"`
		ExpiresAt   time.Time   `json:"expires_at"`
		RespondedAt *time.Time  `json:"responded_at,omitempty"`
		// Only populated after acceptance
		StudentEmail string      `json:"student_email,omitempty"`
		StudentID    interface{} `json:"student_id,omitempty"`
	}

	safe := make([]safeRequest, 0, len(requests))
	for _, r := range requests {
		sr := safeRequest{
			ID:          r.ID,
			Status:      r.Status,
			Message:     r.Message,
			RequestedAt: r.RequestedAt,
			ExpiresAt:   r.ExpiresAt,
			RespondedAt: r.RespondedAt,
		}
		// Only reveal student identity once they have accepted
		if r.Status == "accepted" {
			sr.StudentEmail = r.StudentEmail
			sr.StudentID = r.StudentID
		}
		safe = append(safe, sr)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"requests": safe,
		"total":    len(safe),
	})
}

// =============================================================================
// GET ACCEPTED PROFILES
// =============================================================================

// GetAcceptedProfiles returns full profiles for students who accepted requests.
func GetAcceptedProfiles(c *gin.Context) {
	companyID, _, _, exists := middleware.GetCompanyFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid company identity"})
		return
	}

	var requests []models.ProfileRequest
	if err := database.DB.
		Where("company_id = ? AND status = ?", companyUUID, "accepted").
		Find(&requests).Error; err != nil {
		log.Printf("Error fetching accepted requests for company %s: %v", companyID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
		return
	}

	// Guard against empty IN clause — some DB drivers error on IN ()
	if len(requests) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"profiles": []interface{}{},
			"total":    0,
		})
		return
	}

	emails := make([]string, len(requests))
	for i, r := range requests {
		emails[i] = r.StudentEmail
	}

	// Fetch all students in one query — exclude soft-deleted
	var students []models.Student
	if err := database.DB.
		Where("email IN ? AND deleted_at IS NULL", emails).
		Find(&students).Error; err != nil {
		log.Printf("Error fetching accepted students: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
		return
	}

	if len(students) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"profiles": []interface{}{},
			"total":    0,
		})
		return
	}

	// Collect IDs and emails for batch lookups
	studentIDs := make([]uuid.UUID, len(students))
	sEmails := make([]string, len(students))
	for i, s := range students {
		studentIDs[i] = s.ID
		sEmails[i] = s.Email
	}

	// --- Batch load skills ---
	var allSkills []models.StudentSkill
	database.DB.Where("student_id IN ?", studentIDs).Find(&allSkills)
	skillMap := make(map[uuid.UUID][]models.StudentSkill)
	for _, sk := range allSkills {
		skillMap[sk.StudentID] = append(skillMap[sk.StudentID], sk)
	}

	// --- Batch load education ---
	var allEducation []models.StudentEducation
	database.DB.Preload("University").Where("student_id IN ?", studentIDs).Find(&allEducation)
	eduMap := make(map[uuid.UUID][]models.StudentEducation)
	for _, ed := range allEducation {
		eduMap[ed.StudentID] = append(eduMap[ed.StudentID], ed)
	}

	// --- Batch load cert counts ---
	type certCount struct {
		StudentEmail string
		Count        int64
	}
	var certCounts []certCount
	database.DB.Model(&models.Certificate{}).
		Select("student_email, COUNT(*) as count").
		Where("student_email IN ? AND pdf_path != ''", emails).
		Group("student_email").
		Scan(&certCounts)

	certMap := make(map[string]int64)
	for _, cc := range certCounts {
		certMap[cc.StudentEmail] = cc.Count
	}

	// --- Assemble full profiles ---
	type fullProfile struct {
		Student   models.Student            `json:"student"`
		Skills    []models.StudentSkill     `json:"skills"`
		Education []models.StudentEducation `json:"education"`
		CertCount int64                     `json:"certificates_count"`
	}

	profiles := make([]fullProfile, 0, len(students))
	for _, s := range students {
		profiles = append(profiles, fullProfile{
			Student:   s,
			Skills:    skillMap[s.ID],
			Education: eduMap[s.ID],
			CertCount: certMap[s.Email],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"profiles": profiles,
		"total":    len(profiles),
	})
}
