// Package main is the entry point for the Attendance API.
//
//	@title			Restaurant Attendance API
//	@version		1.0
//	@description	Attendance, Employee Management, and Overtime modules
//	@host			localhost:9003
//
// //	@BasePath		/api/v1
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "restaurant-api/docs"
	"restaurant-api/internal/attendance"
	"restaurant-api/internal/employee"
	"restaurant-api/internal/overtime"
)

const (
	defaultPort  = "9003"
	defaultDBURL = "postgres://postgres:Postgres%40123@localhost:5432/restaurant_db?sslmode=disable"
)

// Context keys
const (
	CtxUserID   = "user_id"
	CtxUserRole = "user_role"
	CtxBranchID = "branch_id"
)

func main() {
	// ============================================================
	// 1. Database
	// ============================================================
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = defaultDBURL
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	// ============================================================
	// 2. Employee Module (before attendance — attendance needs it)
	// ============================================================
	employeeRepo := employee.NewRepository(db)
	employeeSvc := employee.NewService(db, employeeRepo)
	employeeHandler := employee.NewHandler(employeeSvc)

	// ============================================================
	// 3. Attendance Module
	// ============================================================
	attendanceRepo := attendance.NewRepository(db)
	attendanceSvc := attendance.NewService(
		db,
		attendanceRepo,
		employeeReaderAdapter{svc: employeeSvc},
		branchReaderAdapter{db: db},
	)
	attendanceHandler := attendance.NewHandler(attendanceSvc)

	// ============================================================
	// 4. Overtime Module
	// ============================================================
	overtimeRepo := overtime.NewRepository(db)
	overtimeSvc := overtime.NewService(db, overtimeRepo)
	overtimeHandler := overtime.NewHandler(overtimeSvc)

	overtimeHandler.SetEmployeeBranchLookup(func(ctx context.Context, employeeID string) (string, error) {
		emp, err := employeeRepo.GetByID(ctx, employeeID)
		if err != nil {
			return "", err
		}
		if emp == nil {
			return "", overtime.ErrEmployeeNotFound
		}
		return emp.BranchID, nil
	})

	// ============================================================
	// 5. Router
	// ============================================================
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(apiGatewayMiddleware())

	// Swagger (public)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health (public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "attendance-api",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API v1 (protected)
	api := router.Group("/api/v1")
	{
		attendanceHandler.RegisterRoutes(api)
		employeeHandler.RegisterRoutes(api)
		overtimeHandler.RegisterRoutes(api)
	}

	// ============================================================
	// 6. Start Server
	// ============================================================
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("🚀 Attendance API listening on :%s", port)
	log.Printf("   📚 Swagger:  http://localhost:%s/swagger/index.html", port)
	log.Printf("   ❤️  Health:   http://localhost:%s/health", port)
	log.Printf("   🔗 BaseURL:  http://localhost:%s/api/v1", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// ============================================================
// apiGatewayMiddleware — reads Gateway headers
// ============================================================

func apiGatewayMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if isPublicPath(path) {
			c.Next()
			return
		}

		userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
		userRole := strings.TrimSpace(c.GetHeader("X-User-Role"))
		branchID := strings.TrimSpace(c.GetHeader("X-Branch-ID"))

		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "missing X-User-ID header",
				},
			})
			return
		}

		if userRole == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "missing X-User-Role header",
				},
			})
			return
		}

		c.Set(CtxUserID, userID)
		c.Set(CtxUserRole, userRole)
		c.Set(CtxBranchID, branchID)

		c.Next()
	}
}

func isPublicPath(path string) bool {
	return path == "/health" || strings.HasPrefix(path, "/swagger/")
}
