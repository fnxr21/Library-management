// Package repository is the only layer that talks to the database.
//
// Handlers depend on the CategoryRepository interface, not on GORM. That keeps
// SQL out of the handlers and lets tests swap in an in-memory fake
// (see internal/handler/category_test.go).
package repository

import (
	"context"

	"github.com/fnxr21/category-service/internal/model"
	"gorm.io/gorm"
)

// CategoryRepository lists every database operation category-service needs.
// Missing rows come back as gorm.ErrRecordNotFound.
type CategoryRepository interface {
	Create(ctx context.Context, category model.Category) (model.Category, error)
	List(ctx context.Context) ([]model.Category, error)
	GetByID(ctx context.Context, id uint64) (model.Category, error)
	Update(ctx context.Context, category model.Category) (model.Category, error)
	Delete(ctx context.Context, category model.Category) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// Every method passes ctx to GORM with WithContext. If the client gives up
// (timeout, closed connection), the query is cancelled too.

func (r *categoryRepository) Create(ctx context.Context, category model.Category) (model.Category, error) {
	err := r.db.WithContext(ctx).Create(&category).Error
	return category, err
}

func (r *categoryRepository) List(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.WithContext(ctx).Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetByID(ctx context.Context, id uint64) (model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).First(&category, id).Error
	return category, err
}

// Update saves all fields of the category.
func (r *categoryRepository) Update(ctx context.Context, category model.Category) (model.Category, error) {
	err := r.db.WithContext(ctx).Save(&category).Error
	return category, err
}

// Delete soft-deletes the category (see model.Category).
func (r *categoryRepository) Delete(ctx context.Context, category model.Category) error {
	return r.db.WithContext(ctx).Delete(&category).Error
}
