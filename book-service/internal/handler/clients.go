package handler

import (
	"context"
	"log"
	"time"

	"github.com/fnxr21/book-service/internal/model"
	"github.com/fnxr21/book-service/protobuf/authorpb"
	"github.com/fnxr21/book-service/protobuf/categorypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// callTimeout bounds each call to author-service or category-service, so a
// slow or hung service cannot make book requests hang forever.
const callTimeout = 3 * time.Second

// lookupAuthor is service-to-service communication over gRPC: book-service
// asks author-service for the author and keeps a copy of the name.
// It returns InvalidArgument if the author does not exist.
func (h *BookHandler) lookupAuthor(ctx context.Context, id uint64) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	author, err := h.authors.GetAuthor(ctx, &authorpb.GetAuthorRequest{Id: id})
	if err != nil {
		return "", downstreamError("author", id, err)
	}
	return author.Name, nil
}

// lookupCategories resolves category IDs to links, skipping duplicates.
func (h *BookHandler) lookupCategories(ctx context.Context, ids []uint64) ([]model.BookCategory, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	seen := make(map[uint64]bool, len(ids))
	links := make([]model.BookCategory, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true

		category, err := h.categories.GetCategory(ctx, &categorypb.GetCategoryRequest{Id: id})
		if err != nil {
			return nil, downstreamError("category", id, err)
		}
		links = append(links, model.BookCategory{CategoryID: uint(id), CategoryName: category.Name})
	}
	return links, nil
}

// downstreamError turns a failed lookup into a status the client can act on:
//
//	not found in the other service -> InvalidArgument (400): fix your request
//	other service down or timed out -> Unavailable (503): try again later
func downstreamError(kind string, id uint64, err error) error {
	switch status.Code(err) {
	case codes.NotFound, codes.InvalidArgument:
		return status.Errorf(codes.InvalidArgument, "%s %d does not exist", kind, id)
	default:
		log.Printf("[ERROR] %s-service lookup %d: %v", kind, id, err)
		return status.Errorf(codes.Unavailable, "%s-service is unavailable", kind)
	}
}
