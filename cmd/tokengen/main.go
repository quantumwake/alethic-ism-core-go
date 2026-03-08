// tokengen generates HMAC-signed JWT tokens for machine-to-machine authentication
// between alethic-ism services (e.g. file-source → vault-api).
//
// Usage:
//
//	go run ./cmd/tokengen -secret <key> -service <service-id> [-ttl 8760h]
//	go run ./cmd/tokengen -secret <key> -service <service-id> -no-expiry
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/quantumwake/alethic-ism-core-go/pkg/auth"
)

func main() {
	secret := flag.String("secret", "", "HMAC signing key (required, or set SECRET_KEY env)")
	serviceID := flag.String("service", "", "Service/user ID to embed in the token (required)")
	ttl := flag.Duration("ttl", 365*24*time.Hour, "Token time-to-live (e.g. 8760h, 720h)")
	noExpiry := flag.Bool("no-expiry", false, "Generate token with 100-year expiry (effectively unlimited)")
	flag.Parse()

	// Allow secret from env if not passed as flag
	key := *secret
	if key == "" {
		key = os.Getenv("SECRET_KEY")
	}
	if key == "" {
		fmt.Fprintln(os.Stderr, "error: -secret flag or SECRET_KEY env required")
		flag.Usage()
		os.Exit(1)
	}

	if *serviceID == "" {
		fmt.Fprintln(os.Stderr, "error: -service flag required")
		flag.Usage()
		os.Exit(1)
	}

	duration := *ttl
	if *noExpiry {
		duration = 100 * 365 * 24 * time.Hour // ~100 years
	}

	token, err := auth.GenerateToken([]byte(key), *serviceID, duration)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(token)
}
