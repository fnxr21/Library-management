package model

import "gorm.io/gorm"

type Book struct {
	gorm.Model
	Name        string         `gorm:"type:varchar(255);column:name"`
	Description string         `gorm:"type:varchar(255);column:description"`
	AuthorId    int            `gorm:"type:int;column:author_id"`
	Author      string         `gorm:"type:varchar(255);column:author"`
	Stock       int            `gorm:"type:int;column:stock"`
	Categories  []BookCategory `gorm:"many2many:book_categories_books"`
}

type BookCategory struct {
	gorm.Model
	BookID     uint `gorm:"primaryKey;autoIncrement:false" json:"book_id"`     // Book ID
	CategoryID uint `gorm:"primaryKey;autoIncrement:false" json:"category_id"` // Category ID
}
