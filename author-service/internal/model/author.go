// Package model holds the database tables as Go structs (GORM models).
package model

import "gorm.io/gorm"

// Author is one row of the "author" table.
//
// gorm.Model adds four columns: ID (primary key), CreatedAt, UpdatedAt and
// DeletedAt. DeletedAt enables soft delete: Delete() only sets the timestamp
// and normal queries skip those rows, so data can still be recovered.
type Author struct {
	gorm.Model
	Name      string `gorm:"type:varchar(255);column:name"`
	Biography string `gorm:"type:varchar(255);column:biography"`
}
