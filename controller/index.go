package controller

import (
	"GolangOM/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func IndexPageFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// pass username to frontend page for welcome message display
		c.HTML(http.StatusOK, "index.html", nil)
	}
}

func UserInfoFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, _ := c.Get("username")
		response.Success(c, gin.H{
			"username":  username,
			"loginTime": time.Now().Unix(),
			"role":      "admin",
		})
	}
}
