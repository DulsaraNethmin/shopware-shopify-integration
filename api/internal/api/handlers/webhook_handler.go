package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/models"
	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/services"
	"github.com/gin-gonic/gin"
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
	Data struct {
		Payload []struct {
			Entity        string   `json:"entity"`
			Operation     string   `json:"operation"`
			PrimaryKey    string   `json:"primaryKey"`
			UpdatedFields []string `json:"updatedFields"`
			VersionId     string   `json:"versionId"`
		} `json:"payload"`
		Event string `json:"event"`
	} `json:"data"`
	Source struct {
		URL     string `json:"url"`
		EventID string `json:"eventId"`
	} `json:"source"`
	Timestamp int64 `json:"timestamp"`
}

func (h *WebhookHandler) HandleShopwareWebhook(c *gin.Context) {

	println("nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn")
	// Read and validate the webhook payload
	body, err := io.ReadAll(c.Request.Body)
	//print(string(body))
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

	fmt.Print(webhook)

	// Check if there's a valid payload
	if len(webhook.Data.Payload) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No payload in webhook",
		})
		return
	}

	// Determine data type and event type
	var dataflowType models.DataflowType
	if webhook.Data.Event == "product.written" {
		dataflowType = models.DataflowTypeProduct
	} else if webhook.Data.Event == "order.placed" {
		dataflowType = models.DataflowTypeOrder
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported event type: " + webhook.Data.Event,
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

	println("dataflowssssssssssssssssssssssssssssssssssssssssssssssss")
	print(dataflows)

	if len(dataflows) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No active dataflows found for this data type",
		})
		return
	}

	// Extract source identifier from data
	sourceID := ""
	for _, payload := range webhook.Data.Payload {
		if dataflowType == models.DataflowTypeProduct && payload.Entity == "product" {
			sourceID = payload.PrimaryKey
		} else if dataflowType == models.DataflowTypeOrder && payload.Entity == "order" {
			sourceID = payload.PrimaryKey
		}
	}

	if sourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not determine source identifier",
		})
		return
	}

	// For products, we need to fetch the full product data
	var sourceData []byte
	if dataflowType == models.DataflowTypeProduct {
		// Find the Shopware connector that has the URL matching the source URL
		var connector models.Connector
		domain := strings.TrimPrefix(webhook.Source.URL, "https://")
		if err := h.db.Where("type = ? AND url LIKE ?", models.ConnectorTypeShopware, "%"+domain+"%").First(&connector).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Could not find matching connector for the source URL",
			})
			return
		}

		// Get the full product data
		product, err := h.shopwareService.GetProduct(&connector, sourceID)
		println("Proooooooooooooooooooooooduct")
		println(*&product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get product data: " + err.Error(),
			})
			return
		}

		sourceData, err = json.Marshal(product)

		println(string(sourceData))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to marshal product data",
			})
			return
		}
	} else {
		// For orders, just pass the webhook payload as is
		sourceData = body
	}

	// Process each matching dataflow
	for _, dataflow := range dataflows {
		// Create a migration log entry
		migrationLog := models.MigrationLog{
			DataflowID:       dataflow.ID,
			Status:           models.MigrationStatusPending,
			SourceIdentifier: sourceID,
			SourcePayload:    string(sourceData),
		}

		if err := h.db.Create(&migrationLog).Error; err != nil {
			// Log the error but continue with other dataflows
			continue
		}

		// Start a Step Functions execution
		executionARN, err := h.stepFunctionsService.StartExecution(dataflow.ID, migrationLog.ID, sourceData)
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
