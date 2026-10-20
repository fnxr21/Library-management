package repository

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/fnxr21/book-service/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// testDB connects to the Postgres in BOOK_TEST_DSN, e.g.
// "host=localhost port=4013 user=test password=test dbname=book_mst sslmode=disable".
// The test is skipped when it is unset.
func testDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("BOOK_TEST_DSN")
	if dsn == "" {
		t.Skip("BOOK_TEST_DSN not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := db.AutoMigrate(&model.Book{}, &model.BookCategory{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestAdjustStockConcurrentBorrowsNeverOversell(t *testing.T) {
	db := testDB(t)
	repo := NewBookRepository(db)
	ctx := context.Background()

	book, err := repo.Create(ctx, model.Book{Name: "Race Test", AuthorID: 1, Stock: 5})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&book) })

	const borrowers = 20
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok, rejected := 0, 0
	for i := 0; i < borrowers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.AdjustStock(ctx, uint64(book.ID), -1)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				ok++
			case errors.Is(err, ErrInsufficientStock):
				rejected++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	got, err := repo.GetByID(ctx, uint64(book.ID))
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if ok != 5 || rejected != borrowers-5 || got.Stock != 0 {
		t.Fatalf("ok=%d rejected=%d stock=%d, want 5/%d/0", ok, rejected, got.Stock, borrowers-5)
	}
}

func TestUpdateBookReplacesCategoryLinks(t *testing.T) {
	db := testDB(t)
	repo := NewBookRepository(db)
	ctx := context.Background()

	book, err := repo.Create(ctx, model.Book{Name: "Links Test", AuthorID: 1, Categories: []model.BookCategory{
		{CategoryID: 1, CategoryName: "Fiction"}, {CategoryID: 2, CategoryName: "Science"},
	}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&book) })

	book.Name = "Links Test 2e"
	got, err := repo.Update(ctx, book, []model.BookCategory{{CategoryID: 2, CategoryName: "Science"}})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.Name != "Links Test 2e" || len(got.Categories) != 1 || got.Categories[0].CategoryID != 2 {
		t.Fatalf("unexpected book: %+v", got)
	}
}
