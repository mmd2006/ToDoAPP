package repository

import (
	"ToDoAPP/config"
	"ToDoAPP/model"
)

type UserRepository interface {
	Create(user *model.UserSQL) error
	FindByUsername(username string) (*model.UserSQL, error)
	FindAll() ([]model.UserSQL, error)
}

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) Create(user *model.UserSQL) error {
	return config.DB.Create(user).Error
}

func (r *userRepository) FindByUsername(username string) (*model.UserSQL, error) {
	var user model.UserSQL
	err := config.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll() ([]model.UserSQL, error) {
	var users []model.UserSQL
	if err := config.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
