package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"new-api/common"
	"new-api/middleware"
	"new-api/model"
	"new-api/router"
)

func main() {
	// Load environment variables from .env file if present
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Initialize common settings
	common.Init()

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	err = model.InitDB()
	if err != nil {
		common.FatalLog("Failed to initialize database: " + err.Error())
	}
	defer model.CloseDB()

	// Run database migrations
	err = model.MigrateDB()
	if err != nil {
		common.FatalLog("Failed to migrate database: " + err.Error())
	}

	// Initialize Redis if configured
	if os.Getenv("REDIS_CONN_STRING") != "" {
		err = common.InitRedisClient()
		if err != nil {
			// Non-fatal: log the error and continue without Redis
			common.SysLog("Warning: Failed to initialize Redis, continuing without it: " + err.Error())
		}
	}

	// Initialize options from database
	model.InitOptionMap()

	// Create Gin engine with default middleware
	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())
	server.Use(middleware.CORS())

	// Setup all routes
	router.SetRouter(server)

	// Determine port — default changed to 8080 to avoid conflict with other
	// local services I run on 3000 (e.g. Node dev servers).
	// Fallback chain: PORT env var -> FALLBACK_PORT env var -> 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("FALLBACK_PORT")
	}
	if port == "" {
		port = "8080"
	}

	common.SysLog(fmt.Sprintf("Server starting on port %s", port))
	fmt.Printf("[new-api] Listening on http://localhost:%s\n", port)

	// Start the server
	if err := server.Run(":" + port); err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}
