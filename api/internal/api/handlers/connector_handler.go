package handlers

import (
	"errors"
	"fmt"
	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/config"
	"net/http"
	"strconv"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/models"
	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConnectorHandler handles connector API requests
type ConnectorHandler struct {
	service *services.ConnectorService
	config  *config.Config
}

// NewConnectorHandler creates a new connector handler
func NewConnectorHandler(service *services.ConnectorService, config *config.Config) *ConnectorHandler { // Updated param
	return &ConnectorHandler{
		service: service,
		config:  config, // Add this
	}
}

// ConnectorResponse represents a connector response
type ConnectorResponse struct {
	ID        uint                 `json:"id"`
	Name      string               `json:"name"`
	Type      models.ConnectorType `json:"type"`
	URL       string               `json:"url"`
	Username  string               `json:"username,omitempty"`
	IsActive  bool                 `json:"is_active"`
	CreatedAt string               `json:"created_at"`
	UpdatedAt string               `json:"updated_at"`
}

//type CreateConnectorRequest struct {
//	Name      string               `json:"name" binding:"required"`
//	Type      models.ConnectorType `json:"type" binding:"required"`
//	URL       string               `json:"url" binding:"required"`
//	ApiKey    string               `json:"api_key"`
//	ApiSecret string               `json:"api_secret"`
//	Username  string               `json:"username"`
//	Password  string               `json:"password"`
//}

// toResponse converts a connector model to a response
func toConnectorResponse(connector *models.Connector) ConnectorResponse {
	return ConnectorResponse{
		ID:        connector.ID,
		Name:      connector.Name,
		Type:      connector.Type,
		URL:       connector.URL,
		Username:  connector.Username,
		IsActive:  connector.IsActive,
		CreatedAt: connector.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: connector.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// CreateConnector creates a new connector
func (h *ConnectorHandler) CreateConnector(c *gin.Context) {
	var connector models.Connector

	if err := c.ShouldBindJSON(&connector); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if err := h.service.CreateConnector(&connector); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// If it's a Shopware connector, register webhooks automatically
	if connector.Type == models.ConnectorTypeShopware {
		// Get the callback URL from configuration
		callbackURL := h.config.Server.CallbackURL + "/api/v1/webhook/shopware"

		// Register webhooks
		if err := h.service.RegisterWebhooks(connector.ID, callbackURL); err != nil {
			// Log the error but don't fail the connector creation
			c.JSON(http.StatusCreated, gin.H{
				"message": "Connector created successfully, but webhook registration failed",
				"data":    toConnectorResponse(&connector),
				"warning": fmt.Sprintf("Failed to register webhooks: %v", err),
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Connector created successfully",
		"data":    toConnectorResponse(&connector),
	})
}

// GetConnector gets a connector by ID
func (h *ConnectorHandler) GetConnector(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid connector ID",
		})
		return
	}

	connector, err := h.service.GetConnector(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": toConnectorResponse(connector),
	})
}

// ListConnectors lists all connectors
func (h *ConnectorHandler) ListConnectors(c *gin.Context) {
	typeParam := c.Query("type")

	var connectorType *models.ConnectorType
	if typeParam != "" {
		t := models.ConnectorType(typeParam)
		connectorType = &t
	}

	connectors, err := h.service.ListConnectors(connectorType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var response []ConnectorResponse
	for _, connector := range connectors {
		response = append(response, toConnectorResponse(&connector))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// UpdateConnector updates a connector
func (h *ConnectorHandler) UpdateConnector(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid connector ID",
		})
		return
	}

	var connector models.Connector
	if err := c.ShouldBindJSON(&connector); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if err := h.service.UpdateConnector(uint(id), &connector); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Connector updated successfully",
		"data":    toConnectorResponse(&connector),
	})
}

// DeleteConnector deletes a connector
func (h *ConnectorHandler) DeleteConnector(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid connector ID",
		})
		return
	}

	if err := h.service.DeleteConnector(uint(id)); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Connector deleted successfully",
	})
}

// TestConnection tests a connector connection
func (h *ConnectorHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid connector ID",
		})
		return
	}

	if err := h.service.TestConnection(uint(id)); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Connection test successful",
	})
}

// RegisterWebhooks registers webhooks for a connector
func (h *ConnectorHandler) RegisterWebhooks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid connector ID",
		})
		return
	}

	var request struct {
		CallbackURL string `json:"callback_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if err := h.service.RegisterWebhooks(uint(id), request.CallbackURL); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhooks registered successfully",
	})
}

// GetWebhooks gets all webhooks for a connector
func (h *ConnectorHandler) GetWebhooks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid connector ID",
		})
		return
	}

	// Get webhooks from the service
	webhooks, err := h.service.GetWebhooks(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": webhooks,
	})
}
