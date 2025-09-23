package pkg

import "time"

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
