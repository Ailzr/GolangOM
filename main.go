package main

import (
	_ "GolangOM/config"
	"GolangOM/logs"
	"GolangOM/pkg"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

func main() {
	// 确保日志文件关闭
	defer logs.Logger.Sync()

	logs.Logger.Info("Starting...")

	err := pkg.NewConnection(&pkg.ServerConfig{
		IP:         "192.168.80.128",
		Port:       22,
		User:       "ailzr",
		AuthMethod: pkg.AuthMethodPassword,
		Password:   "Dd159753.",
	})
	if err != nil {
		logs.Logger.Error("NewConnection error", zap.Error(err))
	}

	ids := pkg.GetConnectionPool().GetServerIDs()

	appConfig := pkg.AppCheckConfig{
		ID:            uuid.New().String(),
		Name:          "test",
		ServerID:      ids[0],
		CheckTarget:   "server-linux",
		StartScript:   "/home/ailzr/test.sh",
		CheckInterval: 5,
		AutoRestart:   true,
	}

	err = appConfig.StartApp()
	if err != nil {
		logs.Logger.Error("StartApp error", zap.Error(err))
		return
	}

	for times := 0; times < 3; times++ {
		if appConfig.CheckAppStatus() {
			logs.Logger.Debug("App is running")
		} else {
			logs.Logger.Debug("App is not running")
		}
		time.Sleep(3 * time.Second)
	}

	logs.Logger.Info("Exit Successfully")
}
