package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/shopware-shopify-integration/internal/api/handlers"
	"github.com/yourusername/shopware-shopify-integration/internal/api/middleware"
	"github.com/yourusername/shopware-shopify-integration/internal/config"
	"github.com/yourusername/shopware-shopify-integration/internal/services"
	"gorm.io/gorm"
)

// Server is the API server
type Server struct {
	router   *gin.Engine
	config   *config.Config
	database *gorm.DB
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	router := gin.Default()

	server := &Server{
		router:   router,
		config:   cfg,
		database: db,
	}

	server.setupRoutes()

	return server
}

// setupRoutes sets up the API routes
func (s *Server) setupRoutes() {
	// Create services
	connectorService := services.NewConnectorService(s.database)
	dataflowService := services.NewDataflowService(s.database)
	fieldMappingService := services.NewFieldMappingService(s.database)
	shopwareService := services.NewShopwareService(s.database)
	shopifyService := services.NewShopifyService(s.database)
	stepFunctionsService := services.NewStepFunctionsService(s.config.AWS, s.database)

	// Create handlers
	connectorHandler := handlers.NewConnectorHandler(connectorService)
	dataflowHandler := handlers.NewDataflowHandler(dataflowService, fieldMappingService)
	webhookHandler := handlers.NewWebhookHandler(s.database, shopwareService, stepFunctionsService)

	// Public routes (no authentication required)
	publicGroup := s.router.Group("/api/v1")
	{
		// Health check
		publicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Webhook endpoint for Shopware
		publicGroup.POST("/webhook/shopware", webhookHandler.HandleShopwareWebhook)
	}

	// Private routes (authentication required)
	privateGroup := s.router.Group("/api/v1")
	privateGroup.Use(middleware.AuthMiddleware(s.config.Server.Secret))
	{
		// Connector routes
		privateGroup.GET("/connectors", connectorHandler.ListConnectors)
		privateGroup.POST("/connectors", connectorHandler.CreateConnector)
		privateGroup.GET("/connectors/:id", connectorHandler.GetConnector)
		privateGroup.PUT("/connectors/:id", connectorHandler.UpdateConnector)
		privateGroup.DELETE("/connectors/:id", connectorHandler.DeleteConnector)

		// Dataflow routes
		privateGroup.GET("/dataflows", dataflowHandler.ListDataflows)
		privateGroup.POST("/dataflows", dataflowHandler.CreateDataflow)
		privateGroup.GET("/dataflows/:id", dataflowHandler.GetDataflow)
		privateGroup.PUT("/dataflows/:id", dataflowHandler.UpdateDataflow)
		privateGroup.DELETE("/dataflows/:id", dataflowHandler.DeleteDataflow)

		// Field mapping routes
		privateGroup.GET("/dataflows/:id/mappings", dataflowHandler.ListFieldMappings)
		privateGroup.POST("/dataflows/:id/mappings", dataflowHandler.CreateFieldMapping)
		privateGroup.PUT("/dataflows/:id/mappings/:mappingId", dataflowHandler.UpdateFieldMapping)
		privateGroup.DELETE("/dataflows/:id/mappings/:mappingId", dataflowHandler.DeleteFieldMapping)

		// Migration log routes
		privateGroup.GET("/dataflows/:id/logs", dataflowHandler.ListMigrationLogs)
		privateGroup.GET("/dataflows/:id/logs/:logId", dataflowHandler.GetMigrationLog)
	}
}

// Run starts the API server
func (s *Server) Run() error {
	return s.router.Run(fmt.Sprintf(":%d", s.config.Server.Port))
}
