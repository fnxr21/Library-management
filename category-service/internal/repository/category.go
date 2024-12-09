package repository

import (
	"context"

	"github.com/fnxr21/category-service/internal/model"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func RepositoryCategory(db *gorm.DB) *repository {
	return &repository{db}
}

type CategoryService interface {
	CreateCategory(ctx context.Context, category model.Category) (model.Category, error)
	GetCategoryList() ([]model.Category, error)
	GetCategoryById(ctx context.Context, id int) (model.Category, error)
	UpdateCategory(ctx context.Context, category model.Category) (model.Category, error)
	DeleteCategory(ctx context.Context, category model.Category) (model.Category, error)
}

// CreateCategory creates a new category in the database.
func (r *repository) CreateCategory(ctx context.Context, category model.Category) (model.Category, error) {
	err := r.db.WithContext(ctx).Create(&category).Error
	return category, err
}

// GetCategoryList retrieves all categorys from the database.
func (r *repository) GetCategoryList() ([]model.Category, error) {
	var categorys []model.Category
	err := r.db.Find(&categorys).Error
	return categorys, err
}

// GetCategoryById retrieves a specific category by their ID.
func (r *repository) GetCategoryById(ctx context.Context, id int) (model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).First(&category, id).Error
	return category, err
}

// UpdateCategory updates an existing category in the database.
func (r *repository) UpdateCategory(ctx context.Context, category model.Category) (model.Category, error) {
	err := r.db.WithContext(ctx).Save(&category).Error
	return category, err
}

// DeleteCategory deletes an category from the database.
func (r *repository) DeleteCategory(ctx context.Context, category model.Category) (model.Category, error) {
	err := r.db.WithContext(ctx).Delete(&category).Error
	return category, err
}
