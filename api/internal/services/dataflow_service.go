package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/models"
	"gorm.io/gorm"
)

// DataflowService handles dataflow operations
type DataflowService struct {
	db *gorm.DB
}

// NewDataflowService creates a new dataflow service
func NewDataflowService(db *gorm.DB) *DataflowService {
	return &DataflowService{
		db: db,
	}
}

// CreateDataflow creates a new dataflow
func (s *DataflowService) CreateDataflow(dataflow *models.Dataflow) error {
	return s.db.Create(dataflow).Error
}

// GetDataflow gets a dataflow by ID
func (s *DataflowService) GetDataflow(id uint) (*models.Dataflow, error) {
	var dataflow models.Dataflow

	if err := s.db.Preload("SourceConnector").Preload("DestConnector").First(&dataflow, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &dataflow, nil
}

// ListDataflows lists all dataflows
func (s *DataflowService) ListDataflows(dataflowType *models.DataflowType, status *models.DataflowStatus) ([]models.Dataflow, error) {
	var dataflows []models.Dataflow

	query := s.db.Preload("SourceConnector").Preload("DestConnector")

	if dataflowType != nil {
		query = query.Where("type = ?", *dataflowType)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Find(&dataflows).Error; err != nil {
		return nil, err
	}

	return dataflows, nil
}

// UpdateDataflow updates a dataflow
func (s *DataflowService) UpdateDataflow(id uint, dataflow *models.Dataflow) error {
	// Check if the dataflow exists
	existingDataflow, err := s.GetDataflow(id)
	if err != nil {
		return err
	}

	// Update the dataflow
	dataflow.ID = existingDataflow.ID
	return s.db.Save(dataflow).Error
}

// DeleteDataflow deletes a dataflow
func (s *DataflowService) DeleteDataflow(id uint) error {
	// Check if the dataflow exists
	existingDataflow, err := s.GetDataflow(id)
	if err != nil {
		return err
	}

	// Check if there are any migration logs for this dataflow
	var count int64
	if err := s.db.Model(&models.MigrationLog{}).Where("dataflow_id = ?", id).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("dataflow has migration logs and cannot be deleted")
	}

	// Delete field mappings first
	if err := s.db.Where("dataflow_id = ?", id).Delete(&models.FieldMapping{}).Error; err != nil {
		return err
	}

	// Delete the dataflow
	return s.db.Delete(existingDataflow).Error
}

// GetMigrationLogs gets migration logs for a dataflow
func (s *DataflowService) GetMigrationLogs(dataflowID uint, status *models.MigrationStatus, limit, offset int) ([]models.MigrationLog, error) {
	var logs []models.MigrationLog

	query := s.db.Where("dataflow_id = ?", dataflowID)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// GetMigrationLog gets a migration log by ID
func (s *DataflowService) GetMigrationLog(id uint) (*models.MigrationLog, error) {
	var log models.MigrationLog

	if err := s.db.First(&log, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &log, nil
}

// ExecuteDataflow executes a dataflow for the given source data
func (s *DataflowService) ExecuteDataflow(dataflowID uint, sourceIdentifier string, sourceData []byte) error {
	// Get the dataflow
	dataflow, err := s.GetDataflow(dataflowID)
	if err != nil {
		return err
	}

	// Create a migration log
	migrationLog := models.MigrationLog{
		DataflowID:       dataflow.ID,
		SourceIdentifier: sourceIdentifier,
		SourcePayload:    string(sourceData),
		Status:           models.MigrationStatusInProgress,
	}

	if err := s.db.Create(&migrationLog).Error; err != nil {
		return err
	}

	// Execute the dataflow
	// This would normally be handled by the Step Functions workflow
	// For testing purposes, we'll implement a basic flow here

	// 1. Transform the data
	fieldMappingService := NewFieldMappingService(s.db)
	result, err := fieldMappingService.TransformData(dataflow.ID, sourceData)
	if err != nil {
		migrationLog.Status = models.MigrationStatusFailed
		migrationLog.ErrorMessage = fmt.Sprintf("Error transforming data: %v", err)
		s.db.Save(&migrationLog)
		return err
	}

	if result.Error != nil {
		migrationLog.Status = models.MigrationStatusFailed
		migrationLog.ErrorMessage = fmt.Sprintf("Error in transformation: %v", result.Error)
		s.db.Save(&migrationLog)
		return result.Error
	}

	// 2. Upload to Shopify
	shopifyService := NewShopifyService(s.db)

	switch dataflow.Type {
	case models.DataflowTypeProduct:
		// Create a Shopify product
		transformedJSON, err := json.Marshal(result.Data)
		if err != nil {
			migrationLog.Status = models.MigrationStatusFailed
			migrationLog.ErrorMessage = fmt.Sprintf("Error marshaling transformed data: %v", err)
			s.db.Save(&migrationLog)
			return err
		}

		migrationLog.TransformedPayload = string(transformedJSON)

		var productRequest ProductCreateRequest
		if err := json.Unmarshal(transformedJSON, &productRequest); err != nil {
			migrationLog.Status = models.MigrationStatusFailed
			migrationLog.ErrorMessage = fmt.Sprintf("Error unmarshaling transformed data: %v", err)
			s.db.Save(&migrationLog)
			return err
		}

		response, err := shopifyService.CreateProduct(&dataflow.DestConnector, &productRequest)
		if err != nil {
			migrationLog.Status = models.MigrationStatusFailed
			migrationLog.ErrorMessage = fmt.Sprintf("Error creating product in Shopify: %v", err)
			s.db.Save(&migrationLog)
			return err
		}

		migrationLog.DestIdentifier = fmt.Sprintf("%d", response.Product.ID)

	case models.DataflowTypeOrder:
		// Create a Shopify order
		transformedJSON, err := json.Marshal(result.Data)
		if err != nil {
			migrationLog.Status = models.MigrationStatusFailed
			migrationLog.ErrorMessage = fmt.Sprintf("Error marshaling transformed data: %v", err)
			s.db.Save(&migrationLog)
			return err
		}

		migrationLog.TransformedPayload = string(transformedJSON)

		var orderRequest OrderCreateRequest
		if err := json.Unmarshal(transformedJSON, &orderRequest); err != nil {
			migrationLog.Status = models.MigrationStatusFailed
			migrationLog.ErrorMessage = fmt.Sprintf("Error unmarshaling transformed data: %v", err)
			s.db.Save(&migrationLog)
			return err
		}

		response, err := shopifyService.CreateOrder(&dataflow.DestConnector, &orderRequest)
		if err != nil {
			migrationLog.Status = models.MigrationStatusFailed
			migrationLog.ErrorMessage = fmt.Sprintf("Error creating order in Shopify: %v", err)
			s.db.Save(&migrationLog)
			return err
		}

		migrationLog.DestIdentifier = fmt.Sprintf("%d", response.Order.ID)

	default:
		migrationLog.Status = models.MigrationStatusFailed
		migrationLog.ErrorMessage = "Unsupported dataflow type"
		s.db.Save(&migrationLog)
		return fmt.Errorf("unsupported dataflow type: %s", dataflow.Type)
	}

	// Update the migration log
	migrationLog.Status = models.MigrationStatusSuccess
	return s.db.Save(&migrationLog).Error
}
