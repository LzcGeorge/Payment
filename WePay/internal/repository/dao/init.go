package dao

import (
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

func TruncateTable(db *gorm.DB, tableName string) error {
	return db.Exec("TRUNCATE TABLE " + tableName).Error
}
