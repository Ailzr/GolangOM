package controller

import (
	"GolangOM/constant"
	"GolangOM/logs"
	"GolangOM/model"
	"GolangOM/pkg"
	"GolangOM/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type appVo struct {
	ID          uint                  `json:"id"`
	Name        string                `json:"name"`
	ServerID    uint                  `json:"server_id"`
	CheckType   constant.AppCheckType `json:"check_type"`
	CheckTarget string                `json:"check_target"`
	StartScript string                `json:"start_script"`
	AutoRestart bool                  `json:"auto_restart"`
	CheckResult bool                  `json:"check_result"`
	CheckTime   time.Time             `json:"last_check_time"`
}

func GetAppListFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		apps := pkg.GetAppCheckerManager().GetAppCheckers()
		result := make([]appVo, 0, len(apps))
		for _, app := range apps {
			result = append(result, appVo{
				ID:          app.ID,
				Name:        app.Name,
				ServerID:    app.ServerID,
				CheckType:   app.CheckType,
				CheckTarget: app.CheckTarget,
				StartScript: app.StartScript,
				AutoRestart: app.AutoRestart,
				CheckResult: app.LastCheckResult,
				CheckTime:   app.LastCheckTime,
			})
		}
		response.Success(c, gin.H{"apps": result})
	}
}

func GetAppFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID uint `json:"id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, "parameter bind error")
			logs.Logger.Error("parameter bind error: ", zap.Error(err))
			return
		}

		app := &model.AppModel{Model: gorm.Model{ID: req.ID}}

		if !app.IsExists() {
			response.Fail(c, http.StatusBadRequest, constant.TargetNotFound, "app not exists")
			return
		}

		appVo := appVo{
			ID:          app.ID,
			Name:        app.Name,
			ServerID:    app.ServerID,
			CheckType:   app.CheckType,
			StartScript: app.StartScript,
			AutoRestart: app.AutoRestart,
			CheckResult: false, // do not return sensitive information
			CheckTime:   time.Time{},
		}

		response.Success(c, gin.H{"app": appVo})
	}
}
func CreateAppFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		app := &model.AppModel{}
		if err := c.ShouldBindJSON(app); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, err.Error())
			logs.Logger.Error("parameter bind error", zap.Error(err))
			return
		}

		if err := app.CreateApp(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, err.Error())
			logs.Logger.Error("app create failed", zap.Error(err))
			return
		}

		if err := pkg.GetAppCheckerManager().NewAppChecker(&pkg.AppCheckConfig{
			AutoRestart:   app.AutoRestart,
			CheckInterval: app.CheckInterval,
			CheckTarget:   app.CheckTarget,
			CheckType:     app.CheckType,
			ID:            app.ID,
			Name:          app.Name,
			ServerID:      app.ServerID,
			StartScript:   app.StartScript,
		}); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "app create failed")
			logs.Logger.Error("app create failed", zap.Error(err))
		}
		response.Success(c, gin.H{"app": app})
	}
}

func UpdateAppFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		app := &model.AppModel{}
		if err := c.ShouldBindJSON(app); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, err.Error())
			logs.Logger.Error("parameter bind error", zap.Error(err))
			return
		}

		if !app.IsExists() {
			response.Fail(c, http.StatusBadRequest, constant.TargetNotFound, "app not exists")
		}

		if err := app.UpdateApp(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "update app failed")
			logs.Logger.Error("update app failed", zap.Error(err))
			return
		}

		pkg.GetAppCheckerManager().RemoveAppCheckerByID(app.ID)

		if err := pkg.GetAppCheckerManager().NewAppChecker(&pkg.AppCheckConfig{
			AutoRestart:   app.AutoRestart,
			CheckInterval: app.CheckInterval,
			CheckTarget:   app.CheckTarget,
			CheckType:     app.CheckType,
			ID:            app.ID,
			Name:          app.Name,
			ServerID:      app.ServerID,
			StartScript:   app.StartScript,
		}); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "update app failed")
			return
		}
		response.Success(c, gin.H{"app": app})
	}
}

func DeleteAppFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID uint `json:"id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, "parameter bind error")
			logs.Logger.Error("parameter bind error: ", zap.Error(err))
			return
		}

		app := &model.AppModel{Model: gorm.Model{ID: req.ID}}

		if !app.IsExists() {
			response.Fail(c, http.StatusBadRequest, constant.TargetNotFound, "app not exists")
			return
		}

		// remove from app checker first
		pkg.GetAppCheckerManager().RemoveAppCheckerByID(req.ID)

		// delete database record
		if err := app.DeleteApp(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "delete app error")
			logs.Logger.Error("delete app error: ", zap.Error(err))
			return
		}

		response.Success(c, gin.H{"message": "app deleted successfully"})
	}
}
