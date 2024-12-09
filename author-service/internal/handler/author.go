package handler

import (
	"context"
	"log"
	"github.com/fnxr21/author-service/internal/model"
	"github.com/fnxr21/author-service/protobuf/protobuf_author"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"github.com/fnxr21/author-service/internal/repository"
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
func (r *authorServices) GetAuthorById(ctx context.Context, id *wrapperspb.Int64Value) (*protobuf_author.ProtoAuthorRepo_ProtoAuthor, error) {

	data, err := r.repository.GetAuthorById(ctx, int(id.Value))
	if err != nil {
		log.Println("[ERROR] : ", err.Error())
		return nil, err
	}
	return ToProtoAuthor(data), nil
}
func (r *authorServices) GetAuthorList(_ *emptypb.Empty, stream protobuf_author.AuthorService_GetAuthorListServer) error {

	author, err := r.repository.GetAuthorList()
	if err != nil {
		log.Println("[ERROR] : ", err.Error())
		return err
	}
	for _, value := range author {
		if err := stream.Send(ToProtoAuthor(value)); err != nil {
			return err
		}
	}
	return nil

}

func ToProtoAuthor(author model.Author) *protobuf_author.ProtoAuthorRepo_ProtoAuthor {
	return &protobuf_author.ProtoAuthorRepo_ProtoAuthor{
		ID: uint64(author.ID), Name: author.Name, Biography: author.Biography,
	}
}
