// Package ginauth provides Gin middleware for JWT authentication.
// It wraps the shared auth.ParseToken from the auth package and provides
// a reusable WithClaims middleware for all Gin-based alethic-ism services.
//
// Usage:
//
//	ginauth.Init() // loads SECRET_KEY from env
//	router.POST("/api/v1/thing", ginauth.WithClaims(func(c *gin.Context, claims *auth.Claims) {
//	    // claims.UserID is available
//	}))
package ginauth

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quantumwake/alethic-ism-core-go/pkg/auth"
)

var secretKey []byte

// Init loads the shared SECRET_KEY from the environment.
// Call once at startup before registering any WithClaims handlers.
func Init() {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		log.Println("WARNING: SECRET_KEY not set, authenticated endpoints will reject all requests")
	} else {
		log.Printf("AUTH: SECRET_KEY loaded (%d bytes)", len(key))
	}
	secretKey = []byte(key)
}

// AuthenticatedHandler is a Gin handler that receives parsed JWT claims.
type AuthenticatedHandler func(c *gin.Context, claims *auth.Claims)

// WithClaims wraps a handler, extracting and validating the JWT Bearer token
// and injecting the parsed Claims. Returns 401 on missing/invalid tokens.
func WithClaims(handler AuthenticatedHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(secretKey) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth not configured"})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := auth.ParseToken(secretKey, parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		if claims.UserID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user_id in token"})
			return
		}

		handler(c, claims)
	}
}

// CORSMiddleware handles CORS preflight and response headers.
// In production CORS is typically handled by nginx ingress, but this
// is needed for local development.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "600")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
