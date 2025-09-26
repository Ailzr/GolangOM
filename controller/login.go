package controller

import (
	"GolangOM/constant"
	"GolangOM/logs"
	"GolangOM/model"
	"GolangOM/response"
	"GolangOM/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func LoginPageFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") != nil {
			c.Redirect(http.StatusFound, "/golang-om")
			return
		}
		// 未登录：返回登录页面
		c.HTML(http.StatusOK, "login.html", nil)
	}
}

// LoginRequest 定义了登录 API 的请求体结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func LoginLogicFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 创建一个用于接收请求数据的结构体实例
		var req LoginRequest

		// 2. 使用 ShouldBindJSON 将请求体（JSON）绑定到结构体上
		// 如果绑定失败（例如，缺少 username 字段），ShouldBindJSON 会返回一个错误
		if err := c.ShouldBindJSON(&req); err != nil {
			// 绑定失败
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, "请求参数绑定失败")
			logs.Logger.Error("请求参数绑定失败: ", zap.Error(err))
			return // 必须 return，防止继续执行后续代码
		}

		// 3. 从绑定后的结构体中获取数据
		username := strings.TrimSpace(req.Username)
		password := strings.TrimSpace(req.Password)

		user := model.User{Username: username}

		// 4. 业务逻辑验证
		if user.IsExists() && util.GetMd5(password) == user.Password {
			// 登录成功，创建session
			session := sessions.Default(c)
			session.Set("username", username)
			if err := session.Save(); err != nil {
				response.Fail(c, http.StatusInternalServerError, constant.SessionError, "登录失败，会话创建失败")
				logs.Logger.Error("登录时创建session 失败: ", zap.Error(err))
				return
			}

			// 返回成功的 JSON，无需返回 redirect URL
			response.Success(c, gin.H{"redirect": "/golang-om"})
		} else {
			// 登录失败，返回 401 Unauthorized
			response.Fail(c, http.StatusUnauthorized, constant.AuthError, "用户名或密码错误")
		}
	}
}

func LogoutLogicFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()                               // 清除 Session 中的所有数据
		session.Options(sessions.Options{MaxAge: -1}) // 设置 Cookie 立即失效
		if err := session.Save(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.SessionError, "退出登录失败")
			logs.Logger.Error("退出登录失败: ", zap.Error(err))
			return
		}
		response.Success(c, map[string]string{"redirect": "/login"})
	}
}
