package middleware

import (
	"net/http"

	"emailAlert/internal/service"

	"github.com/gin-gonic/gin"
)

// APIResponse 简单的API响应结构
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// AuthMiddleware 认证中间件
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过登录接口
		if c.Request.URL.Path == "/api/v1/auth/login" {
			c.Next()
			return
		}

		var sessionID string

		// 对于监控日志端点（Server-Sent Events），支持通过URL参数传递session ID
		// 因为EventSource不支持自定义请求头
		if c.Request.URL.Path == "/api/v1/monitor/logs" {
			// 优先从查询参数获取session ID
			sessionID = c.Query("session_id")
			if sessionID == "" {
				// 如果查询参数没有session ID，再尝试从Cookie获取
				cookie, err := c.Cookie("session_id")
				if err == nil {
					sessionID = cookie
				}
			}
		} else {
			// 其他接口从Cookie获取session ID
			cookie, err := c.Cookie("session_id")
			if err == nil {
				sessionID = cookie
			}
		}

		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Code:    401,
				Message: "未找到会话信息",
				Error:   "",
			})
			c.Abort()
			return
		}

		user, err := authService.ValidateSession(sessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Code:    401,
				Message: "会话无效",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Set("session_id", sessionID) // 存储session ID，用于后续操作
		c.Next()
	}
}
