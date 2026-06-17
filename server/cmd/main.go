package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
// @version 2.0
// @description Blockchain-based certificate issuance and verification system with privacy-first job matching
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
	// -------------------------------------------------------------------------
	// Environment
	// -------------------------------------------------------------------------
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found — using system environment variables")
	}

	// -------------------------------------------------------------------------
	// Database
	// -------------------------------------------------------------------------
	database.Initialize()

	// -------------------------------------------------------------------------
	// Router
	// -------------------------------------------------------------------------
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	if err := router.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Fatalf("❌ Failed to set trusted proxies: %v", err)
	}

	// Limit request body to 10 MB — prevents large-payload DoS
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)
		c.Next()
	})

	// -------------------------------------------------------------------------
	// CORS — reads CORS_ORIGINS from env (comma-separated); falls back to dev
	// -------------------------------------------------------------------------
	router.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// -------------------------------------------------------------------------
	// Services & handlers
	// -------------------------------------------------------------------------
	ipfsURL := getEnvOrDefault("IPFS_API_URL", "http://localhost:5001")

	universityHandler := handlers.NewUniversityHandler()
	certificateHandler := handlers.NewCertificateHandler(ipfsURL)
	verificationHandler := handlers.NewVerificationHandler()
	batchHandler := handlers.NewBatchAnchorHandler()

	// -------------------------------------------------------------------------
	// Routes
	// -------------------------------------------------------------------------
	api := router.Group("/api")

	// ── Auth (unified login — works for university, company, and student) ────
	auth := api.Group("/auth")
	{
		auth.POST("/login", handlers.UnifiedLoginHandler(database.DB))

		// University-scoped protected routes
		authProtected := auth.Group("")
		authProtected.Use(middleware.AuthMiddleware())
		{
			authProtected.GET("/me", handlers.GetCurrentUserHandler(database.DB))
			authProtected.POST("/refresh", handlers.RefreshTokenHandler(database.DB))
			authProtected.POST("/change-password", handlers.ChangePasswordHandler(database.DB))
			authProtected.POST("/logout", handlers.LogoutHandler(database.DB))
		}
	}

	// ── University ────────────────────────────────────────────────────────────
	api.POST("/university/register", universityHandler.Register)

	uni := api.Group("")
	uni.Use(middleware.AuthMiddleware())
	{
		university := uni.Group("/university")
		{
			university.GET("/:id/domain-proof", universityHandler.GetDomainVerification)
			university.POST("/:id/verify-domain", universityHandler.VerifyDomain)
		}

		certificates := uni.Group("/certificates")
		{
			certificates.POST("/issue", certificateHandler.Issue)
			certificates.POST("/batch-csv", certificateHandler.BatchIssueCSV)
			certificates.POST("/batch-anchor", batchHandler.AnchorBatch)
			certificates.GET("/bulk-download", handlers.BulkDownloadCertificatesHandler(database.DB))
			certificates.GET("/:certID", certificateHandler.Get)
			certificates.GET("/:certID/download", certificateHandler.DownloadPDF)
		}

		// Internal utility — university admins only
		uni.POST("/test-email", func(c *gin.Context) {
			emailService := services.NewEmailService()
			if err := emailService.TestEmailConnection(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Email not working: " + err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "✅ Email service is working!"})
		})
	}

	// ── Company ───────────────────────────────────────────────────────────────
	company := api.Group("/company")
	{
		company.POST("/register", handlers.RegisterCompany)

		cp := company.Group("")
		cp.Use(middleware.CompanyAuthMiddleware())
		{
			cp.GET("/me", handlers.GetCurrentCompany)
			cp.POST("/search", handlers.SearchStudents)
			cp.POST("/request-profile", handlers.RequestProfileAccess)
			cp.GET("/my-requests", handlers.GetMyRequests)
			cp.GET("/accepted-profiles", handlers.GetAcceptedProfiles)
			cp.POST("/refresh-token", handlers.RefreshCompanyTokenHandler)
			cp.POST("/logout", handlers.LogoutHandler(database.DB))
		}
	}

	// ── Student ───────────────────────────────────────────────────────────────
	student := api.Group("/student")
	{
		student.POST("/register", handlers.RegisterStudent)

		sp := student.Group("")
		sp.Use(middleware.StudentAuthMiddleware())
		{
			// Profile
			sp.GET("/me", handlers.GetCurrentStudent)
			sp.PUT("/profile", handlers.UpdateStudentProfile)
			sp.PUT("/privacy", handlers.UpdatePrivacySettings)
			sp.GET("/profile-settings", handlers.GetMyProfileSettings)

			// Profile-access requests (student responding to company requests)
			sp.GET("/profile-requests", handlers.GetMyProfileRequests)
			sp.POST("/profile-requests/:id/respond", handlers.RespondToRequest)

			// Skills
			sp.POST("/skills", handlers.AddSkill)
			sp.DELETE("/skills/:id", handlers.DeleteSkill)

			// Education
			sp.POST("/education", handlers.AddEducation)
			sp.DELETE("/education/:id", handlers.DeleteEducation)

			// Certificate
			sp.POST("/upload-certificate", handlers.UploadMyCertificate)

			// Token management
			sp.POST("/refresh-token", handlers.RefreshStudentTokenHandler)
			sp.POST("/logout", handlers.LogoutHandler(database.DB))
		}
	}

	// ── Verification (fully public — no auth required) ────────────────────────
	verify := api.Group("/verify")
	{
		verify.POST("", verificationHandler.Verify)
		verify.POST("/pdf", verificationHandler.VerifyPDF)
		verify.GET("/blockchain/:certID", verificationHandler.GetBlockchainInfo)
		verify.GET("/certificate/:id", verificationHandler.VerifyByID)
	}

	// ── Health check ──────────────────────────────────────────────────────────
	router.GET("/health", func(c *gin.Context) {
		sqlDB, err := database.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "degraded",
				"database": "unavailable",
				"ipfs":     ipfsURL,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"database": "connected",
			"ipfs":     ipfsURL,
		})
	})

	// ── Swagger ───────────────────────────────────────────────────────────────
	router.StaticFile("/api-docs/swagger.json", "./docs/swagger.json")
	router.GET("/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.URL("/api-docs/swagger.json")))

	// -------------------------------------------------------------------------
	// Start with graceful shutdown
	// -------------------------------------------------------------------------
	port := getEnvOrDefault("PORT", "8080")

	printStartupBanner(port, ipfsURL)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	// Block until OS signals shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down gracefully (30 s timeout)…")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Forced shutdown: %v", err)
	}
	log.Println("✅ Server stopped cleanly")
}

