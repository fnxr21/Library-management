package handler

import (
	"context"
	"testing"

	"github.com/fnxr21/book-service/internal/model"
	"github.com/fnxr21/book-service/internal/repository"
	"github.com/fnxr21/book-service/protobuf/authorpb"
	"github.com/fnxr21/book-service/protobuf/bookpb"
	"github.com/fnxr21/book-service/protobuf/categorypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// fakeRepo is an in-memory repository.BookRepository.
type fakeRepo struct {
	rows   map[uint]model.Book
	nextID uint
}

func newFakeRepo() *fakeRepo { return &fakeRepo{rows: map[uint]model.Book{}, nextID: 1} }

func (f *fakeRepo) Create(_ context.Context, b model.Book) (model.Book, error) {
	b.ID = f.nextID
	f.nextID++
	f.rows[b.ID] = b
	return b, nil
}

func (f *fakeRepo) List(_ context.Context) ([]model.Book, error) {
	var out []model.Book
	for _, b := range f.rows {
		out = append(out, b)
	}
	return out, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id uint64) (model.Book, error) {
	b, ok := f.rows[uint(id)]
	if !ok {
		return model.Book{}, gorm.ErrRecordNotFound
	}
	return b, nil
}

func (f *fakeRepo) Update(_ context.Context, b model.Book, categories []model.BookCategory) (model.Book, error) {
	if categories != nil {
		b.Categories = categories
	}
	f.rows[b.ID] = b
	return b, nil
}

func (f *fakeRepo) AdjustStock(_ context.Context, id uint64, delta int32) (model.Book, error) {
	b, ok := f.rows[uint(id)]
	if !ok {
		return model.Book{}, gorm.ErrRecordNotFound
	}
	if b.Stock+delta < 0 {
		return model.Book{}, repository.ErrInsufficientStock
	}
	b.Stock += delta
	f.rows[b.ID] = b
	return b, nil
}

func (f *fakeRepo) Delete(_ context.Context, b model.Book) error {
	delete(f.rows, b.ID)
	return nil
}

// fakeAuthors stands in for author-service. It knows author 1 only; err,
// when set, simulates the service being down. Embedding the interface means
// we only implement the methods the handler actually calls.
type fakeAuthors struct {
	authorpb.AuthorServiceClient
	err error
}

func (f fakeAuthors) GetAuthor(_ context.Context, in *authorpb.GetAuthorRequest, _ ...grpc.CallOption) (*authorpb.Author, error) {
	if f.err != nil {
		return nil, f.err
	}
	if in.Id != 1 {
		return nil, status.Error(codes.NotFound, "author not found")
	}
	return &authorpb.Author{Id: 1, Name: "Jane Doe"}, nil
}

// fakeCategories stands in for category-service. It knows categories 1 and 2.
type fakeCategories struct {
	categorypb.CategoryServiceClient
}

func (fakeCategories) GetCategory(_ context.Context, in *categorypb.GetCategoryRequest, _ ...grpc.CallOption) (*categorypb.Category, error) {
	names := map[uint64]string{1: "Fiction", 2: "Science"}
	name, ok := names[in.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "category not found")
	}
	return &categorypb.Category{Id: in.Id, Name: name}, nil
}

func newHandler() *BookHandler {
	return NewBookHandler(newFakeRepo(), fakeAuthors{}, fakeCategories{})
}

func validBook() *bookpb.CreateBookRequest {
	return &bookpb.CreateBookRequest{Name: "Sample Book", AuthorId: 1, Stock: 3, CategoryIds: []uint64{1, 2, 1}}
}

func categoryNames(b *bookpb.Book) []string {
	var names []string
	for _, c := range b.Categories {
		names = append(names, c.Name)
	}
	return names
}

func TestCreateBookResolvesAuthorAndCategories(t *testing.T) {
	got, err := newHandler().CreateBook(context.Background(), validBook())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AuthorName != "Jane Doe" {
		t.Fatalf("authorName = %q, want Jane Doe", got.AuthorName)
	}
	if names := categoryNames(got); len(names) != 2 || names[0] != "Fiction" || names[1] != "Science" {
		t.Fatalf("categories = %v, want [Fiction Science] (duplicates dropped)", names)
	}
}

func TestCreateBookValidation(t *testing.T) {
	cases := map[string]func(*bookpb.CreateBookRequest){
		"missing name":     func(b *bookpb.CreateBookRequest) { b.Name = "" },
		"missing author":   func(b *bookpb.CreateBookRequest) { b.AuthorId = 0 },
		"negative stock":   func(b *bookpb.CreateBookRequest) { b.Stock = -1 },
		"unknown author":   func(b *bookpb.CreateBookRequest) { b.AuthorId = 99 },
		"unknown category": func(b *bookpb.CreateBookRequest) { b.CategoryIds = []uint64{1, 99} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := validBook()
			mutate(req)
			_, err := newHandler().CreateBook(context.Background(), req)
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("want InvalidArgument, got %v", err)
			}
		})
	}
}

func TestCreateBookAuthorServiceDown(t *testing.T) {
	h := NewBookHandler(newFakeRepo(), fakeAuthors{err: status.Error(codes.Unavailable, "connection refused")}, fakeCategories{})

	_, err := h.CreateBook(context.Background(), validBook())
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("want Unavailable, got %v", err)
	}
}

func TestAdjustStock(t *testing.T) {
	h := newHandler()
	book, _ := h.CreateBook(context.Background(), validBook())

	got, err := h.AdjustStock(context.Background(), &bookpb.AdjustStockRequest{Id: book.Id, Delta: -3})
	if err != nil || got.Stock != 0 {
		t.Fatalf("borrow all: stock=%v err=%v", got.GetStock(), err)
	}
	_, err = h.AdjustStock(context.Background(), &bookpb.AdjustStockRequest{Id: book.Id, Delta: -1})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("borrow past zero: want FailedPrecondition, got %v", err)
	}
	_, err = h.AdjustStock(context.Background(), &bookpb.AdjustStockRequest{Id: 404, Delta: 1})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("missing book: want NotFound, got %v", err)
	}
}

func TestUpdateBookReplacesCategories(t *testing.T) {
	h := newHandler()
	book, _ := h.CreateBook(context.Background(), validBook())

	got, err := h.UpdateBook(context.Background(), &bookpb.UpdateBookRequest{Id: book.Id, CategoryIds: []uint64{2}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if names := categoryNames(got); got.Name != "Sample Book" || len(names) != 1 || names[0] != "Science" {
		t.Fatalf("unexpected book: %+v", got)
	}
}

func TestDeleteBookNotFound(t *testing.T) {
	_, err := newHandler().DeleteBook(context.Background(), &bookpb.DeleteBookRequest{Id: 7})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("want NotFound, got %v", err)
	}
}
