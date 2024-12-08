package repository

import (
	"context"

	"github.com/fnxr21/author-service/internal/model"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func RepositoryAuthor(db *gorm.DB) *repository {
	return &repository{db}
}

type AuthorService interface {
	CreateAuthor(ctx context.Context, author model.Author) (model.Author, error)
	GetAuthorList() ([]model.Author, error)
	GetAuthorById(ctx context.Context, id int) (model.Author, error)
	UpdateAuthor(ctx context.Context, author model.Author) (model.Author, error)
	DeleteAuthor(ctx context.Context, author model.Author) (model.Author, error)
}

// CreateAuthor creates a new author in the database.
func (r *repository) CreateAuthor(ctx context.Context, author model.Author) (model.Author, error) {
	err := r.db.WithContext(ctx).Create(&author).Error
	return author, err
}

// GetAuthorList retrieves all authors from the database.
func (r *repository) GetAuthorList() ([]model.Author, error) {
	var authors []model.Author
	err := r.db.Find(&authors).Error
	return authors, err
}

// GetAuthorById retrieves a specific author by their ID.
func (r *repository) GetAuthorById(ctx context.Context, id int) (model.Author, error) {
	var author model.Author
	err := r.db.WithContext(ctx).First(&author, id).Error
	return author, err
}

// UpdateAuthor updates an existing author in the database.
func (r *repository) UpdateAuthor(ctx context.Context, author model.Author) (model.Author, error) {
	err := r.db.WithContext(ctx).Save(&author).Error
	return author, err
}

// DeleteAuthor deletes an author from the database.
func (r *repository) DeleteAuthor(ctx context.Context, author model.Author) (model.Author, error) {
	err := r.db.WithContext(ctx).Delete(&author).Error
	return author, err
}

