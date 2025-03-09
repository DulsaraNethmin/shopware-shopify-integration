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
	ApiKey      string        `json:"api_key,omitempty" gorm:"column:api_key"`
	ApiSecret   string        `json:"api_secret,omitempty" gorm:"column:api_secret"`
	AccessToken string        `json:"access_token,omitempty" gorm:"column:access_token"`
	Password    string        `json:"password,omitempty" gorm:"column:password"`
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

	//TODO:Fix This Later
	//switch c.Type {
	//case ConnectorTypeShopware:
	//	fmt.Printf("Api Key: %s, Api Secret: %s\n", c.ApiKey, c.ApiSecret)
	//	if c.ApiKey == "" || c.ApiSecret == "" {
	//		return ErrInvalidCredentials
	//	}
	//case ConnectorTypeShopify:
	//	fmt.Printf("Api Key: %s, Api Secret: %s\n", c.ApiKey, c.ApiSecret)
	//	if c.ApiKey == "" || c.ApiSecret == "" {
	//		return ErrInvalidCredentials
	//	}
	//default:
	//	return ErrInvalidConnectorType
	//}

	return nil
}
