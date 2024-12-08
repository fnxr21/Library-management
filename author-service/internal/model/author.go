package model

import "gorm.io/gorm"

type Author struct {
	gorm.Model
	Name      string `gorm:"type:varchar(255);column:name"`
	Biography string `gorm:"type:varchar(255);column:biography"`
}
