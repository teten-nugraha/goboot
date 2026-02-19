package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiResponse struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	TraceId string      `json:"traceId,omitempty"`
}

// ----------------------
// SUCCESS RESPONSES
// ----------------------
func OK(c *gin.Context, data interface{}) {
	c.JSONP(http.StatusOK, ApiResponse{
		Success: true,
		Code:    "OK",
		Data:    data,
	})
}
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, ApiResponse{
		Success: true,
		Code:    "OK",
		Data:    data,
	})
}

// ----------------------
// ERROR RESPONSES
// ----------------------
func ErrorResponse(c *gin.Context, httpStatus int, code string, message string, details interface{}) {
	traceId := c.GetString("traceId")
	c.JSON(httpStatus, ApiResponse{
		Success: false,
		Code:    code,
		Error: &ErrorBody{
			Message: message,
			Details: details,
			TraceId: traceId,
		},
	})
}
func BadRequest(c *gin.Context, message string, details interface{}) {
	ErrorResponse(c, http.StatusBadRequest, "BAD_RQUEST", message, details)
}

func Unauthorized(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func Forbidden(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

func NotFound(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func InternalError(c *gin.Context, message string, details interface{}) {
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, details)
}
