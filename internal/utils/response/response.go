package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const CodeSuccess = 200

type Body struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"requestId"`
	Data      interface{} `json:"data,omitempty"`
}

// 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code:      CodeSuccess,
		Message:   "ok",
		RequestID: c.GetString("requestId"),
		Data:      data,
	})
}

// 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Body{
		Code:      CodeSuccess,
		Message:   "ok",
		RequestID: c.GetString("requestId"),
		Data:      data,
	})
}

// 错误响应
func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Body{
		Code:      code,
		Message:   message,
		RequestID: c.GetString("requestId"),
	})
}
