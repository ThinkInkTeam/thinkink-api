package database

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ThinkInkTeam/thinkink-core-backend/models"
)

// DatabaseManager handles the database connection and operations
type DatabaseManager struct {
	DB *gorm.DB
}

// NewDatabaseManager creates a new database manager instance
func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{}
}

// Connect establishes a connection to the PostgreSQL database
func (dm *DatabaseManager) Connect(host, user, password, dbname, port, sslMode string) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", 
		host, user, password, dbname, port, sslMode)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	dm.DB = db
	err = dm.MigrateModels()
	if err != nil {
		return fmt.Errorf("failed to migrate database models: %w", err)
	}
	DB = dm.GetDB()
	return nil
}

// MigrateModels runs auto migration for the database models
func (dm *DatabaseManager) MigrateModels() error {
	if dm.DB == nil {
		return fmt.Errorf("database connection not established")
	}
	
	return dm.DB.AutoMigrate(
		&models.User{}, 
		&models.Report{}, 
		&models.BlacklistedToken{}, 
		&models.SingleFile{},
	)
}

// GetDB returns the gorm DB instance
func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.DB
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	if dm.DB == nil {
		return nil
	}
	
	sqlDB, err := dm.DB.DB()
	if err != nil {
		return err
	}
	
	return sqlDB.Close()
}

// For backward compatibility
var DB *gorm.DB

// getEnvWithDefault returns the environment variable value or a default if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
