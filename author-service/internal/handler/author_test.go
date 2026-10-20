package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/fnxr21/author-service/internal/model"
	"github.com/fnxr21/author-service/protobuf/authorpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// fakeRepo is an in-memory repository.AuthorRepository. Because the handler
// only depends on the interface, tests run without a database.
type fakeRepo struct {
	rows   map[uint]model.Author
	nextID uint
	err    error // when set, Create fails with it
}

func newFakeRepo() *fakeRepo { return &fakeRepo{rows: map[uint]model.Author{}, nextID: 1} }

func (f *fakeRepo) Create(_ context.Context, a model.Author) (model.Author, error) {
	if f.err != nil {
		return model.Author{}, f.err
	}
	a.ID = f.nextID
	f.nextID++
	f.rows[a.ID] = a
	return a, nil
}

func (f *fakeRepo) List(_ context.Context) ([]model.Author, error) {
	var out []model.Author
	for _, a := range f.rows {
		out = append(out, a)
	}
	return out, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id uint64) (model.Author, error) {
	a, ok := f.rows[uint(id)]
	if !ok {
		return model.Author{}, gorm.ErrRecordNotFound
	}
	return a, nil
}

func (f *fakeRepo) Update(_ context.Context, a model.Author) (model.Author, error) {
	f.rows[a.ID] = a
	return a, nil
}

func (f *fakeRepo) Delete(_ context.Context, a model.Author) error {
	delete(f.rows, a.ID)
	return nil
}

func TestCreateAuthor(t *testing.T) {
	h := NewAuthorHandler(newFakeRepo())

	got, err := h.CreateAuthor(context.Background(), &authorpb.CreateAuthorRequest{Name: "Jane Doe", Biography: "Writer"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != 1 || got.Name != "Jane Doe" {
		t.Fatalf("unexpected author: %+v", got)
	}
}

func TestCreateAuthorRequiresName(t *testing.T) {
	h := NewAuthorHandler(newFakeRepo())

	_, err := h.CreateAuthor(context.Background(), &authorpb.CreateAuthorRequest{Name: "  "})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("want InvalidArgument, got %v", err)
	}
}

func TestCreateAuthorHidesDatabaseError(t *testing.T) {
	repo := newFakeRepo()
	repo.err = errors.New("connection refused")
	h := NewAuthorHandler(repo)

	_, err := h.CreateAuthor(context.Background(), &authorpb.CreateAuthorRequest{Name: "Jane"})
	if status.Code(err) != codes.Internal || status.Convert(err).Message() != "internal error" {
		t.Fatalf("want generic Internal, got %v", err)
	}
}

func TestGetAuthorNotFound(t *testing.T) {
	h := NewAuthorHandler(newFakeRepo())

	_, err := h.GetAuthor(context.Background(), &authorpb.GetAuthorRequest{Id: 42})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("want NotFound, got %v", err)
	}
}

func TestUpdateAuthorKeepsUnsetFields(t *testing.T) {
	h := NewAuthorHandler(newFakeRepo())
	created, _ := h.CreateAuthor(context.Background(), &authorpb.CreateAuthorRequest{Name: "Jane", Biography: "Old bio"})

	got, err := h.UpdateAuthor(context.Background(), &authorpb.UpdateAuthorRequest{Id: created.Id, Name: "Jane Doe"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Jane Doe" || got.Biography != "Old bio" {
		t.Fatalf("unexpected author: %+v", got)
	}
}

func TestDeleteAuthor(t *testing.T) {
	h := NewAuthorHandler(newFakeRepo())
	created, _ := h.CreateAuthor(context.Background(), &authorpb.CreateAuthorRequest{Name: "Jane"})

	if _, err := h.DeleteAuthor(context.Background(), &authorpb.DeleteAuthorRequest{Id: created.Id}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := h.DeleteAuthor(context.Background(), &authorpb.DeleteAuthorRequest{Id: created.Id}); status.Code(err) != codes.NotFound {
		t.Fatalf("second delete: want NotFound, got %v", err)
	}
}
