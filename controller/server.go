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

func GetServerListFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		servers := pkg.GetConnectionPool().GetServers()
		response.Success(c, gin.H{"servers": servers})
	}
}

func CreateServerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		server := &model.ServerModel{}

		if err := c.ShouldBindJSON(server); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, "请求参数绑定失败")
			logs.Logger.Error("请求参数绑定失败: ", zap.Error(err))
			return
		}

		if err := server.CreateServer(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "服务器创建失败")
			logs.Logger.Error("服务器创建失败: ", zap.Error(err))
			return
		}

		err := pkg.GetConnectionPool().NewConnection(&pkg.ServerConfig{
			ID:         server.ID,
			IP:         server.IP,
			Port:       server.Port,
			User:       server.User,
			AuthMethod: server.AuthMethod,
			Credential: server.Credential,
			Password:   server.Password,
		})
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.ServerConnectError, "服务器连接失败")
			logs.Logger.Error("服务器连接失败: ", zap.Error(err))
			return
		}

		response.Success(c, gin.H{"server": server})
	}
}

func UpdateServerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		server := &model.ServerModel{}

		if err := c.ShouldBindJSON(server); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, "请求参数绑定失败")
			logs.Logger.Error("请求参数绑定失败: ", zap.Error(err))
			return
		}

		if !server.IsExists() {
			response.Fail(c, http.StatusBadRequest, constant.TargetNotFound, "服务器不存在")
			return
		}

		if err := server.UpdateServer(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "服务器更新失败")
			logs.Logger.Error("服务器更新失败: ", zap.Error(err))
			return
		}

		pkg.GetConnectionPool().RemoveServerFromConnectionPoolByID(server.ID)

		if err := pkg.GetConnectionPool().NewConnection(&pkg.ServerConfig{
			ID:         server.ID,
			IP:         server.IP,
			Port:       server.Port,
			User:       server.User,
			AuthMethod: server.AuthMethod,
			Credential: server.Credential,
			Password:   server.Password,
		}); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.ServerConnectError, "服务器连接失败")
			logs.Logger.Error("服务器连接失败: ", zap.Error(err))
			return
		}
		response.Success(c, gin.H{"server": server})
	}
}
