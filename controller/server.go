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

type serverVo struct {
	ID          uint                   `json:"id"`
	AuthMethod  constant.AuthMethod    `json:"auth_method"`
	IP          string                 `json:"ip"`
	Port        int                    `json:"port"`
	User        string                 `json:"user"`
	CheckResult constant.ConnectStatus `json:"check_result"`
	CheckTime   time.Time              `json:"check_time"`
}

func GetServerListFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		servers := pkg.GetConnectionPool().GetServers()
		result := make([]serverVo, 0, len(servers))
		for _, server := range servers {
			result = append(result, serverVo{
				ID:          server.ID,
				AuthMethod:  server.AuthMethod,
				IP:          server.IP,
				Port:        server.Port,
				User:        server.User,
				CheckResult: server.Status,
				CheckTime:   server.LastCheckTime,
			})
		}
		response.Success(c, gin.H{"servers": result})
	}
}

func GetServerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID uint `json:"id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, "parameter bind error")
			logs.Logger.Error("parameter bind error: ", zap.Error(err))
			return
		}

		server := &model.ServerModel{Model: gorm.Model{ID: req.ID}}

		if !server.IsExists() {
			response.Fail(c, http.StatusBadRequest, constant.TargetNotFound, "server not exists")
			return
		}

		serverVo := serverVo{
			ID:          server.ID,
			AuthMethod:  server.AuthMethod,
			IP:          server.IP,
			Port:        server.Port,
			User:        server.User,
			CheckResult: "", // do not return sensitive information
			CheckTime:   time.Time{},
		}

		response.Success(c, gin.H{"server": serverVo})
	}
}

func CreateServerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		server := &model.ServerModel{}

		if err := c.ShouldBindJSON(server); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, "parameter bind error")
			logs.Logger.Error("parameter bind error: ", zap.Error(err))
			return
		}

		if err := server.CreateServer(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "server create failed")
			logs.Logger.Error("server create failed: ", zap.Error(err))
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
			response.Fail(c, http.StatusInternalServerError, constant.ServerConnectError, "server connect failed")
			logs.Logger.Error("server connect failed: ", zap.Error(err))
			return
		}

		response.Success(c, gin.H{"server": server})
	}
}

func UpdateServerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		server := &model.ServerModel{}

		if err := c.ShouldBindJSON(server); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, "parameter bind error")
			logs.Logger.Error("parameter bind error: ", zap.Error(err))
			return
		}

		if !server.IsExists() {
			response.Fail(c, http.StatusBadRequest, constant.TargetNotFound, "server not exists")
			return
		}

		if err := server.UpdateServer(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "server update failed")
			logs.Logger.Error("server update failed: ", zap.Error(err))
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
			response.Fail(c, http.StatusInternalServerError, constant.ServerConnectError, "server connect failed")
			logs.Logger.Error("server connect failed: ", zap.Error(err))
			return
		}
		response.Success(c, gin.H{"server": server})
	}
}

func DeleteServerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID uint `json:"id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, "parameter bind error")
			logs.Logger.Error("parameter bind error: ", zap.Error(err))
			return
		}

		server := &model.ServerModel{Model: gorm.Model{ID: req.ID}}

		if !server.IsExists() {
			response.Fail(c, http.StatusBadRequest, constant.TargetNotFound, "server not exists")
			return
		}

		// remove from connection pool first
		pkg.GetConnectionPool().RemoveServerFromConnectionPoolByID(req.ID)

		// delete database record
		if err := server.DeleteServer(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.UnknownError, "server delete failed")
			logs.Logger.Error("server delete failed: ", zap.Error(err))
			return
		}

		response.Success(c, gin.H{"message": "server deleted successfully"})
	}
}
