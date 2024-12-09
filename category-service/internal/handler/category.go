package handler

import (
	"context"
	"github.com/fnxr21/category-service/internal/model"
	"github.com/fnxr21/category-service/internal/repository"
	"github.com/fnxr21/category-service/protobuf/protobuf_category"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log"
)

type categoryServices struct {
	protobuf_category.UnimplementedCategoryServiceServer
	repository repository.CategoryService
}

func HandlerCategory(CategoryRepository repository.CategoryService) *categoryServices {
	return &categoryServices{repository: CategoryRepository}
}
func (r *categoryServices) CreateCategory(ctx context.Context, req *protobuf_category.ProtoCategoryRepo_ProtoCategory) (*protobuf_category.ProtoCategoryRepo_ProtoCategory, error) {
	ok := model.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	log.Println(ok, "check")

	ok, err := r.repository.CreateCategory(ctx, ok)
	if err != nil {
		log.Println("[ERROR] : ", err.Error())
		return nil, err
	}

	return ToProtoCategory(ok), nil
}
func (r *categoryServices) GetCategoryById(ctx context.Context, id *wrapperspb.Int64Value) (*protobuf_category.ProtoCategoryRepo_ProtoCategory, error) {

	data, err := r.repository.GetCategoryById(ctx, int(id.Value))
	if err != nil {
		log.Println("[ERROR] : ", err.Error())
		return nil, err
	}
	return ToProtoCategory(data), nil
}
func (r *categoryServices) GetCategoryList(_ *emptypb.Empty, stream protobuf_category.CategoryService_GetCategoryListServer) error {

	category, err := r.repository.GetCategoryList()
	if err != nil {
		log.Println("[ERROR] : ", err.Error())
		return err
	}
	for _, value := range category {
		if err := stream.Send(ToProtoCategory(value)); err != nil {
			return err
		}
	}
	return nil

}

func ToProtoCategory(category model.Category) *protobuf_category.ProtoCategoryRepo_ProtoCategory {
	return &protobuf_category.ProtoCategoryRepo_ProtoCategory{
		ID: uint64(category.ID), Name: category.Name, Description: category.Description,
	}
}
