package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"proxy-go/config"
	"proxy-go/db"
	"proxy-go/handlers"
	"proxy-go/middleware"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Init DB
	db.Init(cfg.MongoURI, cfg.MongoDBName)

	// 3. Setup Router
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
    
    // Set Gin Mode
    // gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// 4. Middlewares
	r.Use(middleware.AuthMiddleware(cfg))
	r.Use(middleware.LoggerMiddleware())

	// 5. Handlers
	azureHandler := handlers.NewAzureHandler(cfg)
	bedrockHandler := handlers.NewBedrockHandler(cfg)

	// Azure Routes
    // Wildcard route to handle all Azure OpenAI paths
	r.Any("/azure/*path", azureHandler.Proxy)

	// Bedrock Routes
    // Specific routes for actions
	r.POST("/bedrock/:action/:modelId", bedrockHandler.Proxy)

	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
