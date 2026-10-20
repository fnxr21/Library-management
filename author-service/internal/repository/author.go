// Package repository is the only layer that talks to the database.
//
// Handlers depend on the AuthorRepository interface, not on GORM. That keeps
// SQL out of the handlers and lets tests swap in an in-memory fake
// (see internal/handler/author_test.go).
package repository

import (
	"context"

	"github.com/fnxr21/author-service/internal/model"
	"gorm.io/gorm"
)

// AuthorRepository lists every database operation author-service needs.
// Missing rows come back as gorm.ErrRecordNotFound.
type AuthorRepository interface {
	Create(ctx context.Context, author model.Author) (model.Author, error)
	List(ctx context.Context) ([]model.Author, error)
	GetByID(ctx context.Context, id uint64) (model.Author, error)
	Update(ctx context.Context, author model.Author) (model.Author, error)
	Delete(ctx context.Context, author model.Author) error
}

type authorRepository struct {
	db *gorm.DB
}

func NewAuthorRepository(db *gorm.DB) AuthorRepository {
	return &authorRepository{db: db}
}

// Every method passes ctx to GORM with WithContext. If the client gives up
// (timeout, closed connection), the query is cancelled too.

func (r *authorRepository) Create(ctx context.Context, author model.Author) (model.Author, error) {
	err := r.db.WithContext(ctx).Create(&author).Error
	return author, err
}

func (r *authorRepository) List(ctx context.Context) ([]model.Author, error) {
	var authors []model.Author
	err := r.db.WithContext(ctx).Find(&authors).Error
	return authors, err
}

func (r *authorRepository) GetByID(ctx context.Context, id uint64) (model.Author, error) {
	var author model.Author
	err := r.db.WithContext(ctx).First(&author, id).Error
	return author, err
}

// Update saves all fields of the author.
func (r *authorRepository) Update(ctx context.Context, author model.Author) (model.Author, error) {
	err := r.db.WithContext(ctx).Save(&author).Error
	return author, err
}

// Delete soft-deletes the author (see model.Author).
func (r *authorRepository) Delete(ctx context.Context, author model.Author) error {
	return r.db.WithContext(ctx).Delete(&author).Error
}
