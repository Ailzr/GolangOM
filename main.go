package main

import (
	_ "GolangOM/config"
	"GolangOM/logs"
	"GolangOM/pkg"
	"go.uber.org/zap"
)

func main() {
	// 确保日志文件关闭
	defer logs.Logger.Sync()

	logs.Logger.Info("Starting...")

	err := pkg.NewConnection(&pkg.ServerConfig{
		IP:         "192.168.1.1",
		Port:       22,
		User:       "root",
		AuthMethod: pkg.AuthMethodPassword,
		Password:   "password",
	})
	if err != nil {
		logs.Logger.Error("NewConnection error", zap.Error(err))
	}

	logs.Logger.Info("Exit Successfully")
}
