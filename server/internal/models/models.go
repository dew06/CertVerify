package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// =============================================================================
// UNIVERSITY
// =============================================================================

type University struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey"        json:"id"`
	Email               string    `gorm:"uniqueIndex;not null"        json:"email"`
	Name                string    `gorm:"not null"                    json:"name"`
	Domain              string    `gorm:"uniqueIndex;not null"        json:"domain"`
	Password            string    `gorm:"not null"                    json:"-"`
	CardanoPublicKey    string    `gorm:"not null"                    json:"cardano_public_key"`
	EncryptedPrivateKey string    `gorm:"not null"                    json:"-"`
	EncryptionSalt      string    `gorm:"not null"                    json:"-"`
	IsVerified          bool      `gorm:"default:false"               json:"is_verified"`
	TokenVersion        int       `gorm:"default:1" json:"-"`
	CreatedAt           time.Time `                                   json:"created_at"`
	UpdatedAt           time.Time `                                   json:"updated_at"`
}

func (u *University) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (u *University) HashPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *University) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

func (University) TableName() string { return "universities" }

// =============================================================================
// COMPANY
// =============================================================================

type Company struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"        json:"id"`
	Email        string    `gorm:"uniqueIndex;not null"        json:"email"`
	Name         string    `gorm:"not null"                    json:"name"`
	Password     string    `gorm:"not null"                    json:"-"`
	Industry     string    `                                   json:"industry"`
	CompanySize  string    `                                   json:"company_size"`
	Location     string    `                                   json:"location"`
	Website      string    `                                   json:"website"`
	Description  string    `                                   json:"description"`
	LogoURL      string    `                                   json:"logo_url"`
	IsVerified   bool      `gorm:"default:false"               json:"is_verified"`
	TokenVersion int       `gorm:"default:1" json:"-"`
	CreatedAt    time.Time `                                   json:"created_at"`
	UpdatedAt    time.Time `                                   json:"updated_at"`
}

func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (c *Company) HashPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	c.Password = string(hashed)
	return nil
}

func (c *Company) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(c.Password), []byte(password)) == nil
}

func (Company) TableName() string { return "companies" }

// =============================================================================
// STUDENT
// =============================================================================

type Student struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"                    json:"id"`
	Email             string    `gorm:"uniqueIndex;not null"                    json:"email"`
	Password          string    `gorm:"not null"                                json:"-"`
	Name              string    `gorm:"not null"                                json:"name"`
	Phone             string    `                                               json:"phone"`
	Age               *int      `                                               json:"age"`
	Gender            string    `                                               json:"gender"`
	Nationality       string    `                                               json:"nationality"`
	LinkedInURL       string    `gorm:"column:linkedin_url"                     json:"linkedin_url"`
	ProfilePictureURL string    `gorm:"column:profile_picture_url"              json:"profile_picture_url"`
	TokenVersion      int       `gorm:"default:1" json:"-"`

	// Privacy settings
	IsSearchable    bool `gorm:"default:true"                            json:"is_searchable"`
	ShowName        bool `gorm:"default:false"                           json:"show_name"`
	ShowEmail       bool `gorm:"default:false"                           json:"show_email"`
	ShowPhone       bool `gorm:"default:false"                           json:"show_phone"`
	ShowAge         bool `gorm:"default:false"                           json:"show_age"`
	ShowGender      bool `gorm:"default:false"                           json:"show_gender"`
	ShowNationality bool `gorm:"default:false"                           json:"show_nationality"`
	ShowLinkedIn    bool `gorm:"column:show_linkedin;default:false"       json:"show_linkedin"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login"`

	// Soft delete
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *Student) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *Student) HashPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.Password = string(hashed)
	return nil
}

func (s *Student) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(s.Password), []byte(password)) == nil
}

func (s *Student) ApplyPrivacy() StudentPublicProfile {
	p := StudentPublicProfile{
		ID:           s.ID,
		IsSearchable: s.IsSearchable,
		CreatedAt:    s.CreatedAt,
	}
	if s.ShowName {
		p.Name = s.Name
	}
	if s.ShowEmail {
		p.Email = s.Email
	}
	if s.ShowPhone {
		p.Phone = s.Phone
	}
	if s.ShowAge {
		p.Age = s.Age
	}
	if s.ShowGender {
		p.Gender = s.Gender
	}
	if s.ShowNationality {
		p.Nationality = s.Nationality
	}
	if s.ShowLinkedIn {
		p.LinkedInURL = s.LinkedInURL
	}
	return p
}

func (Student) TableName() string { return "students" }

// =============================================================================
// STUDENT SKILL
// =============================================================================

