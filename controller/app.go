package controller

import (
	"GolangOM/constant"
	"GolangOM/logs"
	"GolangOM/model"
	"GolangOM/pkg"
	"GolangOM/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func GetAppListFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		apps := pkg.GetAppCheckerManager().GetAppCheckers()
		response.Success(c, gin.H{"apps": apps})
	}
}
func CreateAppFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		app := &model.AppModel{}
		if err := c.ShouldBindJSON(app); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, err.Error())
			logs.Logger.Error("参数错误", zap.Error(err))
			return
		}

		if err := app.CreateApp(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, err.Error())
			logs.Logger.Error("创建应用失败", zap.Error(err))
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
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "创建应用失败")
			logs.Logger.Error("创建应用失败", zap.Error(err))
		}
		response.Success(c, gin.H{"app": app})
	}
}

func UpdateAppFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		app := &model.AppModel{}
		if err := c.ShouldBindJSON(app); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, err.Error())
			logs.Logger.Error("参数错误", zap.Error(err))
			return
		}

		if !app.IsExists() {
			response.Fail(c, http.StatusBadRequest, constant.TargetNotFound, "应用不存在")
		}

		if err := app.UpdateApp(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "更新应用失败")
			logs.Logger.Error("更新应用失败", zap.Error(err))
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
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "更新应用失败")
			return
		}
		response.Success(c, gin.H{"app": app})
	}
}
