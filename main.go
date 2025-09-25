package main

import (
	_ "GolangOM/config"
	"GolangOM/logs"
	"GolangOM/router"
)

func main() {
	// 确保日志文件关闭
	defer logs.Logger.Sync()

	router.InitRouter()

	logs.Logger.Info("Exit Successfully")
}
