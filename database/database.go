package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"thinkink-core-backend/models"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := "username:password@tcp(localhost:3306)/thinkink_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	db.AutoMigrate(&models.User{}, &models.Report{}, &models.BlacklistedToken{})
	DB = db
}
