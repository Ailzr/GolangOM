package main

import (
	_ "GolangOM/config"
	"GolangOM/logs"
	"GolangOM/pkg"
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

	for times := 0; times < 3; times++ {
		test()
		time.Sleep(5 * time.Second)
	}

	logs.Logger.Info("Exit Successfully")
}

func test() {
	ids := pkg.GetConnectionPool().GetServerIDs()
	for _, id := range ids {
		server, err := pkg.GetConnectionPool().GetServerByID(id)
		if err != nil {
			logs.Logger.Error("GetServerByID error", zap.Error(err))
			continue
		}
		stdout, err := server.ExecuteCommand("echo 'hello world'")
		if err != nil {
			return
		}
		logs.Logger.Info("ExecuteCommand", zap.String("stdout", stdout))
	}
}
