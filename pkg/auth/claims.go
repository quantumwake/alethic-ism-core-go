// Package auth provides JWT claims generation and parsing for service-to-service
// authentication across alethic-ism services. Uses HMAC-SHA256 signing with a
// shared SECRET_KEY, matching the contract in alethic-ism-vault-api/pkg/middleware/auth.go.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT payload shared across alethic-ism services.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT with the given user_id claim.
// secretKey is the shared HMAC signing key (SECRET_KEY env var).
// ttl controls token expiry; pass 0 for a long-lived service token (1 year).
func GenerateToken(secretKey []byte, userID string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 365 * 24 * time.Hour // Default: 1 year for service tokens
	}

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ParseToken validates a JWT string and returns the parsed Claims.
// secretKey is the shared HMAC signing key.
func ParseToken(secretKey []byte, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GenerateServiceToken is a convenience wrapper for creating a machine-to-machine
// token with a long TTL. Use this for static service credentials between
// processors and APIs (e.g. file-source → vault-api).
func GenerateServiceToken(secretKey []byte, serviceID string) (string, error) {
	return GenerateToken(secretKey, serviceID, 365*24*time.Hour)
}
