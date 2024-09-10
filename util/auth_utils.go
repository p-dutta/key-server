package util

import (
	"errors"
	"strings"
)

func GetTokenFromHeader(authorizationHeader *string) (*string, error) {
	// Split the header by whitespace (" ")
	headerParts := strings.Split(*authorizationHeader, " ")

	// Check if it has two parts and the first part is "Bearer"
	if len(headerParts) == 2 && headerParts[0] == "Bearer" {
		// Extract the token part
		*authorizationHeader = headerParts[1]
	} else {
		return nil, errors.New("invalid authorization header")
	}
	return authorizationHeader, nil
}
