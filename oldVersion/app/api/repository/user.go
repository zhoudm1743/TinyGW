package repository

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"tinyGW/app/models"
)

type UserRepository interface {
	Save(user *models.User) error
	Delete(name string) error
	Find(name string) (models.User, error)
	FindAll() ([]models.User, error)
	List(limit int, offset int) ([]models.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func (c userRepository) Save(user *models.User) error {
	var err error
	var task models.User
	if err = c.db.Where("name = ?", user.Name).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = c.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				UpdateAll: true,
			}).Create(user).Error
			return err
		}
	}
	err = c.db.Model(&task).Save(user).Error
	return err
}

func (c userRepository) Delete(name string) error {
	return c.db.Delete(&models.User{}, "username = ?", name).Error
}

func (c userRepository) Find(name string) (models.User, error) {
	var user models.User
	err := c.db.Where("name = ?", name).First(&user).Error
	return user, err
}

func (c userRepository) FindAll() ([]models.User, error) {
	var users []models.User
	err := c.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (c userRepository) List(limit int, offset int) ([]models.User, int64, error) {
	var users []models.User
	var count int64
	c.db.Model(&models.User{}).Count(&count)
	err := c.db.Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, count, nil
}

func NewUserRepository(db *gorm.DB) UserRepository {
	db.AutoMigrate(&models.User{})
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 {
		db.Create(&models.User{
			Name:     "admin",
			Password: "XAKJ8808856",
		})
	}
	return &userRepository{
		db: db,
	}
}
