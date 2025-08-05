package service

import (
	"fmt"
	"tinyGW/app/api/repository"
	"tinyGW/app/api/schemas/req"
	"tinyGW/app/models"
	"tinyGW/pkg/plugin/response"
)

type UserService interface {
	Add(user *req.UserReq) error
	Update(user *req.UserReq) error
	Delete(name string) error
	Find(name string) (req.UserReq, error)
	FindAll() ([]req.UserReq, error)
	List(page *req.PageReq) (response.PageResp, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func (d userService) Add(user *req.UserReq) error {
	_, err := d.userRepo.Find(user.Username)
	if err == nil {
		return fmt.Errorf("用户 %s <UNK>", user.Username)
	}
	var dv models.User
	response.Copy(&dv, user)
	return d.userRepo.Save(&dv)
}

func (d userService) Update(user *req.UserReq) error {
	var dv models.User
	response.Copy(&dv, user)
	return d.userRepo.Save(&dv)
}

func (d userService) Delete(name string) error {
	return d.userRepo.Delete(name)
}

func (d userService) Find(name string) (req.UserReq, error) {
	f, err := d.userRepo.Find(name)
	if err != nil {
		return req.UserReq{}, fmt.Errorf("用户 %s 不存在", name)
	}
	var user req.UserReq
	response.Copy(&user, f)
	return user, nil
}

func (d userService) FindAll() ([]req.UserReq, error) {
	fs, err := d.userRepo.FindAll()
	if err != nil {
		return nil, err
	}
	var users []req.UserReq
	response.Copy(&users, fs)
	return users, nil
}

func (d userService) List(page *req.PageReq) (response.PageResp, error) {
	limit := page.PageSize
	offset := page.PageSize * (page.PageNo - 1)
	list, total, err := d.userRepo.List(limit, offset)
	if err != nil {
		return response.PageResp{}, err
	}
	var users []req.UserReq
	response.Copy(&users, list)
	return response.PageResp{
		Count:    total,
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
		Lists:    list,
	}, nil
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}
