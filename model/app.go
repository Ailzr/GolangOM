package model

import (
	"GolangOM/constant"
	"GolangOM/database"

	"gorm.io/gorm"
)

type AppModel struct {
	gorm.Model
	ServerID      uint                  `json:"server_id"`
	Name          string                `gorm:"type:varchar(255)" json:"name"`
	CheckType     constant.AppCheckType `gorm:"type:varchar(255)" json:"check_type"`   // pid, port, http
	CheckTarget   string                `gorm:"type:varchar(255)" json:"check_target"` // such as process name, port number, URL
	CheckInterval int                   `gorm:"type:int" json:"check_interval"`        // check interval (seconds)
	StartScript   string                `gorm:"type:varchar(255)" json:"start_script"` // startup script path
	AutoRestart   bool                  `json:"auto_restart"`                          // whether to auto restart
	Server        ServerModel           `gorm:"foreignKey:ServerID"`
}

func (a *AppModel) IsExists() bool {
	return database.DB.Where("id = ?", a.ID).First(a).Error == nil
}

func (a *AppModel) CreateApp() error {
	return database.DB.Create(a).Error
}

func (a *AppModel) UpdateApp() error {
	return database.DB.Save(a).Error
}

func (a *AppModel) DeleteApp() error {
	return database.DB.Delete(a).Error
}

func GetAppList() ([]AppModel, error) {
	var apps []AppModel
	err := database.DB.Preload("Server").Find(&apps).Error
	return apps, err
}
