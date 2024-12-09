package handler

import (
	"context"

	"github.com/fnxr21/book-service/internal/model"
	"github.com/fnxr21/book-service/internal/repository"
	// "github.com/fnxr21/book-service/internal/handler/category"
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
	book := model.Book{
		Name:        req.Name,
		Description: req.Description,
		AuthorId:    uint(req.AuthorId),
		Author:      req.Author,
		Stock:       req.Stock,
	}

	ds, _ := GetCategories()
	// if err != nil {
		// log.Println("[ERROR] : Failed to fetch categories", err)
		// 	return nil, err
		// }
		log.Println(ds,"check this out")


	// book, err := r.repository.CreateBook(ctx, ok)
	// if err != nil {
	// 	log.Println("[ERROR] : ", err.Error())
	// 	return nil, err
	// }

	return ToProtoBook(book), nil
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

	var categoryNames []string

	// Iterate over the BookCategory slice and convert to ProtoCategory
	for _, bc := range book.Categories {
		categoryNames = append(categoryNames, bc.CategoryName) // Assuming Category has a Name field
	}
	return &protobuf_book.ProtoBookRepo_ProtoBook{
		ID: uint64(book.ID), Name: book.Name, Description: book.Description, AuthorId: uint64(book.AuthorId), Author: book.Author, Stock: int32(book.Stock), Categories: categoryNames,
	}
}
