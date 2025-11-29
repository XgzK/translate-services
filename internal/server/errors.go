package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// APIError 统一的 API 错误响应结构 (规范化错误处理喵～)
type APIError struct {
	Code    string `json:"code"`              // 错误代码
	Message string `json:"message"`           // 错误消息
	Details any    `json:"details,omitempty"` // 详细信息（可选）
}

// 预定义的错误代码常量
const (
	ErrCodeInvalidRequest     = "INVALID_REQUEST"
	ErrCodeMissingParameter   = "MISSING_PARAMETER"
	ErrCodeUnsupportedFormat  = "UNSUPPORTED_FORMAT"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeTranslationFailed  = "TRANSLATION_FAILED"
)

// NewAPIError 创建 API 错误，参数: 错误代码与消息，返回: APIError 指针
func NewAPIError(code, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// WithDetails 添加错误详情，参数: 详情信息，返回: 带详情的 APIError 指针
func (e *APIError) WithDetails(details any) *APIError {
	e.Details = details
	return e
}

// Error 实现 error 接口，参数: 无，返回: 错误字符串
func (e *APIError) Error() string {
	return e.Message
}

// ========== 便捷的错误响应函数 ==========

// BadRequest 返回 400 错误响应，参数: Echo 上下文、错误代码、消息，返回: error
func BadRequest(c echo.Context, code, message string) error {
	return c.JSON(http.StatusBadRequest, NewAPIError(code, message))
}

// BadRequestWithDetails 返回带详情的 400 错误响应，参数: Echo 上下文、错误代码、消息、详情，返回: error
func BadRequestWithDetails(c echo.Context, code, message string, details any) error {
	return c.JSON(http.StatusBadRequest, NewAPIError(code, message).WithDetails(details))
}

// BadGateway 返回 502 错误响应，参数: Echo 上下文、错误代码、消息，返回: error
func BadGateway(c echo.Context, code, message string) error {
	return c.JSON(http.StatusBadGateway, NewAPIError(code, message))
}

// BadGatewayWithDetails 返回带详情的 502 错误响应，参数: Echo 上下文、错误代码、消息、详情，返回: error
func BadGatewayWithDetails(c echo.Context, code, message string, details any) error {
	return c.JSON(http.StatusBadGateway, NewAPIError(code, message).WithDetails(details))
}

// InternalError 返回 500 错误响应，参数: Echo 上下文、消息，返回: error
func InternalError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, NewAPIError(ErrCodeInternalError, message))
}
