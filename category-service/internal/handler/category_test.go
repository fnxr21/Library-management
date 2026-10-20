package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/fnxr21/category-service/internal/model"
	"github.com/fnxr21/category-service/protobuf/categorypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// fakeRepo is an in-memory repository.CategoryRepository. Because the handler
// only depends on the interface, tests run without a database.
type fakeRepo struct {
	rows   map[uint]model.Category
	nextID uint
	err    error // when set, Create fails with it
}

func newFakeRepo() *fakeRepo { return &fakeRepo{rows: map[uint]model.Category{}, nextID: 1} }

func (f *fakeRepo) Create(_ context.Context, a model.Category) (model.Category, error) {
	if f.err != nil {
		return model.Category{}, f.err
	}
	a.ID = f.nextID
	f.nextID++
	f.rows[a.ID] = a
	return a, nil
}

func (f *fakeRepo) List(_ context.Context) ([]model.Category, error) {
	var out []model.Category
	for _, a := range f.rows {
		out = append(out, a)
	}
	return out, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id uint64) (model.Category, error) {
	a, ok := f.rows[uint(id)]
	if !ok {
		return model.Category{}, gorm.ErrRecordNotFound
	}
	return a, nil
}

func (f *fakeRepo) Update(_ context.Context, a model.Category) (model.Category, error) {
	f.rows[a.ID] = a
	return a, nil
}

func (f *fakeRepo) Delete(_ context.Context, a model.Category) error {
	delete(f.rows, a.ID)
	return nil
}

func TestCreateCategory(t *testing.T) {
	h := NewCategoryHandler(newFakeRepo())

	got, err := h.CreateCategory(context.Background(), &categorypb.CreateCategoryRequest{Name: "Science Fiction", Description: "Stories"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != 1 || got.Name != "Science Fiction" {
		t.Fatalf("unexpected category: %+v", got)
	}
}

func TestCreateCategoryRequiresName(t *testing.T) {
	h := NewCategoryHandler(newFakeRepo())

	_, err := h.CreateCategory(context.Background(), &categorypb.CreateCategoryRequest{Name: "  "})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("want InvalidArgument, got %v", err)
	}
}

func TestCreateCategoryHidesDatabaseError(t *testing.T) {
	repo := newFakeRepo()
	repo.err = errors.New("connection refused")
	h := NewCategoryHandler(repo)

	_, err := h.CreateCategory(context.Background(), &categorypb.CreateCategoryRequest{Name: "Fiction"})
	if status.Code(err) != codes.Internal || status.Convert(err).Message() != "internal error" {
		t.Fatalf("want generic Internal, got %v", err)
	}
}

func TestGetCategoryNotFound(t *testing.T) {
	h := NewCategoryHandler(newFakeRepo())

	_, err := h.GetCategory(context.Background(), &categorypb.GetCategoryRequest{Id: 42})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("want NotFound, got %v", err)
	}
}

func TestUpdateCategoryKeepsUnsetFields(t *testing.T) {
	h := NewCategoryHandler(newFakeRepo())
	created, _ := h.CreateCategory(context.Background(), &categorypb.CreateCategoryRequest{Name: "Fiction", Description: "Old description"})

	got, err := h.UpdateCategory(context.Background(), &categorypb.UpdateCategoryRequest{Id: created.Id, Name: "Science Fiction"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Science Fiction" || got.Description != "Old description" {
		t.Fatalf("unexpected category: %+v", got)
	}
}

func TestDeleteCategory(t *testing.T) {
	h := NewCategoryHandler(newFakeRepo())
	created, _ := h.CreateCategory(context.Background(), &categorypb.CreateCategoryRequest{Name: "Fiction"})

	if _, err := h.DeleteCategory(context.Background(), &categorypb.DeleteCategoryRequest{Id: created.Id}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := h.DeleteCategory(context.Background(), &categorypb.DeleteCategoryRequest{Id: created.Id}); status.Code(err) != codes.NotFound {
		t.Fatalf("second delete: want NotFound, got %v", err)
	}
}
