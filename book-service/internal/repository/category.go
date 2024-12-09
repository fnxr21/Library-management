package repository

import (
	"context"

	"github.com/fnxr21/book-service/internal/model"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func RepositoryBook(db *gorm.DB) *repository {
	return &repository{db}
}

type BookService interface {
	CreateBook(ctx context.Context, book model.Book) (model.Book, error)
	GetBookList() ([]model.Book, error)
	GetBookById(ctx context.Context, id int) (model.Book, error)
	UpdateBook(ctx context.Context, book model.Book) (model.Book, error)
	DeleteBook(ctx context.Context, book model.Book) (model.Book, error)
}

// CreateBook creates a new book in the database.
func (r *repository) CreateBook(ctx context.Context, book model.Book) (model.Book, error) {
	err := r.db.WithContext(ctx).Create(&book).Error
	return book, err
}

// GetBookList retrieves all books from the database.
func (r *repository) GetBookList() ([]model.Book, error) {
	var books []model.Book
	err := r.db.Find(&books).Error
	return books, err
}

// GetBookById retrieves a specific book by their ID.
func (r *repository) GetBookById(ctx context.Context, id int) (model.Book, error) {
	var book model.Book
	err := r.db.WithContext(ctx).First(&book, id).Error
	return book, err
}

// UpdateBook updates an existing book in the database.
func (r *repository) UpdateBook(ctx context.Context, book model.Book) (model.Book, error) {
	err := r.db.WithContext(ctx).Save(&book).Error
	return book, err
}

// DeleteBook deletes an book from the database.
func (r *repository) DeleteBook(ctx context.Context, book model.Book) (model.Book, error) {
	err := r.db.WithContext(ctx).Delete(&book).Error
	return book, err
}