// =============================================================================
// Helpers
// =============================================================================

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getAllowedOrigins reads CORS_ORIGINS (comma-separated) from the environment.
// Falls back to localhost dev origins when the variable is absent.
func getAllowedOrigins() []string {
	raw := os.Getenv("CORS_ORIGINS")
	if raw == "" {
		return []string{
			"http://localhost:5173",
			"http://localhost:3000",
		}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func printStartupBanner(port, ipfsURL string) {
	sep := "═══════════════════════════════════════════════════"
	log.Println(sep)
	log.Println("🎓 Certificate System — Privacy-First Job Matching")
	log.Println(sep)
	log.Printf("🚀 Server  : http://localhost:%s", port)
	log.Printf("📊 Database: Connected")
	log.Printf("🌐 IPFS    : %s", ipfsURL)
	log.Printf("📧 Email   : %s", func() string {
		if os.Getenv("SMTP_HOST") != "" {
			return "Configured ✅"
		}
		return "Not configured ⚠️"
	}())
	log.Printf("🔐 JWT     : %s", func() string {
		if os.Getenv("JWT_SECRET") != "" {
			return "Custom secret ✅"
		}
		return "Default secret ⚠️  (set JWT_SECRET in production)"
	}())
	log.Printf("📖 Docs    : http://localhost:%s/swagger/index.html", port)
	log.Println(sep)
	log.Println("📡 Login   : POST /api/auth/login  (university | company | student)")
	log.Println(sep)
}
