package middleware

import (
	"GolangOM/logs"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get Session instance
		session := sessions.Default(c)
		// get user information from Session (stored "username" during login)
		username := session.Get("username")
		logs.Logger.Debug("middleware username:", zap.Any("username", username))

		// if no user information in Session, user is not logged in, redirect to login page
		if username == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// logged in: store username in context for subsequent interfaces to use
		c.Set("username", username)
		c.Next() // continue executing subsequent Handler
	}
}
