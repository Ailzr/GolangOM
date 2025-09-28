package response

import (
	"GolangOM/constant"

	"github.com/gin-gonic/gin"
)

// Response is a generic API response struct
type Response struct {
	Code    constant.ServiceErrorCode `json:"code"` // business status code
	Message string                    `json:"msg"`  // user prompt message
	Data    interface{}               `json:"data"` // data returned on success, can be nil
}

// --- your response functions can be simplified and unified ---

// Success returns a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// Fail returns a failed response
// httpCode: HTTP status code (e.g., 400, 401, 500)
// code: custom business error code
// msg: error message
func Fail(c *gin.Context, httpCode int, code constant.ServiceErrorCode, msg string) {
	c.JSON(httpCode, Response{
		Code:    code,
		Message: msg,
		Data:    nil,
	})
}
