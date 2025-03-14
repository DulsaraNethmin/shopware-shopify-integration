package models

import (
	"time"

	"gorm.io/gorm"
)

// TransformationType represents the type of transformation to apply
type TransformationType string

const (
	// TransformationTypeNone means no transformation, direct mapping
	TransformationTypeNone TransformationType = "none"
	// TransformationTypeFormat means format transformation (e.g., date format)
	TransformationTypeFormat TransformationType = "format"
	// TransformationTypeConvert means type conversion
	TransformationTypeConvert TransformationType = "convert"
	// TransformationTypeMap means mapping values (e.g., status codes)
	TransformationTypeMap TransformationType = "map"
	// TransformationTypeTemplate means using a template
	TransformationTypeTemplate     TransformationType = "template"
	TransformationTypeGraphQLID    TransformationType = "graphql_id"
	TransformationTypeArrayMap     TransformationType = "array_map"
	TransformationTypeJsonPath     TransformationType = "json_path"
	TransformationTypeConditional  TransformationType = "conditional"
	TransformationTypeMediaMap     TransformationType = "media_map"
	TransformationTypeMetafield    TransformationType = "metafield"
	TransformationTypeEntityLookup TransformationType = "entity_lookup"
)

// FieldMapping represents a mapping between source and destination fields
type FieldMapping struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	DataflowID      uint               `json:"dataflow_id" gorm:"not null"`
	SourceField     string             `json:"source_field" gorm:"not null"`
	DestField       string             `json:"dest_field" gorm:"not null"`
	IsRequired      bool               `json:"is_required" gorm:"default:false"`
	DefaultValue    string             `json:"default_value"`
	TransformType   TransformationType `json:"transform_type" gorm:"default:'none'"`
	TransformConfig string             `json:"transform_config"` // JSON string with transformation config

	// Relations
	Dataflow Dataflow `json:"-" gorm:"foreignKey:DataflowID"`
}

// BeforeCreate is a GORM hook that runs before creating a new record
func (fm *FieldMapping) BeforeCreate(tx *gorm.DB) error {
	if fm.SourceField == "" || fm.DestField == "" {
		return ErrInvalidFieldMapping
	}

	return nil
}
