package handler

import (
	"context"
	"log"

	"github.com/fnxr21/author-service/internal/model"
	"github.com/fnxr21/author-service/internal/repository"
	"github.com/fnxr21/author-service/protobuf/protobuf_author"
	// "google.golang.org/protobuf/types/known/wrapperspb"
)

type authorServices struct {
	protobuf_author.UnimplementedAuthorServiceServer
	repository repository.AuthorService
}

func HandlerAuthor(AuthorRepository repository.AuthorService) *authorServices {
	return &authorServices{repository: AuthorRepository}
}
func (r *authorServices) CreateAuthor(ctx context.Context, req *protobuf_author.ProtoAuthorRepo_ProtoAuthor) (*protobuf_author.ProtoAuthorRepo_ProtoAuthor, error) {
	ok := model.Author{
		Name:      req.Name,
		Biography: req.Biography,
	}

	log.Println(ok, "check")

	ok, err := r.repository.CreateAuthor(ctx, ok)
	if err != nil {
		log.Println("[ERROR] : ", err.Error())
		return nil, err
	}

	return ToProtoAuthor(ok), nil
}
// func (r *authorServices) GetAuthorById(ctx context.Context, id *wrapperspb.Int64Value) (*protobuf_author.ProtoAuthorRepo_ProtoAuthor, error) {
// 	if id == nil {
// 		// Handle the nil case appropriately
// 		return 0
// 	}
// 	r.repository.GetAuthorById(ctx, int(id.Value))

// 	return ToProtoAuthor(ok), nil
// }

func ToProtoAuthor(author model.Author) *protobuf_author.ProtoAuthorRepo_ProtoAuthor {
	return &protobuf_author.ProtoAuthorRepo_ProtoAuthor{
		ID: uint64(author.ID), Name: author.Name, Biography: author.Biography,
	}
}
