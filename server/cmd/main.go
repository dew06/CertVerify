package main

import (
	"log"
	"os"

	"cert-system/server/internal/database"
	"cert-system/server/internal/handlers"
	"cert-system/server/internal/middleware"
	"cert-system/server/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Certificate System API
// @version 1.0
// @description Blockchain-based certificate issuance and verification system
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No .env file found")
	}

	// Connect to database
	database.Initialize()

	// Create router
	router := gin.Default()

	// CORS for frontend
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Get IPFS URL
	ipfsURL := os.Getenv("IPFS_API_URL")
	if ipfsURL == "" {
		ipfsURL = "localhost:5001"
	}

	// Initialize handlers
	universityHandler := handlers.NewUniversityHandler()
	certificateHandler := handlers.NewCertificateHandler(ipfsURL)
	verificationHandler := handlers.NewVerificationHandler()
	batchHandler := handlers.NewBatchAnchorHandler()

	// API Routes
	api := router.Group("/api")
	{
		// ========== PUBLIC ROUTES (No Authentication) ==========

		// Authentication
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.LoginHandler(database.DB))
			auth.POST("/logout", handlers.LogoutHandler())

			// Protected auth routes
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware())
			{
				authProtected.GET("/me", handlers.GetCurrentUserHandler(database.DB))
				authProtected.POST("/refresh", handlers.RefreshTokenHandler(database.DB))
				authProtected.POST("/change-password", handlers.ChangePasswordHandler(database.DB))
			}
		}

		// University Registration (Public)
		api.POST("/university/register", universityHandler.Register)

		// Verification (Public)
		verify := api.Group("/verify")
		{
			verify.POST("", verificationHandler.Verify)
			verify.POST("/pdf", verificationHandler.VerifyPDF)
			verify.GET("/blockchain/:certID", verificationHandler.GetBlockchainInfo)
			verify.GET("/certificate/:id", verificationHandler.VerifyByID)
		}

		company := api.Group("/company")
		{
			// Public company routes
			company.POST("/register", handlers.RegisterCompany)
			company.POST("/login", handlers.LoginCompany)

			// Protected company routes
			companyProtected := company.Group("")
			companyProtected.Use(middleware.CompanyAuthMiddleware())
			{
				companyProtected.GET("/me", handlers.GetCurrentCompany)
				companyProtected.POST("/search", handlers.SearchStudents)
				companyProtected.POST("/request-profile", handlers.RequestProfileAccess)
				companyProtected.GET("/my-requests", handlers.GetMyRequests)
				companyProtected.GET("/accepted-profiles", handlers.GetAcceptedProfiles)
			}
		}

		student := api.Group("/student")
		{
			// Public student routes (email-based, no JWT required)
			student.GET("/profile-requests", handlers.GetMyProfileRequests)
			student.POST("/respond-request", handlers.RespondToRequest)
			student.POST("/profile-visibility", handlers.UpdateProfileVisibility)
			student.GET("/profile-settings", handlers.GetMyProfileSettings)
		}

		// ========== PROTECTED ROUTES (Require Authentication) ==========

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// University Management
			university := protected.Group("/university")
			{
				university.GET("/:id/domain-proof", universityHandler.GetDomainVerification)
				university.POST("/:id/verify-domain", universityHandler.VerifyDomain)
			}

			// Certificates
			certificates := protected.Group("/certificates")
			{
				certificates.POST("/issue", certificateHandler.Issue)
				certificates.POST("/batch-csv", certificateHandler.BatchIssueCSV)
				certificates.POST("/batch-anchor", batchHandler.AnchorBatch)

				// Bulk download
				certificates.GET("/bulk-download", handlers.BulkDownloadCertificatesHandler(database.DB))

				// Dynamic routes
				certificates.GET("/:certID", certificateHandler.Get)
				certificates.GET("/:certID/download", certificateHandler.DownloadPDF)
			}
		}

		// Email test endpoint (protected)
		protected.POST("/test-email", func(c *gin.Context) {
			emailService := services.NewEmailService()
			err := emailService.TestEmailConnection()
			if err != nil {
				c.JSON(500, gin.H{"error": "Email not working: " + err.Error()})
				return
			}
			c.JSON(200, gin.H{"message": "✅ Email service is working!"})
		})
	}

	// Health check (public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "healthy",
			"database": "connected",
			"ipfs":     ipfsURL,
		})
	})

	// Serve swagger.json at different path to avoid conflict
	router.StaticFile("/api-docs/swagger.json", "./docs/swagger.json")

	// Configure Swagger UI to use the custom path
	url := ginSwagger.URL("/api-docs/swagger.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("═══════════════════════════════════════")
	log.Println("🎓 Certificate System Server")
	log.Println("═══════════════════════════════════════")
	log.Printf("🚀 Server: http://localhost:%s", port)
	log.Printf("📊 Database: Connected")
	log.Printf("🌐 IPFS: %s", ipfsURL)
	log.Printf("📧 Email: %s", func() string {
		if os.Getenv("SMTP_HOST") != "" {
			return "Configured ✅"
		}
		return "Not configured ⚠️"
	}())
	log.Printf("🔐 JWT: %s", func() string {
		if os.Getenv("JWT_SECRET") != "" {
			return "Custom secret ✅"
		}
		return "Default secret ⚠️"
	}())
	log.Println("═══════════════════════════════════════════════════")
	log.Println("📝 UNIVERSITY Routes:")
	log.Println("   Public:")
	log.Println("     POST   /api/university/register")
	log.Println("     POST   /api/auth/login")
	log.Println("   Protected:")
	log.Println("     GET    /api/auth/me")
	log.Println("     POST   /api/certificates/issue")
	log.Println("     POST   /api/certificates/batch-csv")
	log.Println("     POST   /api/certificates/batch-anchor")
	log.Println("═══════════════════════════════════════════════════")
	log.Println("📝 COMPANY Routes:")
	log.Println("   Public:")
	log.Println("     POST   /api/company/register")
	log.Println("     POST   /api/company/login")
	log.Println("   Protected:")
	log.Println("     GET    /api/company/me")
	log.Println("     POST   /api/company/search")
	log.Println("     POST   /api/company/request-profile")
	log.Println("     GET    /api/company/my-requests")
	log.Println("     GET    /api/company/accepted-profiles")
	log.Println("═══════════════════════════════════════════════════")
	log.Println("📝 STUDENT Routes:")
	log.Println("   Public:")
	log.Println("     GET    /api/student/profile-requests")
	log.Println("     POST   /api/student/respond-request")
	log.Println("     POST   /api/student/profile-visibility")
	log.Println("     GET    /api/student/profile-settings")
	log.Println("═══════════════════════════════════════════════════")
	log.Println("📝 VERIFICATION Routes:")
	log.Println("   Public:")
	log.Println("     POST   /api/verify")
	log.Println("     POST   /api/verify/pdf")
	log.Println("     GET    /api/verify/blockchain/:certID")
	log.Println("═══════════════════════════════════════════════════")

	router.Run(":" + port)
}
