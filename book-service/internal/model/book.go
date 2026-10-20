// Package model holds the database tables as Go structs (GORM models).
package model

import "gorm.io/gorm"

// Book is one row of the "book" table. gorm.Model adds ID, timestamps and
// soft delete (see author-service's model for details).
type Book struct {
	gorm.Model
	Name        string `gorm:"type:varchar(255);column:name"`
	Description string `gorm:"type:varchar(255);column:description"`
	AuthorID    uint   `gorm:"type:int;column:author_id"`
	// AuthorName is a copy of the author's name from author-service. Each
	// service owns its own database, so we cannot JOIN across them; copying
	// the name lets a book be read without calling author-service.
	AuthorName string `gorm:"type:varchar(255);column:author"`
	// The CHECK constraint is the database's last line of defence: even a
	// buggy query cannot store a negative stock.
	Stock int32 `gorm:"type:int;column:stock;not null;default:0;check:stock_non_negative,stock >= 0"`
	// Has-many relation: a book has many BookCategory rows (book_id = book.id).
	// Deleting a book row for real also deletes its links (ON DELETE CASCADE).
	Categories []BookCategory `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
}

// BookCategory links a book to a category. The category itself lives in
// category-service's database, so only its ID and a copy of its name are
// stored here. (BookID, CategoryID) is the primary key, so a book cannot be
// linked to the same category twice.
type BookCategory struct {
	BookID       uint   `gorm:"primaryKey;autoIncrement:false"`
	CategoryID   uint   `gorm:"primaryKey;autoIncrement:false"`
	CategoryName string `gorm:"type:varchar(255);column:categoryname"`
}
