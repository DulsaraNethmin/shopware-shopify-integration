package services

import (
	"errors"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/models"
	"gorm.io/gorm"
)

// ConnectorService handles connector operations
type ConnectorService struct {
	db *gorm.DB
}

// NewConnectorService creates a new connector service
func NewConnectorService(db *gorm.DB) *ConnectorService {
	return &ConnectorService{
		db: db,
	}
}

// CreateConnector creates a new connector
func (s *ConnectorService) CreateConnector(connector *models.Connector) error {
	return s.db.Create(connector).Error
}

// GetConnector gets a connector by ID
func (s *ConnectorService) GetConnector(id uint) (*models.Connector, error) {
	var connector models.Connector

	if err := s.db.First(&connector, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &connector, nil
}

// ListConnectors lists all connectors
func (s *ConnectorService) ListConnectors(connectorType *models.ConnectorType) ([]models.Connector, error) {
	var connectors []models.Connector

	query := s.db

	if connectorType != nil {
		query = query.Where("type = ?", *connectorType)
	}

	if err := query.Find(&connectors).Error; err != nil {
		return nil, err
	}

	return connectors, nil
}

// UpdateConnector updates a connector
func (s *ConnectorService) UpdateConnector(id uint, connector *models.Connector) error {
	// Check if the connector exists
	existingConnector, err := s.GetConnector(id)
	if err != nil {
		return err
	}

	// Update the connector
	connector.ID = existingConnector.ID
	return s.db.Save(connector).Error
}

// DeleteConnector deletes a connector
func (s *ConnectorService) DeleteConnector(id uint) error {
	// Check if the connector exists
	existingConnector, err := s.GetConnector(id)
	if err != nil {
		return err
	}

	// Check if the connector is used in any dataflows
	var count int64
	if err := s.db.Model(&models.Dataflow{}).Where("source_connector_id = ? OR dest_connector_id = ?", id, id).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("connector is used in dataflows and cannot be deleted")
	}

	// Delete the connector
	return s.db.Delete(existingConnector).Error
}

// TestConnection tests the connection to the connector
func (s *ConnectorService) TestConnection(id uint) error {
	connector, err := s.GetConnector(id)
	if err != nil {
		return err
	}

	switch connector.Type {
	case models.ConnectorTypeShopware:
		shopwareService := NewShopwareService(s.db)
		return shopwareService.TestConnection(connector)
	case models.ConnectorTypeShopify:
		shopifyService := NewShopifyService(s.db)
		return shopifyService.TestConnection(connector)
	default:
		return models.ErrInvalidConnectorType
	}
}

// RegisterWebhooks registers webhooks for the connector
func (s *ConnectorService) RegisterWebhooks(id uint, callbackURL string) error {
	connector, err := s.GetConnector(id)
	if err != nil {
		return err
	}

	switch connector.Type {
	case models.ConnectorTypeShopware:
		shopwareService := NewShopwareService(s.db)
		return shopwareService.RegisterWebhooks(connector, callbackURL)
	case models.ConnectorTypeShopify:
		shopifyService := NewShopifyService(s.db)
		return shopifyService.RegisterWebhooks(connector, callbackURL)
	default:
		return models.ErrInvalidConnectorType
	}
}
