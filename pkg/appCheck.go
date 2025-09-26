package pkg

import (
	"GolangOM/constant"
	"GolangOM/logs"
	"GolangOM/ws"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

type AppCheckConfig struct {
	ID              uint
	ServerID        uint
	Name            string
	CheckType       constant.AppCheckType // pid, port, http
	CheckTarget     string                // 如进程名、端口号、URL
	CheckInterval   int                   // 检查间隔（秒）
	StartScript     string                // 启动脚本路径
	LastCheckResult bool
	AutoRestart     bool // 是否自动重启
	LastCheckTime   time.Time
	ctx             context.Context
	cancel          context.CancelFunc
}

type AppCheckerManager struct {
	AppCheckerMap   map[uint]*AppCheckConfig
	AppCheckerMutex sync.RWMutex
}

var appCheckerManager = AppCheckerManager{
	AppCheckerMap:   make(map[uint]*AppCheckConfig),
	AppCheckerMutex: sync.RWMutex{},
}

func GetAppCheckerManager() *AppCheckerManager {
	return &appCheckerManager
}

func (a *AppCheckerManager) NewAppChecker(app *AppCheckConfig) error {
	a.AppCheckerMutex.Lock()
	defer a.AppCheckerMutex.Unlock()
	if a.AppCheckerMap[app.ID] != nil {
		return fmt.Errorf("app already exists")
	}
	a.AppCheckerMap[app.ID] = app
	app.StartAppChecker()
	return nil
}

func (a *AppCheckerManager) GetAppCheckerByID(appID uint) *AppCheckConfig {
	if _, ok := a.AppCheckerMap[appID]; !ok {
		return nil
	}
	return a.AppCheckerMap[appID]
}

func (a *AppCheckerManager) GetAppCheckers() map[uint]*AppCheckConfig {
	a.AppCheckerMutex.RLock()
	defer a.AppCheckerMutex.RUnlock()
	appCheckers := make(map[uint]*AppCheckConfig)
	for id, app := range a.AppCheckerMap {
		appCheckers[id] = app
	}
	return appCheckers
}

func (a *AppCheckerManager) RemoveAppCheckerByID(appID uint) {
	a.AppCheckerMutex.Lock()
	defer a.AppCheckerMutex.Unlock()
	if a.AppCheckerMap[appID] == nil {
		return
	}
	a.AppCheckerMap[appID].StopAppChecker()
	delete(a.AppCheckerMap, appID)
}

// 启动App检测器
func (app *AppCheckConfig) StartAppChecker() {
	if app.cancel != nil {
		app.cancel()
	}

	app.ctx, app.cancel = context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(time.Duration(app.CheckInterval) * time.Second)
		defer ticker.Stop()
		for {
			app.LastCheckTime = time.Now()
			if !app.CheckAppStatus() {
				app.LastCheckResult = false
				ws.SendMessage(ws.Message{
					AppID:     app.ID,
					AppStatus: app.LastCheckResult,
				})
				logs.Logger.Warn("App not running", zap.String("app", app.Name))
				// 如果设置为自动重启
				if app.AutoRestart {
					logs.Logger.Info("App restarting...", zap.String("app", app.Name))
					err := app.StartApp()
					if err != nil {
						logs.Logger.Error("App start error", zap.Error(err))
						continue
					}
					app.LastCheckResult = true
					ws.SendMessage(ws.Message{
						AppID:     app.ID,
						AppStatus: app.LastCheckResult,
					})
				}
			}
			time.Sleep(time.Duration(app.CheckInterval) * time.Second)
		}
	}()
}

func (app *AppCheckConfig) StopAppChecker() {
	if app.cancel != nil {
		app.cancel()
	}
}

func (app *AppCheckConfig) CheckAppStatus() bool {
	server := GetConnectionPool().GetServerByID(app.ServerID)
	if server == nil {
		logs.Logger.Error("GetServerByID error", zap.Error(errors.New("server not exists")), zap.String("server_id", strconv.Itoa(int(app.ServerID))))
		return false
	}
	if app.CheckType == constant.AppCheckTypePid {
		result, err := server.ExecuteCommand(fmt.Sprintf("ps -ef | grep %s | grep -v grep | awk '{print $2}'", app.CheckTarget))
		if err != nil {
			logs.Logger.Error("ExecuteCommand error", zap.Error(err))
			return false
		}
		logs.Logger.Debug("app check result:", zap.String("pid", result))
		return len(result) > 0
	} else if app.CheckType == constant.AppCheckTypePort {
		result, err := server.ExecuteCommand(fmt.Sprintf("lsof -i :%s | grep LISTEN | awk '{print $2}'", app.CheckTarget))
		if err != nil {
			logs.Logger.Error("ExecuteCommand error", zap.Error(err))
			return false
		}
		logs.Logger.Debug("app check result:", zap.String("pid", result))
		return len(result) > 0
	} else if app.CheckType == constant.AppCheckTypeHttp {
	}
	return false
}

func (app *AppCheckConfig) StartApp() error {
	server := GetConnectionPool().GetServerByID(app.ServerID)
	if server == nil {
		return fmt.Errorf("server not exists")
	}
	_, err := server.ExecuteCommand(app.StartScript)
	if err != nil {
		return err
	}
	return nil
}
