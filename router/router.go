package router

import (
	"GolangOM/controller"
	"GolangOM/middleware"
	"GolangOM/ws"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func InitRouter() {
	r := gin.Default()

	store := memstore.NewStore([]byte("secret-key-123")) // encryption key (production environment needs complex key)
	// configure Session features:
	// - CookieName: key name for Session ID stored in Cookie
	// - MaxAge: 0 means "browser session level" (Cookie expires after browser closes)
	// - Path: "/" means effective for all paths
	// - HttpOnly: true prevents frontend JS from reading Cookie, avoiding XSS attacks
	devMode := viper.GetBool("Server.Development")
	secure := true
	if devMode {
		secure = false
	} else {
		secure = true
	}
	r.Use(sessions.Sessions("gin-session-id", store), func(c *gin.Context) {
		session := sessions.Default(c)
		session.Options(sessions.Options{
			MaxAge:   0,                    // session level Cookie (expires when browser closes)
			Path:     "/",                  // effective for all paths
			HttpOnly: true,                 // prevent frontend JS from accessing Cookie
			Secure:   secure,               // false for development (HTTP), true for production (HTTPS)
			SameSite: http.SameSiteLaxMode, // prevent CSRF attacks
		})
		c.Next()
	})

	r.LoadHTMLGlob("templates/*")

	login := r.Group("")
	{
		login.GET("/login", controller.LoginPageFunc())
		login.POST("/api/login", controller.LoginLogicFunc())
	}

	pages := r.Group("")
	pages.Use(middleware.AuthMiddleware())
	{
		pages.GET("/golang-om", controller.IndexPageFunc())
	}

	apis := r.Group("/api")
	apis.Use(middleware.AuthMiddleware())
	{
		apis.POST("/logout", controller.LogoutLogicFunc())
		apis.GET("/user/info", controller.UserInfoFunc())
		apis.GET("/server/list", controller.GetServerListFunc())
		apis.POST("/server/get", controller.GetServerFunc())
		apis.POST("/server/create", controller.CreateServerFunc())
		apis.POST("/server/update", controller.UpdateServerFunc())
		apis.POST("/server/delete", controller.DeleteServerFunc())
		apis.GET("/app/list", controller.GetAppListFunc())
		apis.POST("/app/get", controller.GetAppFunc())
		apis.POST("/app/create", controller.CreateAppFunc())
		apis.POST("/app/update", controller.UpdateAppFunc())
		apis.POST("/app/delete", controller.DeleteAppFunc())

		apis.GET("/ws", ws.WebsocketFunc())
	}

	port := viper.GetInt("Server.WebUIPort")
	r.Run(fmt.Sprintf(":%d", port))
}