type StudentSkill struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"        json:"id"`
	StudentID         uuid.UUID `gorm:"type:uuid;not null;index"    json:"student_id"`
	SkillName         string    `gorm:"not null"                    json:"skill_name"`
	ProficiencyLevel  string    `                                   json:"proficiency_level"`
	YearsOfExperience int       `                                   json:"years_of_experience"`
	CreatedAt         time.Time `                                   json:"created_at"`
}

func (s *StudentSkill) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (StudentSkill) TableName() string { return "student_skills" }

// =============================================================================
// STUDENT EDUCATION
// =============================================================================

type StudentEducation struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"        json:"id"`
	StudentID    uuid.UUID  `gorm:"type:uuid;not null;index"    json:"student_id"`
	UniversityID *uuid.UUID `gorm:"type:uuid;index"             json:"university_id"`
	Degree       string     `gorm:"not null"                    json:"degree"`
	FieldOfStudy string     `                                   json:"field_of_study"`
	GPA          float64    `                                   json:"gpa"`
	StartDate    *time.Time `                                   json:"start_date"`
	EndDate      *time.Time `                                   json:"end_date"`
	IsCurrent    bool       `gorm:"default:false"               json:"is_current"`
	CreatedAt    time.Time  `                                   json:"created_at"`

	University *University `gorm:"foreignKey:UniversityID"     json:"university,omitempty"`
}

func (e *StudentEducation) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (StudentEducation) TableName() string { return "student_education" }

// =============================================================================
// CERTIFICATE
// =============================================================================

type Certificate struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"        json:"id"`
	CertID       string    `gorm:"uniqueIndex;not null"        json:"cert_id"`
	UniversityID uuid.UUID `gorm:"type:uuid;not null;index"    json:"university_id"`

	StudentName     string         `gorm:"not null;index"                        json:"student_name"`
	StudentEmail    string         `gorm:"not null;index"                        json:"student_email"`
	Degree          string         `gorm:"not null;index"                        json:"degree"`
	GPA             float64        `                                             json:"gpa"`
	IssueDate       time.Time      `gorm:"index"                                 json:"issue_date"`
	Age             *int           `gorm:"type:integer"                          json:"age"`
	Gender          string         `gorm:"type:varchar(20)"                      json:"gender"`
	Nationality     string         `gorm:"type:varchar(100)"                     json:"nationality"`
	Phone           string         `gorm:"type:varchar(50)"                      json:"phone"`
	LinkedInURL     string         `gorm:"type:varchar(255);column:linkedin_url" json:"linkedin_url"`
	Skills          pq.StringArray `gorm:"type:text[]"                           json:"skills"`
	ExperienceYears int            `gorm:"type:integer"                          json:"experience_years"`
	IsSearchable    bool           `gorm:"default:true"                          json:"is_searchable"`

	CertificateHash string `gorm:"not null"    json:"certificate_hash"`
	PDFHash         string `gorm:"not null"    json:"pdf_hash"`
	Salt            string `gorm:"not null"    json:"-"`

	CardanoTxID string `gorm:"index"       json:"cardano_tx_id"`
	IPFSPDFHash string `gorm:"not null"    json:"ipfs_pdf_hash"`
	PDFPath     string `                   json:"pdf_path,omitempty"`

	MerkleRootHash   string     `gorm:"index"           json:"merkle_root_hash"`
	MerkleProof      string     `gorm:"type:text"       json:"merkle_proof"`
	MerkleRoot       string     `                       json:"merkle_root"`
	BatchAnchorID    *uuid.UUID `gorm:"type:uuid;index" json:"batch_anchor_id"`
	BlockchainStatus string     `gorm:"default:'pending'" json:"blockchain_status"`
	BlockchainTxHash string     `                       json:"blockchain_tx_hash"`
	EmailSent        bool       `gorm:"default:false"   json:"email_sent"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cert *Certificate) BeforeCreate(tx *gorm.DB) error {
	if cert.ID == uuid.Nil {
		cert.ID = uuid.New()
	}
	return nil
}

func (Certificate) TableName() string { return "certificates" }

// =============================================================================
// USER (admin / issuer)
// =============================================================================

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"        json:"id"`
	UniversityID uuid.UUID `gorm:"type:uuid;not null;index"    json:"university_id"`
	Email        string    `gorm:"uniqueIndex;not null"        json:"email"`
	PasswordHash string    `gorm:"not null"                    json:"-"`
	Role         string    `gorm:"default:'issuer'"            json:"role"`
	CreatedAt    time.Time `                                   json:"created_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (User) TableName() string { return "users" }

// =============================================================================
// BATCH ANCHOR
// =============================================================================

type BatchAnchor struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"        json:"id"`
	UniversityID     uuid.UUID `gorm:"type:uuid;not null;index"    json:"university_id"`
	MerkleRootHash   string    `gorm:"not null"                    json:"merkle_root_hash"`
	CardanoTxID      string    `gorm:"not null"                    json:"cardano_tx_id"`
	CertificateCount int       `gorm:"not null"                    json:"certificate_count"`
	CreatedAt        time.Time `                                   json:"created_at"`
}

