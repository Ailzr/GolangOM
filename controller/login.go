package controller

import (
	"GolangOM/model"
	"GolangOM/response"
	"GolangOM/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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
			// 绑定失败，返回 400 Bad Request 和详细的错误信息
			response.Fail(c, http.StatusBadRequest, 400, "无效的请求格式: "+err.Error())
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
				response.Fail(c, http.StatusInternalServerError, 500, "登录失败，会话创建失败")
				return
			}

			// 返回成功的 JSON，无需返回 redirect URL
			response.Success(c, gin.H{"redirect": "/golang-om"})
		} else {
			// 登录失败，返回 401 Unauthorized
			response.Fail(c, http.StatusUnauthorized, 401, "用户名或密码错误")
		}
	}
}

func LogoutLogicFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()                               // 清除 Session 中的所有数据
		session.Options(sessions.Options{MaxAge: -1}) // 设置 Cookie 立即失效
		if err := session.Save(); err != nil {
			response.Fail(c, http.StatusInternalServerError, 500, "退出登录失败")
			return
		}
		response.Success(c, map[string]string{"redirect": "/login"})
	}
}
