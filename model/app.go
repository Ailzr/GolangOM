package model

import (
	"GolangOM/constant"
	"gorm.io/gorm"
)

type AppModel struct {
	gorm.Model
	AppID         string
	ServerID      string
	Name          string
	CheckType     constant.AppCheckType // pid, port, http
	CheckTarget   string                // 如进程名、端口号、URL
	CheckInterval int                   // 检查间隔（秒）
	StartScript   string                // 启动脚本路径
	AutoRestart   bool                  // 是否自动重启
}
