package models

import (
	"time"

	"gorm.io/gorm"
)

// DataflowType represents the type of data that flows
type DataflowType string

const (
	// DataflowTypeProduct represents a product dataflow
	DataflowTypeProduct DataflowType = "product"
	// DataflowTypeOrder represents an order dataflow
	DataflowTypeOrder DataflowType = "order"
)

// DataflowStatus represents the status of a dataflow
type DataflowStatus string

const (
	// DataflowStatusActive represents an active dataflow
	DataflowStatusActive DataflowStatus = "active"
	// DataflowStatusInactive represents an inactive dataflow
	DataflowStatusInactive DataflowStatus = "inactive"
)

// Dataflow represents a data flow between connectors
type Dataflow struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name              string         `json:"name" gorm:"not null"`
	Description       string         `json:"description"`
	Type              DataflowType   `json:"type" gorm:"not null"`
	Status            DataflowStatus `json:"status" gorm:"default:'active'"`
	SourceConnectorID uint           `json:"source_connector_id" gorm:"not null"`
	DestConnectorID   uint           `json:"dest_connector_id" gorm:"not null"`

	// Relations
	SourceConnector Connector      `json:"source_connector" gorm:"foreignKey:SourceConnectorID"`
	DestConnector   Connector      `json:"dest_connector" gorm:"foreignKey:DestConnectorID"`
	FieldMappings   []FieldMapping `json:"field_mappings" gorm:"foreignKey:DataflowID"`
	MigrationLogs   []MigrationLog `json:"-" gorm:"foreignKey:DataflowID"`
}

// BeforeCreate is a GORM hook that runs before creating a new record
func (d *Dataflow) BeforeCreate(tx *gorm.DB) error {
	if d.Name == "" {
		return ErrInvalidDataflow
	}

	// Ensure source and destination connectors are different
	if d.SourceConnectorID == d.DestConnectorID {
		return ErrSameConnector
	}

	// Validate connector types
	var sourceConnector, destConnector Connector

	if err := tx.First(&sourceConnector, d.SourceConnectorID).Error; err != nil {
		return err
	}

	if err := tx.First(&destConnector, d.DestConnectorID).Error; err != nil {
		return err
	}

	// For this project, source must be Shopware and dest must be Shopify
	if sourceConnector.Type != ConnectorTypeShopware {
		return ErrInvalidSourceConnector
	}

	if destConnector.Type != ConnectorTypeShopify {
		return ErrInvalidDestConnector
	}

	return nil
}
