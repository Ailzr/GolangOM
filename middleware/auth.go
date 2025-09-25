package middleware

import (
	"GolangOM/logs"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Session 实例
		session := sessions.Default(c)
		// 从 Session 中获取用户信息（登录时存储的 "username"）
		username := session.Get("username")
		logs.Logger.Debug("middleware username:", zap.Any("username", username))

		// 若 Session 中无用户信息，说明未登录，重定向到登录页
		if username == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 已登录：将用户名存入上下文，供后续接口使用
		c.Set("username", username)
		c.Next() // 继续执行后续 Handler
	}
}
