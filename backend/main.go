// Package main is the entry point for the Flickey Go backend.
// It wires all packages, configures Gin, and starts the HTTP server.
//
//	@title			Flickey Go API
//	@version		1.0
//	@description	Property-rental platform backend. Provides OTP-based authentication, multi-step listing wizard, media upload workflow, amenity catalogue, and an admin moderation panel.
//	@termsOfService	http://flickey.com/terms/
//
//	@contact.name	Flickey Dev Team
//	@contact.email	dev@flickey.com
//
//	@license.name	MIT
//
//	@host		localhost:8000
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT access token. Format: **Bearer {token}**
//
//	@securityDefinitions.apikey	CookieAuth
//	@in							cookie
//	@name						refresh_token
//	@description				HttpOnly refresh token cookie set by /auth/verify-otp and rotated by /auth/refresh.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flickey/go-backend/admin"
	"flickey/go-backend/amenities"
	"flickey/go-backend/auth"
	"flickey/go-backend/config"
	"flickey/go-backend/db"
	_ "flickey/go-backend/docs"
	"flickey/go-backend/geo"
	"flickey/go-backend/listings"
	"flickey/go-backend/media"
	"flickey/go-backend/notifications"
	"flickey/go-backend/sms"
	"flickey/go-backend/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func main() {
	// ── Logger ────────────────────────────────────────────────────────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.IsDev() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// ── Database ──────────────────────────────────────────────────────────────
	database, err := db.Open(cfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	sqlDB, err := database.DB()
	if err != nil {
		logger.Error("failed to get sql.DB", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()
	logger.Info("connected to PostgreSQL")

	// Ensure default persistent admin account exists
	ensureDefaultAdmin(database, logger)

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb, err := db.NewRedis(cfg)
	if err != nil {
		logger.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	logger.Info("connected to Redis")

	// ── Storage ───────────────────────────────────────────────────────────────
	var store storage.StorageProvider
	if cfg.S3AccessKey != "" && cfg.S3SecretKey != "" {
		s3Store, err := storage.NewS3Storage(cfg)
		if err != nil {
			logger.Error("failed to initialize S3 storage", "error", err)
			os.Exit(1)
		}
		store = s3Store
		logger.Info("using S3 storage", "endpoint", cfg.S3EndpointURL, "bucket", cfg.S3Bucket)
	} else {
		baseURL := "http://localhost:" + cfg.Port + "/api/v1/media/dev-upload/"
		localStorage, err := storage.NewLocalStorage("./local-storage", baseURL)
		if err != nil {
			logger.Error("failed to initialize local storage", "error", err)
			os.Exit(1)
		}
		store = localStorage
		logger.Info("using local file storage (dev mode)", "upload_url", baseURL)
	}

	// ── SMS Sender ────────────────────────────────────────────────────────────
	var smsSender sms.SMSSender = sms.NewConsoleSMSSender(logger)
	logger.Info("SMS sender: console")

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := &auth.AuthService{
		DB:        database,
		Redis:     rdb,
		Cfg:       cfg,
		SMSSender: smsSender,
		Logger:    logger,
	}

	mediaSvc := media.NewMediaService(database, store, cfg)
	listingSvc := listings.NewListingService(database, rdb, cfg)
	notifSvc := notifications.NewNotificationService(database, rdb, logger)
	geoSvc := geo.NewGeoService(rdb, cfg.GeocoderURL, logger)

	// ── Middleware ────────────────────────────────────────────────────────────
	authMiddleware := auth.RequireAuth(cfg, database)
	activeMiddleware := auth.RequireActiveUser()

	// ── Gin Router ────────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggerMiddleware(logger))
	r.Use(corsMiddleware(cfg))

	// ── Routes ────────────────────────────────────────────────────────────────
	v1 := r.Group("/api/v1")

	// Auth
	authHandler := auth.NewHandler(authSvc, cfg)
	authGroup := v1.Group("/auth")
	authHandler.RegisterRoutes(authGroup, authMiddleware, activeMiddleware)

	// Listings
	listingHandler := listings.NewHandler(listingSvc)
	listingGroup := v1.Group("/listings")
	listingHandler.RegisterRoutes(listingGroup, authMiddleware, activeMiddleware)

	// Media
	mediaHandler := media.NewHandler(mediaSvc)
	mediaGroup := v1.Group("/media")
	mediaHandler.RegisterRoutes(mediaGroup, authMiddleware, activeMiddleware)

	// Amenities
	amenitiesGroup := v1.Group("/amenities")
	amenities.RegisterRoutes(amenitiesGroup)

	// Geo (public geocoding and address autocomplete)
	geoHandler := geo.NewHandler(geoSvc)
	geoGroup := v1.Group("/geo")
	geoHandler.RegisterRoutes(geoGroup)

	// Notifications
	notifHandler := notifications.NewHandler(notifSvc, cfg, database)
	notifGroup := v1.Group("/notifications")
	notifHandler.RegisterRoutes(notifGroup, authMiddleware, activeMiddleware)

	// Admin
	adminSvc := admin.NewAdminService(database, notifSvc)
	adminHandler := admin.NewHandler(adminSvc, cfg)
	adminGroup := v1.Group("/admin")
	adminHandler.RegisterRoutes(adminGroup, authMiddleware, activeMiddleware)

	// Health check.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "flickey-go"})
	})

	// Swagger UI — dev only.
	if cfg.IsDev() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		logger.Info("Swagger UI enabled", "url", "http://localhost:"+cfg.Port+"/swagger/index.html")
	}

	// ── HTTP Server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server in background.
	go func() {
		logger.Info("starting Flickey Go API", "addr", srv.Addr, "env", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("server stopped")
}

// ─────────────────────────────────────────────────────────────────────────────
// Middleware
// ─────────────────────────────────────────────────────────────────────────────

func loggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http_request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

// corsMiddleware applies CORS headers matching the frontend configuration.
func corsMiddleware(cfg *config.Settings) gin.HandlerFunc {
	allowedOrigins := map[string]bool{
		"http://localhost:5173": true,
		"http://127.0.0.1:5173": true,
		"http://127.0.0.1:5500": true,
		"http://localhost:5500": true,
		"http://127.0.0.1:8000": true,
		"http://localhost:8000": true,
		"http://localhost:3000": true,
		"http://127.0.0.1:3000": true,
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Idempotency-Key, X-Forwarded-For")
			c.Header("Access-Control-Expose-Headers", "Content-Length")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func ensureDefaultAdmin(database *gorm.DB, logger *slog.Logger) {
	ctx := context.Background()
	phone := "+375290000000"
	var existing db.User
	err := database.WithContext(ctx).Where("phone = ?", phone).First(&existing).Error
	if err != nil {
		firstName := "Admin"
		lastName := "Flickey"
		adminUser := db.User{
			ID:        uuid.New(),
			Phone:     phone,
			FirstName: &firstName,
			LastName:  &lastName,
			Role:      db.UserRoleAdmin,
			Status:    db.UserStatusActive,
		}
		if createErr := database.WithContext(ctx).Create(&adminUser).Error; createErr == nil {
			logger.Info("seeded default admin user", "phone", phone, "id", adminUser.ID)
		}
	} else if existing.Role != db.UserRoleAdmin {
		database.WithContext(ctx).Model(&existing).Update("role", db.UserRoleAdmin)
	}
}
