package main

import (
	_ "GolangOM/config"
	"GolangOM/constant"
	"GolangOM/database"
	"GolangOM/logs"
	"GolangOM/model"
	"GolangOM/pkg"
	"GolangOM/router"
	"GolangOM/util"
	"strconv"

	"go.uber.org/zap"
)

func main() {
	// ensure log file is closed
	defer logs.Logger.Sync()

	Init()

	router.InitRouter()

	logs.Logger.Info("Exit Successfully")
}

func Init() {
	err := database.DB.AutoMigrate(&model.User{}, &model.ServerModel{}, &model.AppModel{})
	if err != nil {
		logs.Logger.Error("AutoMigrate failed", zap.Error(err))
		panic(err)
	}

	user := model.User{Username: "admin"}
	// if no admin user exists, create an admin user
	if !user.IsExists() {
		user.Password = util.GetMd5("123456")
		user.Role = "admin"
		err := user.CreateUser()
		if err != nil {
			logs.Logger.Error("CreateUser failed", zap.Error(err))
			panic(err)
		}
	}

	localServer := model.ServerModel{}
	localServer.ID = constant.LocalServerID
	if !localServer.IsExists() {
		localServer.IP = "127.0.0.1"
		err := localServer.CreateServer()
		if err != nil {
			logs.Logger.Error("CreateLocalServer failed", zap.Error(err))
			panic(err)
		}
	}

	servers, err := model.GetServerList()
	if err != nil {
		logs.Logger.Error("GetServerList failed", zap.Error(err))
	}

	for _, server := range servers {
		if server.ID == constant.LocalServerID {
			continue
		}
		// use goroutine to establish connection, avoid blocking main thread
		go func() {
			err := pkg.GetConnectionPool().NewConnection(&pkg.ServerConfig{
				AuthMethod: server.AuthMethod,
				Credential: server.Credential,
				ID:         server.ID,
				IP:         server.IP,
				Password:   server.Password,
				Port:       server.Port,
				User:       server.User,
			})
			if err != nil {
				logs.Logger.Error("NewConnection failed", zap.String("server_id", strconv.Itoa(int(server.ID))), zap.Error(err))
			}
		}()
	}

	apps, err := model.GetAppList()
	if err != nil {
		logs.Logger.Error("GetAppList failed", zap.Error(err))
	}

	for _, app := range apps {
		// use goroutine to establish connection, avoid blocking main thread
		go func() {
			err := pkg.GetAppCheckerManager().NewAppChecker(&pkg.AppCheckConfig{
				AutoRestart:   app.AutoRestart,
				CheckInterval: app.CheckInterval,
				CheckTarget:   app.CheckTarget,
				CheckType:     app.CheckType,
				ID:            app.ID,
				Name:          app.Name,
				ServerID:      app.ServerID,
				StartScript:   app.StartScript,
			})
			if err != nil {
				logs.Logger.Error("NewAppChecker failed", zap.String("app_id", strconv.Itoa(int(app.ID))), zap.Error(err))
			}
		}()
	}

}
