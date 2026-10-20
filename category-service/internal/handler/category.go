// Package handler implements the gRPC service defined in category.proto.
//
// A handler's job, in order: validate the request, call the repository,
// convert the result to a protobuf message. Errors are returned as gRPC
// status codes (codes.NotFound, codes.InvalidArgument, ...), which the REST
// gateway turns into HTTP status codes (404, 400, ...).
package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/fnxr21/category-service/internal/model"
	"github.com/fnxr21/category-service/internal/repository"
	"github.com/fnxr21/category-service/protobuf/categorypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CategoryHandler implements categorypb.CategoryServiceServer.
type CategoryHandler struct {
	// Embedding the generated Unimplemented server makes any RPC we have not
	// written yet return codes.Unimplemented instead of failing to compile.
	categorypb.UnimplementedCategoryServiceServer
	repo repository.CategoryRepository
}

func NewCategoryHandler(repo repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

func (h *CategoryHandler) CreateCategory(ctx context.Context, req *categorypb.CreateCategoryRequest) (*categorypb.Category, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	category, err := h.repo.Create(ctx, model.Category{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, toStatusError("CreateCategory", err)
	}
	return toProto(category), nil
}

// ListCategories sends each category on the stream. The REST gateway writes each
// one as a separate JSON line.
func (h *CategoryHandler) ListCategories(_ *emptypb.Empty, stream categorypb.CategoryService_ListCategoriesServer) error {
	categories, err := h.repo.List(stream.Context())
	if err != nil {
		return toStatusError("ListCategories", err)
	}
	for _, category := range categories {
		if err := stream.Send(toProto(category)); err != nil {
			return err
		}
	}
	return nil
}

func (h *CategoryHandler) GetCategory(ctx context.Context, req *categorypb.GetCategoryRequest) (*categorypb.Category, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	category, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toStatusError("GetCategory", err)
	}
	return toProto(category), nil
}

// UpdateCategory changes only the fields that are set in the request.
// proto3 strings have no "unset" state, so an empty string means "keep".
func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *categorypb.UpdateCategoryRequest) (*categorypb.Category, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	category, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toStatusError("UpdateCategory", err)
	}
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}

	category, err = h.repo.Update(ctx, category)
	if err != nil {
		return nil, toStatusError("UpdateCategory", err)
	}
	return toProto(category), nil
}

func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *categorypb.DeleteCategoryRequest) (*categorypb.DeleteCategoryResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Load first so a missing category returns 404 instead of silently succeeding.
	category, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toStatusError("DeleteCategory", err)
	}
	if err := h.repo.Delete(ctx, category); err != nil {
		return nil, toStatusError("DeleteCategory", err)
	}
	return &categorypb.DeleteCategoryResponse{Message: fmt.Sprintf("category %d deleted", req.Id)}, nil
}

// toProto converts the database model to the API message.
func toProto(category model.Category) *categorypb.Category {
	return &categorypb.Category{
		Id:          uint64(category.ID),
		Name:        category.Name,
		Description: category.Description,
	}
}
