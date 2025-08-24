// pkg/gateway/middleware.go
package gateway

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtacn/agenticai/internal/logger"
	"github.com/turtacn/agenticai/pkg/security"
)

type middlewareChain struct{}

// Auth 认证：提取 JWT 或 SPIFFE header
func (mc *middlewareChain) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != "Bearer valid-token" { // TODO: 支持 SPIFFE
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("subject", "user@example.com")
		c.Next()
	}
}

// Authorize 授权：RBAC 检查
func (mc *middlewareChain) Authorize(action, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sub := c.MustGet("subject").(string)
		rbac := security.MustRBAC()
		if err := rbac.Authorize(c, sub, action, resource); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.Next()
	}
}

// Recover 统一 panic 处理
func (mc *middlewareChain) Recover() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.ErrorWithCtx(c, "panic recovered", zap.Any("err", recovered))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	})
}

// Logging 记录请求/响应链
func (mc *middlewareChain) Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.InfoWithCtx(c, "http access",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
	}
}

// TraceID 注入
func (mc *middlewareChain) TraceID(traceHeader string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := trace.SpanContextFromContext(c)
		c.Header(traceHeader, ctx.TraceID().String())
		c.Next()
	}
}
//Personal.AI order the ending
