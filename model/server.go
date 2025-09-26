package model

import (
	"GolangOM/constant"
	"GolangOM/database"
	"gorm.io/gorm"
)

type ServerModel struct {
	gorm.Model
	IP         string              `gorm:"type:varchar(255)"`
	Port       int                 `gorm:"type:int"`
	User       string              `gorm:"type:varchar(255)"`
	AuthMethod constant.AuthMethod `gorm:"type:varchar(255)"` // password 或 key
	Credential string              `gorm:"type:varchar(255)"` // 密钥路径
	Password   string              `gorm:"type:varchar(255)"` // 密码 或 密钥的密码
}

func (s *ServerModel) IsExists() bool {
	return database.DB.Where("id = ?", s.ID).First(s).Error == nil
}

func (s *ServerModel) CreateServer() error {
	return database.DB.Create(s).Error
}

func (s *ServerModel) UpdateServer() error {
	return database.DB.Save(s).Error
}

func GetServerList() ([]ServerModel, error) {
	var servers []ServerModel
	err := database.DB.Find(&servers).Error
	return servers, err
}
