// Package handler implements the gRPC service defined in author.proto.
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

	"github.com/fnxr21/author-service/internal/model"
	"github.com/fnxr21/author-service/internal/repository"
	"github.com/fnxr21/author-service/protobuf/authorpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AuthorHandler implements authorpb.AuthorServiceServer.
type AuthorHandler struct {
	// Embedding the generated Unimplemented server makes any RPC we have not
	// written yet return codes.Unimplemented instead of failing to compile.
	authorpb.UnimplementedAuthorServiceServer
	repo repository.AuthorRepository
}

func NewAuthorHandler(repo repository.AuthorRepository) *AuthorHandler {
	return &AuthorHandler{repo: repo}
}

func (h *AuthorHandler) CreateAuthor(ctx context.Context, req *authorpb.CreateAuthorRequest) (*authorpb.Author, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	author, err := h.repo.Create(ctx, model.Author{
		Name:      req.Name,
		Biography: req.Biography,
	})
	if err != nil {
		return nil, toStatusError("CreateAuthor", err)
	}
	return toProto(author), nil
}

// ListAuthors sends each author on the stream. The REST gateway writes each
// one as a separate JSON line.
func (h *AuthorHandler) ListAuthors(_ *emptypb.Empty, stream authorpb.AuthorService_ListAuthorsServer) error {
	authors, err := h.repo.List(stream.Context())
	if err != nil {
		return toStatusError("ListAuthors", err)
	}
	for _, author := range authors {
		if err := stream.Send(toProto(author)); err != nil {
			return err
		}
	}
	return nil
}

func (h *AuthorHandler) GetAuthor(ctx context.Context, req *authorpb.GetAuthorRequest) (*authorpb.Author, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	author, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toStatusError("GetAuthor", err)
	}
	return toProto(author), nil
}

// UpdateAuthor changes only the fields that are set in the request.
// proto3 strings have no "unset" state, so an empty string means "keep".
func (h *AuthorHandler) UpdateAuthor(ctx context.Context, req *authorpb.UpdateAuthorRequest) (*authorpb.Author, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	author, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toStatusError("UpdateAuthor", err)
	}
	if req.Name != "" {
		author.Name = req.Name
	}
	if req.Biography != "" {
		author.Biography = req.Biography
	}

	author, err = h.repo.Update(ctx, author)
	if err != nil {
		return nil, toStatusError("UpdateAuthor", err)
	}
	return toProto(author), nil
}

func (h *AuthorHandler) DeleteAuthor(ctx context.Context, req *authorpb.DeleteAuthorRequest) (*authorpb.DeleteAuthorResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Load first so a missing author returns 404 instead of silently succeeding.
	author, err := h.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toStatusError("DeleteAuthor", err)
	}
	if err := h.repo.Delete(ctx, author); err != nil {
		return nil, toStatusError("DeleteAuthor", err)
	}
	return &authorpb.DeleteAuthorResponse{Message: fmt.Sprintf("author %d deleted", req.Id)}, nil
}

// toProto converts the database model to the API message.
func toProto(author model.Author) *authorpb.Author {
	return &authorpb.Author{
		Id:        uint64(author.ID),
		Name:      author.Name,
		Biography: author.Biography,
	}
}
