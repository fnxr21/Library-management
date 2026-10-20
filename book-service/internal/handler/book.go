// Package handler implements the gRPC service defined in book.proto.
// See author-service's handler package for the general pattern.
package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/fnxr21/book-service/internal/model"
	"github.com/fnxr21/book-service/internal/repository"
	"github.com/fnxr21/book-service/protobuf/authorpb"
	"github.com/fnxr21/book-service/protobuf/bookpb"
	"github.com/fnxr21/book-service/protobuf/categorypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// BookHandler implements bookpb.BookServiceServer. Besides its own
// repository it holds gRPC clients for author-service and category-service,
// which it uses to check that referenced authors and categories exist.
type BookHandler struct {
	bookpb.UnimplementedBookServiceServer
	repo       repository.BookRepository
	authors    authorpb.AuthorServiceClient
	categories categorypb.CategoryServiceClient
}

func NewBookHandler(repo repository.BookRepository, authors authorpb.AuthorServiceClient, categories categorypb.CategoryServiceClient) *BookHandler {
	return &BookHandler{repo: repo, authors: authors, categories: categories}
}

func (h *BookHandler) CreateBook(ctx context.Context, req *bookpb.CreateBookRequest) (*bookpb.Book, error) {
	switch {
	case strings.TrimSpace(req.Name) == "":
		return nil, status.Error(codes.InvalidArgument, "name is required")
	case req.AuthorId == 0:
		return nil, status.Error(codes.InvalidArgument, "authorId is required")
	case req.Stock < 0:
		return nil, status.Error(codes.InvalidArgument, "stock cannot be negative")
	}

	// Ask the other services before saving anything (see clients.go).
	authorName, err := h.lookupAuthor(ctx, req.AuthorId)
	if err != nil {
		return nil, err
	}
	categories, err := h.lookupCategories(ctx, req.CategoryIds)
	if err != nil {
		return nil, err
	}

	book, err := h.repo.Create(ctx, model.Book{
		Name:        req.Name,
		Description: req.Description,
		AuthorID:    uint(req.AuthorId),
		AuthorName:  authorName,
		Stock:       req.Stock,
		Categories:  categories,
	})
	if err != nil {
		return nil, toStatusError("CreateBook", err)
	}
	return toProto(book), nil
}

func (h *BookHandler) ListBooks(_ *emptypb.Empty, stream bookpb.BookService_ListBooksServer) error {
	books, err := h.repo.List(stream.Context())
	if err != nil {
		return toStatusError("ListBooks", err)
	}
	for _, book := range books {
		if err := stream.Send(toProto(book)); err != nil {
			return err
		}
	}
	return nil
}

func (h *BookHandler) GetBook(ctx context.Context, req *bookpb.GetBookRequest) (*bookpb.Book, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	book, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toStatusError("GetBook", err)
	}
	return toProto(book), nil
}

// UpdateBook changes only the fields that are set in the request.
// A non-empty categoryIds list replaces the book's categories.
func (h *BookHandler) UpdateBook(ctx context.Context, req *bookpb.UpdateBookRequest) (*bookpb.Book, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	book, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toStatusError("UpdateBook", err)
	}
	if req.Name != "" {
		book.Name = req.Name
	}
	if req.Description != "" {
		book.Description = req.Description
	}
	if req.AuthorId != 0 && uint(req.AuthorId) != book.AuthorID {
		if book.AuthorName, err = h.lookupAuthor(ctx, req.AuthorId); err != nil {
			return nil, err
		}
		book.AuthorID = uint(req.AuthorId)
	}

	// nil means "keep the current categories" (see repository.Update).
	var categories []model.BookCategory
	if len(req.CategoryIds) > 0 {
		if categories, err = h.lookupCategories(ctx, req.CategoryIds); err != nil {
			return nil, err
		}
	}

	book, err = h.repo.Update(ctx, book, categories)
	if err != nil {
		return nil, toStatusError("UpdateBook", err)
	}
	return toProto(book), nil
}

// AdjustStock is the only way to change stock. The race-safe logic lives in
// repository.AdjustStock.
func (h *BookHandler) AdjustStock(ctx context.Context, req *bookpb.AdjustStockRequest) (*bookpb.Book, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Delta == 0 {
		return nil, status.Error(codes.InvalidArgument, "delta must not be zero")
	}

	book, err := h.repo.AdjustStock(ctx, req.Id, req.Delta)
	if err != nil {
		return nil, toStatusError("AdjustStock", err)
	}
	return toProto(book), nil
}

func (h *BookHandler) DeleteBook(ctx context.Context, req *bookpb.DeleteBookRequest) (*bookpb.DeleteBookResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	book, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toStatusError("DeleteBook", err)
	}
	if err := h.repo.Delete(ctx, book); err != nil {
		return nil, toStatusError("DeleteBook", err)
	}
	return &bookpb.DeleteBookResponse{Message: fmt.Sprintf("book %d deleted", req.Id)}, nil
}

// toProto converts the database model to the API message.
func toProto(book model.Book) *bookpb.Book {
	categories := make([]*bookpb.BookCategory, 0, len(book.Categories))
	for _, c := range book.Categories {
		categories = append(categories, &bookpb.BookCategory{Id: uint64(c.CategoryID), Name: c.CategoryName})
	}
	return &bookpb.Book{
		Id:          uint64(book.ID),
		Name:        book.Name,
		Description: book.Description,
		AuthorId:    uint64(book.AuthorID),
		AuthorName:  book.AuthorName,
		Stock:       book.Stock,
		Categories:  categories,
	}
}
