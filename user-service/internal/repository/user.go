// Package repository is the only layer that talks to the database.
// See author-service's repository package for the general idea.
package repository

import (
	"context"

	"github.com/fnxr21/user-service/internal/model"
	"gorm.io/gorm"
)

// UserRepository lists every database operation user-service needs.
// Missing rows come back as gorm.ErrRecordNotFound.
type UserRepository interface {
	// Create returns gorm.ErrDuplicatedKey if the email is already taken.
	Create(ctx context.Context, user model.User) (model.User, error)
	GetByID(ctx context.Context, id uint64) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user model.User) (model.User, error) {
	err := r.db.WithContext(ctx).Create(&user).Error
	return user, err
}

func (r *userRepository) GetByID(ctx context.Context, id uint64) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return user, err
}

// GetByEmail uses a "?" placeholder: GORM sends the value separately from the
// SQL, so user input can never change the query (no SQL injection).
func (r *userRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}
