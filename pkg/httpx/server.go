package httpx

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gostratum/core/pkg/configx"
)

// New constructs an HTTP handler backed by Gin while keeping the framework
// hidden from callers.
func New(cfg *configx.Config) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	if cfg != nil && cfg.Security.EnableCORS {
		engine.Use(corsMiddleware(cfg.Security.AllowedOrigins))
	}

	engine.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	engine.GET("/readyz", func(c *gin.Context) {
		c.String(http.StatusOK, "ready")
	})

	return engine
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowAll := len(allowedOrigins) == 0
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAll = true
			break
		}
	}

	normalized := make([]string, 0, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		normalized = append(normalized, strings.TrimSpace(origin))
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := ""
		switch {
		case allowAll && origin == "":
			allowed = "*"
		case allowAll:
			allowed = origin
		default:
			for _, candidate := range normalized {
				if candidate == origin {
					allowed = origin
					break
				}
			}
		}

		header := c.Writer.Header()
		if allowed != "" {
			header.Set("Access-Control-Allow-Origin", allowed)
			header.Set("Vary", "Origin")
		}
		header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
		header.Set("Access-Control-Max-Age", "600")
		if allowed != "*" && allowed != "" {
			header.Set("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
