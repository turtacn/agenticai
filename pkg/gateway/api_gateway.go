// pkg/gateway/api_gateway.go
package gateway

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/turtacn/agenticai/internal/config"
	"github.com/turtacn/agenticai/internal/logger"
)

type Gateway struct {
	*http.Server
	ctx       context.Context
	mu        sync.RWMutex
	rateLimit *rate.Limiter
}

func New(cfg *config.GatewayConfig) *Gateway {
	r := gin.New()
	r.Use(otelgin.Middleware("gateway"), gin.Recovery())
	g := &Gateway{
		ctx:       context.Background(),
		rateLimit: rate.NewLimiter(rate.Every(time.Second), 100),
		Server: &http.Server{
			Addr:    cfg.Listen,
			Handler: r,
		},
	}
	g.setupRoutes(r)
	return g
}

func (g *Gateway) setupRoutes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	// api := r.Group("/api/v1", g.rateMiddleware())
	// api.POST("/agents/deploy", g.handleDeployAgent())
	// api.GET("/tasks/status/:id", g.handleTaskStatus())
}

func (g *Gateway) rateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !g.rateLimit.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
			return
		}
		c.Next()
	}
}

/*
func (g *Gateway) handleTaskStatus() gin.HandlerFunc {
	client := NewAgentClient() // 内部实现
	return func(c *gin.Context) {
		id := c.Param("id")
		res, err := client.GetTaskStatus(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, res)
	}
}
*/

func (g *Gateway) Start() error {
	logger.Info(g.ctx, "gateway starting", zap.String("addr", g.Addr))
	return g.ListenAndServe()
}
func (g *Gateway) Stop() error {
	return g.Shutdown(g.ctx)
}
//Personal.AI order the ending
