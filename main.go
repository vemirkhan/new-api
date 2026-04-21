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
			common.FatalLog("Failed to initialize Redis: " + err.Error())
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

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	common.SysLog(fmt.Sprintf("Server starting on port %s", port))

	// Start the server
	if err := server.Run(":" + port); err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}
