package response

import (
	"github.com/gin-gonic/gin"
)

// Response 是一个通用的 API 响应结构体
type Response struct {
	Code    int         `json:"code"` // 业务状态码
	Message string      `json:"msg"`  // 给用户的提示信息
	Data    interface{} `json:"data"` // 成功时返回的数据，可为 nil
}

// --- 你的响应函数可以简化和统一 ---

// Success 返回一个成功的响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// Fail 返回一个失败的响应
// httpCode: HTTP 状态码 (e.g., 400, 401, 500)
// code: 自定义业务错误码
// msg: 错误信息
func Fail(c *gin.Context, httpCode int, code int, msg string) {
	c.JSON(httpCode, Response{
		Code:    code,
		Message: msg,
		Data:    nil,
	})
}
