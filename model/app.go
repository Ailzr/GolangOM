package model

import (
	"GolangOM/constant"
	"GolangOM/database"
	"gorm.io/gorm"
)

type AppModel struct {
	gorm.Model
	ServerID      uint
	Name          string                `gorm:"type:varchar(255)"`
	CheckType     constant.AppCheckType `gorm:"type:varchar(255)"` // pid, port, http
	CheckTarget   string                `gorm:"type:varchar(255)"` // 如进程名、端口号、URL
	CheckInterval int                   `gorm:"type:int"`          // 检查间隔（秒）
	StartScript   string                `gorm:"type:varchar(255)"` // 启动脚本路径
	AutoRestart   bool                  // 是否自动重启
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
