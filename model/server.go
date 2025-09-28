package model

import (
	"GolangOM/constant"
	"GolangOM/database"

	"gorm.io/gorm"
)

type ServerModel struct {
	gorm.Model
	IP         string              `gorm:"type:varchar(255)" json:"ip"`
	Port       int                 `gorm:"type:int" json:"port"`
	User       string              `gorm:"type:varchar(255)" json:"user"`
	AuthMethod constant.AuthMethod `gorm:"type:varchar(255)" json:"auth_method"` // password or key
	Credential string              `gorm:"type:varchar(255)" json:"credential"`  // key path
	Password   string              `gorm:"type:varchar(255)" json:"password"`    // password or key password
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

func (s *ServerModel) DeleteServer() error {
	return database.DB.Delete(s).Error
}

func GetServerList() ([]ServerModel, error) {
	var servers []ServerModel
	err := database.DB.Find(&servers).Error
	return servers, err
}
