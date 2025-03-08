package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/shopware-shopify-integration/internal/models"
	"github.com/yourusername/shopware-shopify-integration/internal/services"
	"gorm.io/gorm"
)

// WebhookHandler handles webhook requests
type WebhookHandler struct {
	db                   *gorm.DB
	shopwareService      *services.ShopwareService
	stepFunctionsService *services.StepFunctionsService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	db *gorm.DB,
	shopwareService *services.ShopwareService,
	stepFunctionsService *services.StepFunctionsService,
) *WebhookHandler {
	return &WebhookHandler{
		db:                   db,
		shopwareService:      shopwareService,
		stepFunctionsService: stepFunctionsService,
	}
}

// ShopwareWebhookRequest represents a webhook request from Shopware
type ShopwareWebhookRequest struct {
	Source string          `json:"source"` // "product" or "order"
	Event  string          `json:"event"`  // "created", "updated", "deleted"
	Data   json.RawMessage `json:"data"`   // Raw JSON data
}

// HandleShopwareWebhook handles a webhook from Shopware
func (h *WebhookHandler) HandleShopwareWebhook(c *gin.Context) {
	// Read and validate the webhook payload
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error reading request body",
		})
		return
	}

	var webhook ShopwareWebhookRequest
	if err := json.Unmarshal(body, &webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON payload",
		})
		return
	}

	// Determine data type and event type
	var dataflowType models.DataflowType
	switch webhook.Source {
	case "product":
		dataflowType = models.DataflowTypeProduct
	case "order":
		dataflowType = models.DataflowTypeOrder
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported source type",
		})
		return
	}

	// Find active dataflows for this data type
	var dataflows []models.Dataflow
	if err := h.db.Preload("SourceConnector").Preload("DestConnector").
		Where("type = ? AND status = ?", dataflowType, models.DataflowStatusActive).
		Find(&dataflows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error finding dataflows",
		})
		return
	}

	if len(dataflows) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No active dataflows found for this data type",
		})
		return
	}

	// Extract source identifier from data
	var sourceID string
	switch dataflowType {
	case models.DataflowTypeProduct:
		var product struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(webhook.Data, &product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid product data",
			})
			return
		}
		sourceID = product.ID
	case models.DataflowTypeOrder:
		var order struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(webhook.Data, &order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid order data",
			})
			return
		}
		sourceID = order.ID
	}

	// Process each matching dataflow
	for _, dataflow := range dataflows {
		// Create a migration log entry
		migrationLog := models.MigrationLog{
			DataflowID:       dataflow.ID,
			Status:           models.MigrationStatusPending,
			SourceIdentifier: sourceID,
			SourcePayload:    string(webhook.Data),
		}

		if err := h.db.Create(&migrationLog).Error; err != nil {
			// Log the error but continue with other dataflows
			continue
		}

		// Start a Step Functions execution
		executionARN, err := h.stepFunctionsService.StartExecution(dataflow.ID, migrationLog.ID, webhook.Data)
		if err != nil {
			migrationLog.Status = models.MigrationStatusFailed
			migrationLog.ErrorMessage = err.Error()
			h.db.Save(&migrationLog)
			continue
		}

		// Update the migration log with the execution ARN
		migrationLog.Status = models.MigrationStatusInProgress
		migrationLog.ExecutionARN = executionARN
		h.db.Save(&migrationLog)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook processed successfully",
	})
}

// UpdateMigrationStatus updates the status of a migration
func (h *WebhookHandler) UpdateMigrationStatus(c *gin.Context) {
	var request struct {
		MigrationID     uint                   `json:"migration_id"`
		Status          models.MigrationStatus `json:"status"`
		DestIdentifier  string                 `json:"dest_identifier,omitempty"`
		ErrorMessage    string                 `json:"error_message,omitempty"`
		TransformedData json.RawMessage        `json:"transformed_data,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	var migrationLog models.MigrationLog
	if err := h.db.First(&migrationLog, request.MigrationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Migration log not found",
		})
		return
	}

	migrationLog.Status = request.Status
	if request.DestIdentifier != "" {
		migrationLog.DestIdentifier = request.DestIdentifier
	}
	if request.ErrorMessage != "" {
		migrationLog.ErrorMessage = request.ErrorMessage
	}
	if len(request.TransformedData) > 0 {
		migrationLog.TransformedPayload = string(request.TransformedData)
	}

	if request.Status == models.MigrationStatusSuccess || request.Status == models.MigrationStatusFailed {
		now := time.Now()
		migrationLog.CompletedAt = &now
	}

	if err := h.db.Save(&migrationLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error updating migration log",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Migration status updated",
	})
}
