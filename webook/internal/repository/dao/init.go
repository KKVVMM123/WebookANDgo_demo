package dao

import "github.com/jinzhu/gorm"

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{}).Error
}
