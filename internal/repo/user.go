package repo

import (
	"TinyGW/models"
	"errors"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	db.AutoMigrate(&models.User{})
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 {
		db.Create(&models.User{Name: "admin", Password: "123456"})
	}
	return &UserRepo{db: db}
}

// 创建用户
func (r *UserRepo) Create(user *models.User) error {
	var count int64
	r.db.Model(&models.User{}).Where("name = ?", user.Name).Count(&count)
	if count > 0 {
		return errors.New("用户已存在")
	}
	return r.db.Create(user).Error
}

// 更新用户
func (r *UserRepo) Update(user *models.User) error {
	var count int64
	r.db.Model(&models.User{}).Where("name = ? and id != ?", user.Name, user.ID).Count(&count)
	if count == 0 {
		return errors.New("用户不存在")
	}
	return r.db.Save(user).Error
}

// 删除用户
func (r *UserRepo) Delete(user *models.User) error {
	return r.db.Delete(user).Error
}

// 获取用户
func (r *UserRepo) Get(name string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("name = ?", name).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 获取用户
func (r *UserRepo) Find(page, pageSize int, conds map[string]interface{}) (int64, []models.User, error) {
	var users []models.User
	db := r.db.Model(&models.User{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	var count int64
	db.Count(&count)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)
	return count, users, nil
}

// 获取用户数量
func (r *UserRepo) Count(conds map[string]interface{}) (int64, error) {
	var count int64
	db := r.db.Model(&models.User{})
	for k, v := range conds {
		db = db.Where(k, v)
	}
	db.Count(&count)
	return count, nil
}

// 获取所有用户
func (r *UserRepo) GetAll() ([]models.User, error) {
	var users []models.User
	r.db.Find(&users)
	return users, nil
}
