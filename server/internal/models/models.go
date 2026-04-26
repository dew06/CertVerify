package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"golang.org/x/crypto/bcrypt"
)

type University struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email               string    `gorm:"unique;not null"`
	Name                string    `gorm:"not null"`
	Domain              string    `gorm:"unique;not null"` // e.g., "mit.edu"
	Password            string    `gorm:"not null" json:"-"`
	CardanoPublicKey    string    `gorm:"not null"`
	EncryptedPrivateKey string    `gorm:"not null"` // AES-256 encrypted
	EncryptionSalt      string    `gorm:"not null"`
	IsVerified          bool      `gorm:"default:false"` // Domain verification
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Certificate represents a student credential
type Certificate struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CertID       string    `gorm:"unique;not null;index"`
	UniversityID uuid.UUID `gorm:"not null"`

	// Student Information (stored in PostgreSQL for fast queries)
	StudentName     string `gorm:"not null;index"`
	StudentEmail    string `gorm:"not null"`
	Degree          string `gorm:"not null;index"`
	GPA             float64
	IssueDate       time.Time      `gorm:"index"`
	Age             *int           `gorm:"type:integer"`
	Gender          string         `gorm:"type:varchar(20)"`
	Nationality     string         `gorm:"type:varchar(100)"`
	Phone           string         `gorm:"type:varchar(50)"`
	LinkedInURL     string         `gorm:"type:varchar(255);column:linkedin_url"`
	Skills          pq.StringArray `gorm:"type:text[]"`
	ExperienceYears int            `gorm:"type:integer"`
	IsSearchable    bool           `gorm:"default:true"`

	// Cryptographic Data
	CertificateHash string `gorm:"not null"` // SHA-256 of cert data
	PDFHash         string `gorm:"not null"` // SHA-256 of PDF file
	Salt            string `gorm:"not null"`

	// Storage & Blockchain References
	CardanoTxID string `gorm:"index"`
	IPFSPDFHash string `gorm:"not null"`
	PDFPath     string

	// Merkle Tree (Bulk Anchoring)
	MerkleRootHash   string `gorm:"index"`
	MerkleProof      string `gorm:"type:text"`
	BatchAnchorID    *uuid.UUID
	BlockchainStatus string `gorm:"default:'pending'"`
	BlockchainTxHash string
	EmailSent        bool `gorm:"default:false"`
	MerkleRoot       string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Company struct {
	ID          uuid.UUID
	Name        string
	Email       string
	Password    string
	Industry    string
	CompanySize string
	Location    string
	Website     string
	Description string
	LogoURL     string
	IsVerified  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// User represents an admin
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UniversityID uuid.UUID `gorm:"not null"`
	Email        string    `gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"default:'issuer'"`
	CreatedAt    time.Time
}

// BatchAnchor represents a Merkle tree batch
type BatchAnchor struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UniversityID     uuid.UUID `gorm:"not null"`
	MerkleRootHash   string    `gorm:"not null"`
	CardanoTxID      string    `gorm:"not null"`
	CertificateCount int       `gorm:"not null"`
	CreatedAt        time.Time
}

type ProfileRequest struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID     uuid.UUID  `gorm:"type:uuid;not null" json:"company_id"`
	CertificateID uuid.UUID  `gorm:"type:uuid;not null" json:"certificate_id"`
	StudentEmail  string     `gorm:"not null" json:"student_email"`
	Status        string     `gorm:"default:'pending'" json:"status"`
	Message       string     `json:"message"`
	RequestedAt   time.Time  `json:"requested_at"`
	RespondedAt   *time.Time `json:"responded_at"`
	ExpiresAt     time.Time  `json:"expires_at"`

	// Relations - ADD THESE IF MISSING
	Company     Company     `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Certificate Certificate `gorm:"foreignKey:CertificateID" json:"certificate,omitempty"`
}

// StudentProfileHidden represents a student profile with hidden details
type StudentProfileHidden struct {
	ID               uuid.UUID      `json:"id"`
	CertID           string         `json:"cert_id"`
	Degree           string         `json:"degree"`
	GPA              float64        `json:"gpa"`
	Skills           pq.StringArray `json:"skills" gorm:"type:text[]"`
	ExperienceYears  int            `json:"experience_years"`
	IssueDate        time.Time      `json:"issue_date"`
	UniversityID     uuid.UUID      `json:"university_id"`
	UniversityName   string         `json:"university_name"`
	UniversityDomain string         `json:"university_domain"`
	BlockchainStatus string         `json:"blockchain_status"`
	CardanoTxID      string         `json:"cardano_tx_id"`

	// Hidden fields
	StudentName  string `json:"student_name"`  // Will be "***"
	StudentEmail string `json:"student_email"` // Will be "***"
	Age          *int   `json:"age"`           // Will be null
	Gender       string `json:"gender"`        // Will be "***"
	Nationality  string `json:"nationality"`   // Will be "***"
	Phone        string `json:"phone"`         // Will be "***"
	LinkedInURL  string `json:"linkedin_url"`  // Will be "***"
}

// StudentProfileFull represents a student profile with all details
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
	Skills             pq.StringArray `json:"skills" gorm:"type:text[]"`
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

// SearchFilters for company job search
type SearchFilters struct {
	Degree        string   `json:"degree"`
	MinGPA        float64  `json:"min_gpa"`
	MaxGPA        float64  `json:"max_gpa"`
	Skills        []string `json:"skills"`
	MinExperience int      `json:"min_experience"`
	UniversityID  string   `json:"university_id"`
	Nationality   string   `json:"nationality"`
	Gender        string   `json:"gender"`
	Limit         int      `json:"limit"`
	Offset        int      `json:"offset"`
}

// CheckPassword verifies the password using bcrypt
func (u *University) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// HashPassword hashes a password
func (u *University) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies company password
func (c *Company) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(c.Password), []byte(password))
	return err == nil
}

// HashPassword hashes the company password
func (c *Company) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	c.Password = string(hashedPassword)
	return nil
}
