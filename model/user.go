package model

import (
	"GolangOM/database"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"type:varchar(63)"`
	Password string `json:"password" gorm:"type:varchar(255)"`
	Role     string `json:"role" gorm:"type:varchar(31)"`
}

func (u *User) IsExists() bool {
	return database.DB.Where("username = ?", u.Username).First(u).Error == nil
}

func (u *User) CreateUser() error {
	return database.DB.Create(u).Error
}

func (u *User) UpdateUser() error {
	return database.DB.Save(u).Error
}
