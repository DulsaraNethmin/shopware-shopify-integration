package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/shopware-shopify-integration/internal/models"
	"github.com/yourusername/shopware-shopify-integration/internal/services"
	"gorm.io/gorm"
)

// ConnectorHandler handles connector API requests
type ConnectorHandler struct {
	service *services.ConnectorService
}

// NewConnectorHandler creates a new connector handler
func NewConnectorHandler(service *services.ConnectorService) *ConnectorHandler {
	return &ConnectorHandler{
		service: service,
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
