package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthRequired JWT 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "缺少认证令牌",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "认证令牌格式错误",
				"data":    nil,
			})
			c.Abort()
			return
		}

		token := parts[1]

		// TODO: 验证 JWT token
		// 现在先简单验证非空
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "无效的认证令牌",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", "admin") // TODO: 从 JWT 中解析
		c.Next()
	}
}
