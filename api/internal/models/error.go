package models

import "errors"

// Application errors
var (
	ErrInvalidConnector       = errors.New("invalid connector: name and URL are required")
	ErrInvalidConnectorType   = errors.New("invalid connector type")
	ErrInvalidCredentials     = errors.New("invalid credentials for connector type")
	ErrInvalidDataflow        = errors.New("invalid dataflow: name is required")
	ErrSameConnector          = errors.New("source and destination connectors must be different")
	ErrInvalidSourceConnector = errors.New("source connector must be a Shopware connector")
	ErrInvalidDestConnector   = errors.New("destination connector must be a Shopify connector")
	ErrInvalidFieldMapping    = errors.New("invalid field mapping: source and destination fields are required")
)
