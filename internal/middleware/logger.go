package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		slog.Info("HTTP Request",
			"status", statusCode,
			"method", method,
			"path", path,
			"ip", clientIP,
			"latency_ms", float64(latency.Microseconds())/1000.0,
			"errors", c.Errors.String(),
		)
	}
}
