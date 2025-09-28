package controller

import (
	"GolangOM/constant"
	"GolangOM/logs"
	"GolangOM/model"
	"GolangOM/response"
	"GolangOM/util"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoginPageFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") != nil {
			c.Redirect(http.StatusFound, "/golang-om")
			return
		}
		// not logged in: return login page
		c.HTML(http.StatusOK, "login.html", nil)
	}
}

// LoginRequest defines the request body structure for login API
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func LoginLogicFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. create a struct instance to receive request data
		var req LoginRequest

		// 2. use ShouldBindJSON to bind request body (JSON) to struct
		// if binding fails (e.g., missing username field), ShouldBindJSON returns an error
		if err := c.ShouldBindJSON(&req); err != nil {
			// binding failed
			response.Fail(c, http.StatusBadRequest, constant.ParameterError, "parameter bind error")
			logs.Logger.Error("parameter bind error: ", zap.Error(err))
			return // must return to prevent continuing execution
		}

		// 3. get data from bound struct
		username := strings.TrimSpace(req.Username)
		password := strings.TrimSpace(req.Password)

		user := model.User{Username: username}

		// 4. business logic validation
		if user.IsExists() && util.GetMd5(password) == user.Password {
			// login successful, create session
			session := sessions.Default(c)
			session.Set("username", username)
			if err := session.Save(); err != nil {
				response.Fail(c, http.StatusInternalServerError, constant.SessionError, "login failed, session creation failed")
				logs.Logger.Error("session creation failed during login: ", zap.Error(err))
				return
			}

			// return successful JSON, no need to return redirect URL
			response.Success(c, gin.H{"redirect": "/golang-om"})
		} else {
			// login failed, return 401 Unauthorized
			response.Fail(c, http.StatusUnauthorized, constant.AuthError, "invalid username or password")
		}
	}
}

func LogoutLogicFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()                               // clear all data in Session
		session.Options(sessions.Options{MaxAge: -1}) // set Cookie to expire immediately
		if err := session.Save(); err != nil {
			response.Fail(c, http.StatusInternalServerError, constant.SessionError, "logout failed")
			logs.Logger.Error("logout failed: ", zap.Error(err))
			return
		}
		response.Success(c, map[string]string{"redirect": "/login"})
	}
}
