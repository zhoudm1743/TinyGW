package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 状态码
	Message string      `json:"message"` // 消息
	Data    interface{} `json:"data"`    // 数据
}

// 成功响应
func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "操作成功",
		Data:    data,
	})
}

// 成功响应带消息
func SuccessWithMsg(ctx *gin.Context, message string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// 分页数据结构
type PageData struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Data     interface{} `json:"data"`
}

// 分页响应
func SuccessWithPage(ctx *gin.Context, total int64, page, pageSize int, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "操作成功",
		Data: PageData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			Data:     data,
		},
	})
}

// 失败响应
func Fail(ctx *gin.Context, code int, message string) {
	ctx.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// 参数错误响应
func FailWithParamError(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusOK, Response{
		Code:    400,
		Message: message,
		Data:    nil,
	})
}

// 未授权响应
func FailWithUnauthorized(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{
		Code:    401,
		Message: "未授权或授权已过期",
		Data:    nil,
	})
}

// 禁止访问响应
func FailWithForbidden(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{
		Code:    403,
		Message: "无权访问",
		Data:    nil,
	})
}

// 资源不存在响应
func FailWithNotFound(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusOK, Response{
		Code:    404,
		Message: message,
		Data:    nil,
	})
}

// 服务器错误响应
func FailWithServerError(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusOK, Response{
		Code:    500,
		Message: message,
		Data:    nil,
	})
}
