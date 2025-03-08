package models

import (
	"time"

	"gorm.io/gorm"
)

// ConnectorType represents the type of connector
type ConnectorType string

const (
	// ConnectorTypeShopware represents a Shopware connector
	ConnectorTypeShopware ConnectorType = "shopware"
	// ConnectorTypeShopify represents a Shopify connector
	ConnectorTypeShopify ConnectorType = "shopify"
)

// Connector represents a connection to an external system
type Connector struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string        `json:"name" gorm:"not null"`
	Type        ConnectorType `json:"type" gorm:"not null"`
	URL         string        `json:"url" gorm:"not null"`
	Username    string        `json:"username"`
	Password    string        `json:"-"` // Sensitive information not returned in JSON
	ApiKey      string        `json:"-"` // Sensitive information not returned in JSON
	ApiSecret   string        `json:"-"` // Sensitive information not returned in JSON
	AccessToken string        `json:"-"` // Sensitive information not returned in JSON
	IsActive    bool          `json:"is_active" gorm:"default:true"`

	// Relations
	Dataflows []Dataflow `json:"-" gorm:"foreignKey:SourceConnectorID;references:ID"`
}

// BeforeCreate is a GORM hook that runs before creating a new record
func (c *Connector) BeforeCreate(tx *gorm.DB) error {
	// Validate required fields
	if c.Name == "" || c.URL == "" {
		return ErrInvalidConnector
	}

	// Additional validation based on connector type
	switch c.Type {
	case ConnectorTypeShopware:
		if c.Username == "" || c.Password == "" {
			return ErrInvalidCredentials
		}
	case ConnectorTypeShopify:
		if c.ApiKey == "" || c.ApiSecret == "" {
			return ErrInvalidCredentials
		}
	default:
		return ErrInvalidConnectorType
	}

	return nil
}
