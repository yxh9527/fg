package dao

import (
	"app/tables/manager"
	"errors"

	"gorm.io/gorm"
)

func UpsertPoolConfig(key, value string) error {
	var row manager.PoolConfig
	db := Mysql().Manager
	err := db.Where("`key` = ?", key).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&manager.PoolConfig{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&row).Update("value", value).Error
}
