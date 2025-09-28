package pkg

import (
	"GolangOM/constant"
	"GolangOM/logs"
	"GolangOM/ws"
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type AppCheckConfig struct {
	ID              uint
	ServerID        uint
	Name            string
	CheckType       constant.AppCheckType // pid, port, http
	CheckTarget     string                // such as process name, port number, URL
	CheckInterval   int                   // check interval (seconds)
	StartScript     string                // startup script path
	LastCheckResult bool
	AutoRestart     bool // whether to auto restart
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

func (a *AppCheckerManager) GetAppCheckers() []*AppCheckConfig {
	a.AppCheckerMutex.RLock()
	defer a.AppCheckerMutex.RUnlock()
	appCheckers := make([]*AppCheckConfig, 0, len(a.AppCheckerMap))
	for _, app := range a.AppCheckerMap {
		appCheckers = append(appCheckers, app)
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

// start App checker
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
			isRunning := app.CheckAppStatus()

			if !isRunning {
				app.LastCheckResult = false
				ws.SendMessage(ws.Message{
					AppID:     app.ID,
					AppStatus: app.LastCheckResult,
				})
				logs.Logger.Warn("App not running", zap.String("app", app.Name))
				// if auto restart is enabled
				if app.AutoRestart {
					logs.Logger.Info("App restarting...", zap.String("app", app.Name))
					err := app.StartApp()
					if err != nil {
						logs.Logger.Error("App start error", zap.Error(err))
						time.Sleep(time.Duration(app.CheckInterval) * time.Second)
						continue
					}
					app.LastCheckResult = true
					ws.SendMessage(ws.Message{
						AppID:     app.ID,
						AppStatus: app.LastCheckResult,
					})
				}
			} else {
				// also update status when app is running normally
				if !app.LastCheckResult {
					app.LastCheckResult = true
					ws.SendMessage(ws.Message{
						AppID:     app.ID,
						AppStatus: app.LastCheckResult,
					})
				}
			}

			select {
			case <-app.ctx.Done():
				return
			case <-ticker.C:
				continue
			}
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
		// HTTP check: use curl command to check if URL is accessible
		result, err := server.ExecuteCommand(fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' %s", app.CheckTarget))
		if err != nil {
			logs.Logger.Error("HTTP check error", zap.Error(err))
			return false
		}
		logs.Logger.Debug("HTTP check result:", zap.String("status_code", result))
		// check if HTTP status code is 2xx or 3xx
		return len(result) > 0 && (result[0] == '2' || result[0] == '3')
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
