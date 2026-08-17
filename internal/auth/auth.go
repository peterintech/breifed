package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// GetApiKey extracts the API key from the request headers. It looks for the "Authorization" header and expects it to be in the format "Bearer <api_key>". If the header is not present or is malformed, it returns an error.
func GetApiKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header is missing")
	}

	// Split the header into two parts
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid Authorization header format")
	}
	if parts[0] != "ApiKey" {
		return "", fmt.Errorf("Authorization header must start with 'ApiKey'")
	}

	return parts[1], nil
}
