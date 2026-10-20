// Package repository is the only layer that talks to the database.
// See author-service's repository package for the general idea.
package repository

import (
	"context"
	"errors"

	"github.com/fnxr21/book-service/internal/model"
	"gorm.io/gorm"
)

// ErrInsufficientStock is returned when a stock change would go below zero.
var ErrInsufficientStock = errors.New("insufficient stock")

// BookRepository lists every database operation book-service needs.
// Missing rows come back as gorm.ErrRecordNotFound.
type BookRepository interface {
	Create(ctx context.Context, book model.Book) (model.Book, error)
	List(ctx context.Context) ([]model.Book, error)
	GetByID(ctx context.Context, id uint64) (model.Book, error)
	Update(ctx context.Context, book model.Book, categories []model.BookCategory) (model.Book, error)
	AdjustStock(ctx context.Context, id uint64, delta int32) (model.Book, error)
	Delete(ctx context.Context, book model.Book) error
}

type bookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) BookRepository {
	return &bookRepository{db: db}
}

// Create inserts the book and its category links. GORM saves the
// Categories slice too, inside one transaction: all rows or none.
func (r *bookRepository) Create(ctx context.Context, book model.Book) (model.Book, error) {
	err := r.db.WithContext(ctx).Create(&book).Error
	return book, err
}

// List returns all books. Preload("Categories") runs a second query to fill
// each book's Categories; without it the slice would stay empty.
func (r *bookRepository) List(ctx context.Context) ([]model.Book, error) {
	var books []model.Book
	err := r.db.WithContext(ctx).Preload("Categories").Find(&books).Error
	return books, err
}

func (r *bookRepository) GetByID(ctx context.Context, id uint64) (model.Book, error) {
	var book model.Book
	err := r.db.WithContext(ctx).Preload("Categories").First(&book, id).Error
	return book, err
}

// Update saves the book's own fields. Stock is left alone; use AdjustStock.
// When categories is non-nil it replaces the book's category links.
//
// Everything runs in one transaction: if any step fails, the whole update
// is rolled back and the book is left as it was.
func (r *bookRepository) Update(ctx context.Context, book model.Book, categories []model.BookCategory) (model.Book, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Select limits the UPDATE to these columns, so a stale Stock value
		// in book can never overwrite a concurrent stock change.
		err := tx.Model(&book).Select("name", "description", "author_id", "author").Updates(&book).Error
		if err != nil {
			return err
		}
		if categories == nil {
			return nil
		}
		if err := tx.Where("book_id = ?", book.ID).Delete(&model.BookCategory{}).Error; err != nil {
			return err
		}
		for i := range categories {
			categories[i].BookID = book.ID
		}
		if len(categories) > 0 {
			return tx.Create(&categories).Error
		}
		return nil
	})
	if err != nil {
		return book, err
	}
	return r.GetByID(ctx, uint64(book.ID))
}

// AdjustStock adds delta to the stock in ONE conditional UPDATE:
//
//	UPDATE book SET stock = stock + ? WHERE id = ? AND stock + ? >= 0
//
// Why not "read stock, check it, write stock" in Go? Two requests could both
// read stock=1, both decide "ok", and both write 0, so two copies are lent
// out but only one existed (a race condition). Here the database checks and
// updates in a single step, so concurrent requests cannot oversell.
func (r *bookRepository) AdjustStock(ctx context.Context, id uint64, delta int32) (model.Book, error) {
	res := r.db.WithContext(ctx).Model(&model.Book{}).
		Where("id = ? AND stock + ? >= 0", id, delta).
		Update("stock", gorm.Expr("stock + ?", delta))
	if res.Error != nil {
		return model.Book{}, res.Error
	}
	if res.RowsAffected == 0 {
		// No row matched: either the book is missing or the stock is too low.
		if _, err := r.GetByID(ctx, id); err != nil {
			return model.Book{}, err
		}
		return model.Book{}, ErrInsufficientStock
	}
	return r.GetByID(ctx, id)
}

// Delete soft-deletes the book. Its category links stay, so a restored book
// keeps its categories.
func (r *bookRepository) Delete(ctx context.Context, book model.Book) error {
	return r.db.WithContext(ctx).Delete(&book).Error
}
