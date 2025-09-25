package router

import (
	"GolangOM/middleware"
	"GolangOM/webUI"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

func InitRouter() {
	r := gin.Default()

	store := memstore.NewStore([]byte("secret-key-123")) // 加密密钥（生产环境需复杂密钥）
	// 配置 Session 特性：
	// - CookieName: Session ID 存储在 Cookie 中的键名
	// - MaxAge: 0 表示“浏览器会话级”（关闭浏览器后 Cookie 失效）
	// - Path: "/" 表示所有路径都生效
	// - HttpOnly: true 防止前端 JS 读取 Cookie，避免 XSS 攻击
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
			MaxAge:   0,                    // 会话级 Cookie（关闭浏览器失效）
			Path:     "/",                  // 所有路径生效
			HttpOnly: true,                 // 禁止前端 JS 访问 Cookie
			Secure:   secure,               // 开发环境用 false（HTTP），生产环境用 true（HTTPS）
			SameSite: http.SameSiteLaxMode, // 防止 CSRF 攻击
		})
		c.Next()
	})

	r.LoadHTMLGlob("templates/*")

	login := r.Group("")
	{
		login.GET("/login", webUI.LoginPageFunc())
		login.POST("/api/login", webUI.LoginLogicFunc())
	}

	pages := r.Group("")
	pages.Use(middleware.AuthMiddleware())
	{
		pages.GET("/golang-om", webUI.IndexPageFunc())
	}

	apis := r.Group("/api")
	apis.Use(middleware.AuthMiddleware())
	{
		apis.POST("/logout", webUI.LogoutLogicFunc())
		apis.GET("/user/info", webUI.UserInfoFunc())
	}

	port := viper.GetInt("Server.WebUIPort")
	r.Run(fmt.Sprintf(":%d", port))
}
