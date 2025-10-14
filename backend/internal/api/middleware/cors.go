package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件（基于白名单）
func CORS(allowedOrigins []string) gin.HandlerFunc {
	normalized := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			normalized[trimmed] = struct{}{}
		}
	}

	_, allowAll := normalized["*"]
	if allowAll {
		delete(normalized, "*")
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		headers := c.Writer.Header()

		if origin != "" {
			if allowAll {
				headers.Set("Access-Control-Allow-Origin", origin)
				headers.Set("Vary", "Origin")
			} else if _, ok := normalized[origin]; ok {
				headers.Set("Access-Control-Allow-Origin", origin)
				headers.Set("Vary", "Origin")
			}
		}

		if headers.Get("Access-Control-Allow-Origin") != "" {
			headers.Set("Access-Control-Allow-Credentials", "true")
			headers.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, Cache-Control, X-Requested-With, X-Request-ID")
			headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
