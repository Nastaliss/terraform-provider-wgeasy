// Package client provides the HTTP client for interacting with the wg-easy REST API.
package client

import "fmt"

// NotFoundError is returned when a client/peer is not found.
type NotFoundError struct {
	ID       string
	FoundIDs []string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("client with ID %s not found (available IDs: %v)", e.ID, e.FoundIDs)
}

// AuthenticationError is returned when authentication fails.
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication failed: %s", e.Message)
}
