package service

import (
	"TinyGW/internal/repository"
	"TinyGW/models"
)

type UserService interface {
	Save(user *models.User) error
	Delete(name string) error
	Find(name string) (models.User, error)
	FindAll() ([]models.User, error)
	List(limit int, offset int) ([]models.User, int64, error)
}

type userService struct {
	repo repository.UserRepository
}

func (s *userService) Save(user *models.User) error {
	return s.repo.Save(user)
}
func (s *userService) Delete(name string) error {
	return s.repo.Delete(name)
}
func (s *userService) Find(name string) (models.User, error) {
	return s.repo.Find(name)
}
func (s *userService) FindAll() ([]models.User, error) {
	return s.repo.FindAll()
}
func (s *userService) List(limit int, offset int) ([]models.User, int64, error) {
	return s.repo.List(limit, offset)
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}
