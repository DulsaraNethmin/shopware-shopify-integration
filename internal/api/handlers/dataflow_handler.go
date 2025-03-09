package handlers

import (
	"errors"
	"github.com/goccy/go-json"
	"net/http"
	"strconv"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/models"
	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DataflowHandler handles dataflow API requests
type DataflowHandler struct {
	dataflowService     *services.DataflowService
	fieldMappingService *services.FieldMappingService
}

// NewDataflowHandler creates a new dataflow handler
func NewDataflowHandler(
	dataflowService *services.DataflowService,
	fieldMappingService *services.FieldMappingService,
) *DataflowHandler {
	return &DataflowHandler{
		dataflowService:     dataflowService,
		fieldMappingService: fieldMappingService,
	}
}

// DataflowResponse represents a dataflow response
type DataflowResponse struct {
	ID                uint                  `json:"id"`
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	Type              models.DataflowType   `json:"type"`
	Status            models.DataflowStatus `json:"status"`
	SourceConnectorID uint                  `json:"source_connector_id"`
	DestConnectorID   uint                  `json:"dest_connector_id"`
	SourceConnector   ConnectorResponse     `json:"source_connector"`
	DestConnector     ConnectorResponse     `json:"dest_connector"`
	CreatedAt         string                `json:"created_at"`
	UpdatedAt         string                `json:"updated_at"`
}

// toResponse converts a dataflow model to a response
func toDataflowResponse(dataflow *models.Dataflow) DataflowResponse {
	return DataflowResponse{
		ID:                dataflow.ID,
		Name:              dataflow.Name,
		Description:       dataflow.Description,
		Type:              dataflow.Type,
		Status:            dataflow.Status,
		SourceConnectorID: dataflow.SourceConnectorID,
		DestConnectorID:   dataflow.DestConnectorID,
		SourceConnector:   toConnectorResponse(&dataflow.SourceConnector),
		DestConnector:     toConnectorResponse(&dataflow.DestConnector),
		CreatedAt:         dataflow.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         dataflow.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// FieldMappingResponse represents a field mapping response
type FieldMappingResponse struct {
	ID              uint                      `json:"id"`
	DataflowID      uint                      `json:"dataflow_id"`
	SourceField     string                    `json:"source_field"`
	DestField       string                    `json:"dest_field"`
	IsRequired      bool                      `json:"is_required"`
	DefaultValue    string                    `json:"default_value"`
	TransformType   models.TransformationType `json:"transform_type"`
	TransformConfig string                    `json:"transform_config"`
	CreatedAt       string                    `json:"created_at"`
	UpdatedAt       string                    `json:"updated_at"`
}

// toFieldMappingResponse converts a field mapping model to a response
func toFieldMappingResponse(fieldMapping *models.FieldMapping) FieldMappingResponse {
	return FieldMappingResponse{
		ID:              fieldMapping.ID,
		DataflowID:      fieldMapping.DataflowID,
		SourceField:     fieldMapping.SourceField,
		DestField:       fieldMapping.DestField,
		IsRequired:      fieldMapping.IsRequired,
		DefaultValue:    fieldMapping.DefaultValue,
		TransformType:   fieldMapping.TransformType,
		TransformConfig: fieldMapping.TransformConfig,
		CreatedAt:       fieldMapping.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       fieldMapping.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// MigrationLogResponse represents a migration log response
type MigrationLogResponse struct {
	ID               uint                   `json:"id"`
	DataflowID       uint                   `json:"dataflow_id"`
	Status           models.MigrationStatus `json:"status"`
	SourceIdentifier string                 `json:"source_identifier"`
	DestIdentifier   string                 `json:"dest_identifier"`
	ExecutionARN     string                 `json:"execution_arn"`
	ErrorMessage     string                 `json:"error_message"`
	CompletedAt      string                 `json:"completed_at,omitempty"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

// toMigrationLogResponse converts a migration log model to a response
func toMigrationLogResponse(log *models.MigrationLog) MigrationLogResponse {
	response := MigrationLogResponse{
		ID:               log.ID,
		DataflowID:       log.DataflowID,
		Status:           log.Status,
		SourceIdentifier: log.SourceIdentifier,
		DestIdentifier:   log.DestIdentifier,
		ExecutionARN:     log.ExecutionARN,
		ErrorMessage:     log.ErrorMessage,
		CreatedAt:        log.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        log.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if log.CompletedAt != nil {
		response.CompletedAt = log.CompletedAt.Format("2006-01-02T15:04:05Z")
	}

	return response
}

// CreateDataflow creates a new dataflow
func (h *DataflowHandler) CreateDataflow(c *gin.Context) {
	var dataflow models.Dataflow

	if err := c.ShouldBindJSON(&dataflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if err := h.dataflowService.CreateDataflow(&dataflow); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get the full dataflow with connector details
	fullDataflow, err := h.dataflowService.GetDataflow(dataflow.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Dataflow created successfully",
		"data":    toDataflowResponse(fullDataflow),
	})
}

// GetDataflow gets a dataflow by ID
func (h *DataflowHandler) GetDataflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid dataflow ID",
		})
		return
	}

	dataflow, err := h.dataflowService.GetDataflow(uint(id))
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
		"data": toDataflowResponse(dataflow),
	})
}

// ListDataflows lists all dataflows
func (h *DataflowHandler) ListDataflows(c *gin.Context) {
	typeParam := c.Query("type")
	statusParam := c.Query("status")

	var dataflowType *models.DataflowType
	if typeParam != "" {
		t := models.DataflowType(typeParam)
		dataflowType = &t
	}

	var status *models.DataflowStatus
	if statusParam != "" {
		s := models.DataflowStatus(statusParam)
		status = &s
	}

	dataflows, err := h.dataflowService.ListDataflows(dataflowType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var response []DataflowResponse
	for _, dataflow := range dataflows {
		response = append(response, toDataflowResponse(&dataflow))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// UpdateDataflow updates a dataflow
func (h *DataflowHandler) UpdateDataflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid dataflow ID",
		})
		return
	}

	var dataflow models.Dataflow
	if err := c.ShouldBindJSON(&dataflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if err := h.dataflowService.UpdateDataflow(uint(id), &dataflow); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get the full dataflow with connector details
	fullDataflow, err := h.dataflowService.GetDataflow(dataflow.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Dataflow updated successfully",
		"data":    toDataflowResponse(fullDataflow),
	})
}

// DeleteDataflow deletes a dataflow
func (h *DataflowHandler) DeleteDataflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid dataflow ID",
		})
		return
	}

	if err := h.dataflowService.DeleteDataflow(uint(id)); err != nil {
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
		"message": "Dataflow deleted successfully",
	})
}

// ListFieldMappings lists all field mappings for a dataflow
func (h *DataflowHandler) ListFieldMappings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid dataflow ID",
		})
		return
	}

	// Verify the dataflow exists
	if _, err := h.dataflowService.GetDataflow(uint(id)); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	fieldMappings, err := h.fieldMappingService.ListFieldMappings(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var response []FieldMappingResponse
	for _, mapping := range fieldMappings {
		response = append(response, toFieldMappingResponse(&mapping))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// CreateFieldMapping creates a new field mapping for a dataflow
func (h *DataflowHandler) CreateFieldMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid dataflow ID",
		})
		return
	}

	// Verify the dataflow exists
	if _, err := h.dataflowService.GetDataflow(uint(id)); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	var fieldMapping models.FieldMapping
	if err := c.ShouldBindJSON(&fieldMapping); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Set the dataflow ID
	fieldMapping.DataflowID = uint(id)

	if err := h.fieldMappingService.CreateFieldMapping(&fieldMapping); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Field mapping created successfully",
		"data":    toFieldMappingResponse(&fieldMapping),
	})
}

// UpdateFieldMapping updates a field mapping
func (h *DataflowHandler) UpdateFieldMapping(c *gin.Context) {
	dataflowID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid dataflow ID",
		})
		return
	}

	mappingID, err := strconv.ParseUint(c.Param("mappingId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid mapping ID",
		})
		return
	}

	// Verify the dataflow exists
	if _, err := h.dataflowService.GetDataflow(uint(dataflowID)); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	var fieldMapping models.FieldMapping
	if err := c.ShouldBindJSON(&fieldMapping); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Set the dataflow ID
	fieldMapping.DataflowID = uint(dataflowID)

	if err := h.fieldMappingService.UpdateFieldMapping(uint(mappingID), &fieldMapping); err != nil {
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
		"message": "Field mapping updated successfully",
		"data":    toFieldMappingResponse(&fieldMapping),
	})
}

// DeleteFieldMapping deletes a field mapping
func (h *DataflowHandler) DeleteFieldMapping(c *gin.Context) {
	mappingID, err := strconv.ParseUint(c.Param("mappingId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid mapping ID",
		})
		return
	}

	if err := h.fieldMappingService.DeleteFieldMapping(uint(mappingID)); err != nil {
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
		"message": "Field mapping deleted successfully",
	})
}

// ListMigrationLogs lists migration logs for a dataflow
func (h *DataflowHandler) ListMigrationLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid dataflow ID",
		})
		return
	}

	// Verify the dataflow exists
	if _, err := h.dataflowService.GetDataflow(uint(id)); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse query parameters
	statusParam := c.Query("status")
	limitParam := c.DefaultQuery("limit", "20")
	offsetParam := c.DefaultQuery("offset", "0")

	var migrationStatus *models.MigrationStatus
	if statusParam != "" {
		status := models.MigrationStatus(statusParam)
		migrationStatus = &status
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetParam)
	if err != nil || offset < 0 {
		offset = 0
	}

	logs, err := h.dataflowService.GetMigrationLogs(uint(id), migrationStatus, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var response []MigrationLogResponse
	for _, log := range logs {
		response = append(response, toMigrationLogResponse(&log))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// GetMigrationLog gets a migration log by ID
func (h *DataflowHandler) GetMigrationLog(c *gin.Context) {
	dataflowID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid dataflow ID",
		})
		return
	}

	logID, err := strconv.ParseUint(c.Param("logId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid log ID",
		})
		return
	}

	// Verify the dataflow exists
	if _, err := h.dataflowService.GetDataflow(uint(dataflowID)); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	log, err := h.dataflowService.GetMigrationLog(uint(logID))
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

	// Check if the log belongs to the requested dataflow
	if log.DataflowID != uint(dataflowID) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Migration log not found for this dataflow",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": toMigrationLogResponse(log),
	})
}

// ExecuteDataflow manually executes a dataflow with the provided data
func (h *DataflowHandler) ExecuteDataflow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid dataflow ID",
		})
		return
	}

	var request struct {
		SourceIdentifier string          `json:"source_identifier" binding:"required"`
		SourceData       json.RawMessage `json:"source_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if err := h.dataflowService.ExecuteDataflow(uint(id), request.SourceIdentifier, request.SourceData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Dataflow execution started successfully",
	})
}
