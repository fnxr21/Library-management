package handler

import (
	"context"
	"github.com/fnxr21/book-service/internal/model"
	"github.com/fnxr21/book-service/internal/repository"
	"github.com/fnxr21/book-service/protobuf/protobuf_book"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log"
)

type bookServices struct {
	protobuf_book.UnimplementedBookServiceServer
	repository repository.BookService
}

func HandlerBook(BookRepository repository.BookService) *bookServices {
	return &bookServices{repository: BookRepository}
}
func (r *bookServices) CreateBook(ctx context.Context, req *protobuf_book.ProtoBookRepo_ProtoBook) (*protobuf_book.ProtoBookRepo_ProtoBook, error) {
	ok := model.Book{
		Name:        req.Name,
		Description: req.Description,
	}

	log.Println(ok, "check")

	ok, err := r.repository.CreateBook(ctx, ok)
	if err != nil {
		log.Println("[ERROR] : ", err.Error())
		return nil, err
	}

	return ToProtoBook(ok), nil
}
func (r *bookServices) GetBookById(ctx context.Context, id *wrapperspb.Int64Value) (*protobuf_book.ProtoBookRepo_ProtoBook, error) {

	data, err := r.repository.GetBookById(ctx, int(id.Value))
	if err != nil {
		log.Println("[ERROR] : ", err.Error())
		return nil, err
	}
	return ToProtoBook(data), nil
}
func (r *bookServices) GetBookList(_ *emptypb.Empty, stream protobuf_book.BookService_GetBookListServer) error {

	book, err := r.repository.GetBookList()
	if err != nil {
		log.Println("[ERROR] : ", err.Error())
		return err
	}
	for _, value := range book {
		if err := stream.Send(ToProtoBook(value)); err != nil {
			return err
		}
	}
	return nil

}

func ToProtoBook(book model.Book) *protobuf_book.ProtoBookRepo_ProtoBook {
	return &protobuf_book.ProtoBookRepo_ProtoBook{
		ID: uint64(book.ID), Name: book.Name, Description: book.Description,
	}
}
