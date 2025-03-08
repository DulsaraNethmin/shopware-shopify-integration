package models

import (
	"time"

	"gorm.io/gorm"
)

// MigrationStatus represents the status of a migration
type MigrationStatus string

const (
	// MigrationStatusPending represents a pending migration
	MigrationStatusPending MigrationStatus = "pending"
	// MigrationStatusInProgress represents a migration in progress
	MigrationStatusInProgress MigrationStatus = "in_progress"
	// MigrationStatusSuccess represents a successful migration
	MigrationStatusSuccess MigrationStatus = "success"
	// MigrationStatusFailed represents a failed migration
	MigrationStatusFailed MigrationStatus = "failed"
)

// MigrationLog represents a log entry for a migration
type MigrationLog struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	DataflowID         uint            `json:"dataflow_id" gorm:"not null"`
	Status             MigrationStatus `json:"status" gorm:"default:'pending'"`
	SourceIdentifier   string          `json:"source_identifier" gorm:"not null"` // ID in the source system
	DestIdentifier     string          `json:"dest_identifier"`                   // ID in the destination system
	ExecutionARN       string          `json:"execution_arn"`                     // AWS Step Functions execution ARN
	SourcePayload      string          `json:"source_payload"`                    // JSON string with source data
	TransformedPayload string          `json:"transformed_payload"`               // JSON string with transformed data
	ErrorMessage       string          `json:"error_message"`
	CompletedAt        *time.Time      `json:"completed_at"`

	// Relations
	Dataflow Dataflow `json:"-" gorm:"foreignKey:DataflowID"`
}
