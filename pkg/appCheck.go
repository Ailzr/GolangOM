package pkg

import (
	"GolangOM/logs"
	"fmt"
	"go.uber.org/zap"
	"time"
)

type AppCheckConfig struct {
	ID              string
	ServerID        string
	Name            string
	CheckType       string // pid, port, http
	CheckTarget     string // 如进程名、端口号、URL
	CheckInterval   int    // 检查间隔（秒）
	StartScript     string // 启动脚本路径
	LastCheckResult bool
	AutoRestart     bool // 是否自动重启
	LastCheckTime   time.Time
}

func (app *AppCheckConfig) CheckAppStatus() bool {
	server, err := GetConnectionPool().GetServerByID(app.ServerID)
	if err != nil {
		logs.Logger.Error("GetServerByID error", zap.Error(err))
		return false
	}
	result, err := server.ExecuteCommand(fmt.Sprintf("ps -ef | grep %s | grep -v grep | awk '{print $2}'", app.CheckTarget))
	if err != nil {
		logs.Logger.Error("ExecuteCommand error", zap.Error(err))
		return false
	}
	logs.Logger.Debug("app check result:", zap.String("pid", result))
	return len(result) > 0
}

func (app *AppCheckConfig) StartApp() error {
	server, err := GetConnectionPool().GetServerByID(app.ServerID)
	if err != nil {
		return err
	}
	_, err = server.ExecuteCommand(app.StartScript)
	if err != nil {
		return err
	}
	return nil
}