func (b *BatchAnchor) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

func (BatchAnchor) TableName() string { return "batch_anchors" }

// =============================================================================
// PROFILE REQUEST
// =============================================================================

type ProfileRequest struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"        json:"id"`
	CompanyID     uuid.UUID  `gorm:"type:uuid;not null;index"    json:"company_id"`
	CertificateID *uuid.UUID `json:"certificate_id" gorm:"type:uuid"`
	StudentID     *uuid.UUID `gorm:"type:uuid;index"             json:"student_id"`
	StudentEmail  string     `gorm:"not null;index"              json:"student_email"`
	Status        string     `gorm:"default:'pending';index"     json:"status"`
	Message       string     `                                   json:"message"`
	RequestedAt   time.Time  `                                   json:"requested_at"`
	RespondedAt   *time.Time `                                   json:"responded_at"`
	ExpiresAt     time.Time  `                                   json:"expires_at"`

	Company     Company     `gorm:"foreignKey:CompanyID"     json:"company,omitempty"`
	Certificate Certificate `gorm:"foreignKey:CertificateID" json:"certificate,omitempty"`
}

func (p *ProfileRequest) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (ProfileRequest) TableName() string { return "profile_requests" }

// =============================================================================
// VIEW TYPES — not stored in DB
// =============================================================================

type StudentPublicProfile struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name,omitempty"`
	Email        string    `json:"email,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	Age          *int      `json:"age,omitempty"`
	Gender       string    `json:"gender,omitempty"`
	Nationality  string    `json:"nationality,omitempty"`
	LinkedInURL  string    `json:"linkedin_url,omitempty"`
	IsSearchable bool      `json:"is_searchable"`
	CreatedAt    time.Time `json:"created_at"`
}

type StudentProfileHidden struct {
	ID               uuid.UUID      `json:"id"`
	CertID           string         `json:"cert_id"`
	Degree           string         `json:"degree"`
	GPA              float64        `json:"gpa"`
	Skills           pq.StringArray `json:"skills"            gorm:"type:text[]"`
	ExperienceYears  int            `json:"experience_years"`
	IssueDate        time.Time      `json:"issue_date"`
	UniversityID     uuid.UUID      `json:"university_id"`
	UniversityName   string         `json:"university_name"`
	UniversityDomain string         `json:"university_domain"`
	BlockchainStatus string         `json:"blockchain_status"`
	CardanoTxID      string         `json:"cardano_tx_id"`

	StudentName  string `json:"student_name,omitempty"`
	StudentEmail string `json:"student_email,omitempty"`
	Age          *int   `json:"age,omitempty"`
	Gender       string `json:"gender,omitempty"`
	Nationality  string `json:"nationality,omitempty"`
	Phone        string `json:"phone,omitempty"`
	LinkedInURL  string `json:"linkedin_url,omitempty"`
}

type StudentProfileFull struct {
	ID                 uuid.UUID      `json:"id"`
	CertID             string         `json:"cert_id"`
	StudentName        string         `json:"student_name"`
	StudentEmail       string         `json:"student_email"`
	Degree             string         `json:"degree"`
	GPA                float64        `json:"gpa"`
	Age                *int           `json:"age"`
	Gender             string         `json:"gender"`
	Nationality        string         `json:"nationality"`
	Phone              string         `json:"phone"`
	LinkedInURL        string         `json:"linkedin_url"`
	Skills             pq.StringArray `json:"skills"            gorm:"type:text[]"`
	ExperienceYears    int            `json:"experience_years"`
	IssueDate          time.Time      `json:"issue_date"`
	UniversityID       uuid.UUID      `json:"university_id"`
	UniversityName     string         `json:"university_name"`
	UniversityDomain   string         `json:"university_domain"`
	UniversityVerified bool           `json:"university_verified"`
	BlockchainStatus   string         `json:"blockchain_status"`
	CardanoTxID        string         `json:"cardano_tx_id"`
	IPFSPDFHash        string         `json:"ipfs_pdf_hash"`
}

type StudentProfile struct {
	Student      Student            `json:"student"`
	Skills       []StudentSkill     `json:"skills"`
	Education    []StudentEducation `json:"education"`
	Certificates []Certificate      `json:"certificates,omitempty"`
}

type SearchFilters struct {
	Degree        string   `json:"degree"`
	MinGPA        float64  `json:"min_gpa"`
	MaxGPA        float64  `json:"max_gpa"`
	Skills        []string `json:"skills"`
	MinExperience int      `json:"min_experience"`
	UniversityID  string   `json:"university_id"`
	Limit         int      `json:"limit"`
	Offset        int      `json:"offset"`
}
