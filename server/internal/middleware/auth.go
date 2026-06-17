package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"cert-system/server/internal/database"
	"cert-system/server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

// InitJWT must be called from main() after godotenv.Load().
// Panicking here is intentional — a missing JWT secret is a hard startup error.
func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is not set — refusing to start")
	}
	if len(secret) < 32 {
		panic(fmt.Sprintf("JWT_SECRET must be at least 32 characters (got %d)", len(secret)))
	}
	jwtSecret = []byte(secret)
}

const (
	RoleUniversity = "university"
	RoleCompany    = "company"
	RoleStudent    = "student"
)

type Claims struct {
	ActorID      string `json:"actor_id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

func newClaims(actorID, email, name, role string, tokenVersion int) Claims {
	now := time.Now()
	return Claims{
		ActorID:      actorID,
		Email:        email,
		Name:         name,
		Role:         role,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   actorID,
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
}

func GenerateToken(universityID, email, name string, tokenVersion int) (string, error) {
	claims := newClaims(universityID, email, name, RoleUniversity, tokenVersion)
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func GenerateCompanyToken(companyID, email, name string, tokenVersion int) (string, error) {
	claims := newClaims(companyID, email, name, RoleCompany, tokenVersion)
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func GenerateStudentToken(studentID, email, name string, tokenVersion int) (string, error) {
	claims := newClaims(studentID, email, name, RoleStudent, tokenVersion)
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func extractBearer(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization header required")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization format — use: Bearer <token>")
	}
	return parts[1], nil
}

func parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret, nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func validateTokenVersion(claims *Claims) error {
	switch claims.Role {
	case RoleStudent:
		var s models.Student
		if err := database.DB.Select("token_version").
			Where("id = ? AND deleted_at IS NULL", claims.ActorID).
			First(&s).Error; err != nil {
			return errors.New("user not found")
		}
		if s.TokenVersion != claims.TokenVersion {
			return errors.New("session expired")
		}
	case RoleCompany:
		var co models.Company
		if err := database.DB.Select("token_version").
			Where("id = ?", claims.ActorID).
			First(&co).Error; err != nil {
			return errors.New("user not found")
		}
		if co.TokenVersion != claims.TokenVersion {
			return errors.New("session expired")
		}
	case RoleUniversity:
		var u models.University
		if err := database.DB.Select("token_version").
			Where("id = ?", claims.ActorID).
			First(&u).Error; err != nil {
			return errors.New("user not found")
		}
		if u.TokenVersion != claims.TokenVersion {
			return errors.New("session expired")
		}
	}
	return nil
}

func requireRole(role, idKey, emailKey, nameKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearer(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		claims, err := parseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}
		if claims.Role != role {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied — token role does not match required role",
			})
			c.Abort()
			return
		}
		if err := validateTokenVersion(claims); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired, please log in again"})
			c.Abort()
			return
		}
		c.Set(idKey, claims.ActorID)
		c.Set(emailKey, claims.Email)
		c.Set(nameKey, claims.Name)
		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return requireRole(RoleUniversity, "university_id", "email", "name")
}

func CompanyAuthMiddleware() gin.HandlerFunc {
	return requireRole(RoleCompany, "company_id", "company_email", "company_name")
}

func StudentAuthMiddleware() gin.HandlerFunc {
	return requireRole(RoleStudent, "student_id", "student_email", "student_name")
}

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearer(c)
		if err != nil {
			c.Next()
			return
		}
		claims, err := parseToken(tokenString)
		if err != nil {
			c.Next()
			return
		}
		switch claims.Role {
		case RoleUniversity:
			c.Set("university_id", claims.ActorID)
			c.Set("email", claims.Email)
			c.Set("name", claims.Name)
		case RoleCompany:
			c.Set("company_id", claims.ActorID)
			c.Set("company_email", claims.Email)
			c.Set("company_name", claims.Name)
		case RoleStudent:
			c.Set("student_id", claims.ActorID)
			c.Set("student_email", claims.Email)
			c.Set("student_name", claims.Name)
		}
		c.Set("actor_role", claims.Role)
		c.Next()
	}
}

func GetUserFromContext(c *gin.Context) (universityID, email, name string, exists bool) {
	id, ok1 := c.Get("university_id")
	em, ok2 := c.Get("email")
	nm, ok3 := c.Get("name")
	if !ok1 || !ok2 || !ok3 {
		return "", "", "", false
	}
	return id.(string), em.(string), nm.(string), true
}

func GetCompanyFromContext(c *gin.Context) (companyID, email, name string, exists bool) {
	id, ok1 := c.Get("company_id")
	em, ok2 := c.Get("company_email")
	nm, ok3 := c.Get("company_name")
	if !ok1 || !ok2 || !ok3 {
		return "", "", "", false
	}
	return id.(string), em.(string), nm.(string), true
}

func GetStudentFromContext(c *gin.Context) (string, string, string, bool) {
	id, idOK := c.Get("student_id")
	email, emailOK := c.Get("student_email")
	name, _ := c.Get("student_name")
	if !idOK || !emailOK {
		return "", "", "", false
	}
	nameStr, _ := name.(string)
	return id.(string), email.(string), nameStr, true
}
