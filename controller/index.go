package controller

import (
	"GolangOM/response"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func IndexPageFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 传递用户名到前端页面，用于显示欢迎信息
		c.HTML(http.StatusOK, "index.html", nil)
	}
}

func UserInfoFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, _ := c.Get("username")
		response.Success(c, gin.H{
			"username":  username,
			"loginTime": time.Now().Unix(),
			"role":      "管理员",
		})
	}
}
