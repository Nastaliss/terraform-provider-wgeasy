// Package client provides the HTTP client for interacting with the wg-easy REST API.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
)

// WGEasyClient is the HTTP client for the wg-easy REST API.
type WGEasyClient struct {
	endpoint   string
	username   string
	password   string
	httpClient *http.Client
	loginMu    sync.Mutex // Serializes login attempts
	loggedIn   bool       // Tracks if we've successfully logged in
}

// NewWGEasyClient creates a new API client for wg-easy.
func NewWGEasyClient(endpoint, username, password string) (*WGEasyClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	return &WGEasyClient{
		endpoint:   strings.TrimRight(endpoint, "/"),
		username:   username,
		password:   password,
		httpClient: &http.Client{
			Jar: jar,
		},
	}, nil
}

// login authenticates with the wg-easy API via POST /api/session.
func (c *WGEasyClient) login() error {
	body, err := json.Marshal(map[string]interface{}{
		"username": c.username,
		"password": c.password,
		"remember": true,
	})
	if err != nil {
		return fmt.Errorf("marshaling login request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/session", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-wgeasy/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return &AuthenticationError{
			Message: fmt.Sprintf("status %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	return nil
}

// ensureLoggedIn performs login if not already authenticated.
// Uses mutex to prevent concurrent login attempts.
func (c *WGEasyClient) ensureLoggedIn() error {
	c.loginMu.Lock()
	defer c.loginMu.Unlock()

	if c.loggedIn {
		return nil
	}

	if err := c.login(); err != nil {
		return err
	}
	c.loggedIn = true
	return nil
}

// doRequest performs an HTTP request with automatic re-login on 401.
func (c *WGEasyClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
	// Ensure we're logged in before making requests
	if err := c.ensureLoggedIn(); err != nil {
		return nil, err
	}

	resp, err := c.doRequestOnce(method, path, body)
	if err != nil {
		return nil, err
	}

	// If we get 401, session expired - re-login and retry once.
	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()

		c.loginMu.Lock()
		c.loggedIn = false
		err := c.login()
		if err == nil {
			c.loggedIn = true
		}
		c.loginMu.Unlock()

		if err != nil {
			return nil, err
		}
		return c.doRequestOnce(method, path, body)
	}

	return resp, nil
}

func (c *WGEasyClient) doRequestOnce(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.endpoint+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "terraform-provider-wgeasy/1.0")
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// GetClients returns all WireGuard clients/peers.
func (c *WGEasyClient) GetClients() ([]Client, error) {
	resp, err := c.doRequest(http.MethodGet, "/api/client", nil)
	if err != nil {
		return nil, fmt.Errorf("fetching clients: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading clients response: %w", err)
	}

	var clients []Client
	if err := json.Unmarshal(body, &clients); err != nil {
		return nil, fmt.Errorf("decoding clients response: %w (body: %s)", err, string(body[:min(500, len(body))]))
	}

	return clients, nil
}

// GetClient returns a single WireGuard client by ID.
func (c *WGEasyClient) GetClient(id string) (*Client, error) {
	clients, err := c.GetClients()
	if err != nil {
		return nil, err
	}

	var foundIDs []string
	for _, client := range clients {
		foundIDs = append(foundIDs, client.ID.String())
		if client.ID.String() == id {
			return &client, nil
		}
	}

	return nil, &NotFoundError{ID: id, FoundIDs: foundIDs}
}

// CreateClient creates a new WireGuard client/peer.
// Returns the client ID from the response.
func (c *WGEasyClient) CreateClient(req CreateClientRequest) (string, error) {
	resp, err := c.doRequest(http.MethodPost, "/api/client", req)
	if err != nil {
		return "", fmt.Errorf("creating client: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d creating client: %s", resp.StatusCode, string(respBody))
	}

	var createResp CreateClientResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return "", fmt.Errorf("decoding create response: %w", err)
	}

	if createResp.ClientID.String() == "" {
		return "", fmt.Errorf("create response missing clientId")
	}

	return createResp.ClientID.String(), nil
}

// UpdateClient updates an existing WireGuard client/peer.
func (c *WGEasyClient) UpdateClient(id string, req UpdateClientRequest) (*Client, error) {
	// ServerAllowedIPs is non-nullable - ensure it's an array, not null.
	// AllowedIPs and DNS are nullable - nil is OK (serializes to JSON null).
	if req.ServerAllowedIPs == nil {
		req.ServerAllowedIPs = []string{}
	}

	path := fmt.Sprintf("/api/client/%s", id)
	resp, err := c.doRequest(http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("updating client %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{ID: id}
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d updating client %s: %s", resp.StatusCode, id, string(respBody))
	}

	// Read back the updated client to get server-authoritative values.
	return c.GetClient(id)
}

// DeleteClient deletes a WireGuard client/peer.
func (c *WGEasyClient) DeleteClient(id string) error {
	path := fmt.Sprintf("/api/client/%s", id)
	resp, err := c.doRequest(http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("deleting client %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d deleting client %s: %s", resp.StatusCode, id, string(respBody))
	}

	return nil
}
