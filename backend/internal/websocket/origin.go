package websocket

import (
	"log"
	"strings"
)

// isOriginAllowed checks if the given origin is allowed based on the configured patterns.
// Supports exact matches and wildcard patterns (e.g., *.clpr.gg, *.clpr.tv).
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	// Check against each allowed origin pattern
	for _, pattern := range allowedOrigins {
		if matchesPattern(origin, pattern) {
			return true
		}
	}

	return false
}

// matchesPattern checks if an origin matches a pattern.
// Supports exact matches and wildcard patterns with a single asterisk at the start.
func matchesPattern(origin, pattern string) bool {
	// Exact match
	if origin == pattern {
		return true
	}

	// Wildcard pattern matching (e.g., *.clpr.gg)
	if strings.HasPrefix(pattern, "*.") {
		// Extract the domain suffix from the pattern (e.g., "clpr.gg" from "*.clpr.gg")
		suffix := pattern[2:] // Remove "*."

		// Extract the domain from the origin (remove protocol)
		originDomain := strings.TrimPrefix(origin, "http://")
		originDomain = strings.TrimPrefix(originDomain, "https://")
		// Remove port if present
		if idx := strings.Index(originDomain, ":"); idx != -1 {
			originDomain = originDomain[:idx]
		}

		// Check if origin domain ends with the suffix
		if strings.HasSuffix(originDomain, suffix) {
			// Ensure it's a proper subdomain match
			// Either exact match or has a dot before the suffix
			if originDomain == suffix {
				return true
			}
			if len(originDomain) > len(suffix) && originDomain[len(originDomain)-len(suffix)-1] == '.' {
				return true
			}
		}
	}

	return false
}

// validateAllowedOrigins validates the configured allowed origins on startup.
// Logs warnings for potentially insecure configurations.
func validateAllowedOrigins(allowedOrigins []string) {
	if len(allowedOrigins) == 0 {
		log.Println("WARNING: No WebSocket allowed origins configured. All origins will be rejected.")
		return
	}

	for _, origin := range allowedOrigins {
		// Check for wildcard "*" (open CORS)
		if origin == "*" {
			log.Println("SECURITY WARNING: WebSocket allows all origins (*). This is insecure and should never be used in production!")
		}

		// Check for overly broad wildcards
		if origin == "*.*" {
			log.Println("SECURITY WARNING: WebSocket allows all subdomains (*.*). This is very insecure!")
		}
	}

	log.Printf("WebSocket CORS configured with %d allowed origin(s)", len(allowedOrigins))
}
