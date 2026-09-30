package model

import (
	"gorm.io/gorm"
)

type UserSQL struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;size:50"`
	Password string
	Role     string
}

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)
