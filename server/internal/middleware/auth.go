package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-super-secret-key-change-this-in-production"
	}
	jwtSecret = []byte(secret)
}

// Claims represents JWT claims
type Claims struct {
	UniversityID string `json:"university_id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token
func GenerateToken(universityID, email, name string) (string, error) {
	claims := Claims{
		UniversityID: universityID,
		Email:        email,
		Name:         name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// AuthMiddleware validates JWT token for universities
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Use: Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			// Set user info in context
			c.Set("university_id", claims.UniversityID)
			c.Set("email", claims.Email)
			c.Set("name", claims.Name)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
	}
}

// OptionalAuthMiddleware allows both authenticated and unauthenticated requests
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]
				token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})

				if err == nil {
					if claims, ok := token.Claims.(*Claims); ok && token.Valid {
						c.Set("university_id", claims.UniversityID)
						c.Set("email", claims.Email)
						c.Set("name", claims.Name)
					}
				}
			}
		}
		c.Next()
	}
}

// GetUserFromContext extracts user info from context
func GetUserFromContext(c *gin.Context) (universityID string, email string, name string, exists bool) {
	universityIDVal, ok1 := c.Get("university_id")
	emailVal, ok2 := c.Get("email")
	nameVal, ok3 := c.Get("name")

	if !ok1 || !ok2 || !ok3 {
		return "", "", "", false
	}

	return universityIDVal.(string), emailVal.(string), nameVal.(string), true
}

// ========== COMPANY AUTHENTICATION ==========

// CompanyClaims represents JWT claims for companies
type CompanyClaims struct {
	CompanyID string `json:"company_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	jwt.RegisteredClaims
}

// GenerateCompanyToken generates a new JWT token for company
func GenerateCompanyToken(companyID, email, name string) (string, error) {
	claims := CompanyClaims{
		CompanyID: companyID,
		Email:     email,
		Name:      name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// CompanyAuthMiddleware validates JWT token for companies
func CompanyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Use: Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &CompanyClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*CompanyClaims); ok && token.Valid {
			// Set company info in context
			c.Set("company_id", claims.CompanyID)
			c.Set("company_email", claims.Email)
			c.Set("company_name", claims.Name)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
	}
}

// GetCompanyFromContext extracts company info from context
func GetCompanyFromContext(c *gin.Context) (companyID string, email string, name string, exists bool) {
	companyIDVal, ok1 := c.Get("company_id")
	emailVal, ok2 := c.Get("company_email")
	nameVal, ok3 := c.Get("company_name")

	if !ok1 || !ok2 || !ok3 {
		return "", "", "", false
	}

	return companyIDVal.(string), emailVal.(string), nameVal.(string), true
}
