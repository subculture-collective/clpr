package services

import "os"

// getTestEnv reads an environment variable or returns a default value
func getTestEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
