package db

import (
	"fmt"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/config"
	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Init initializes the database connection
func Init(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// AutoMigrate automatically migrates the database schema
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Connector{},
		&models.Dataflow{},
		&models.FieldMapping{},
		&models.MigrationLog{},
	)
}
