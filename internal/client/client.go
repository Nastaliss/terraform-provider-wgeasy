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
	"time"

	"github.com/pquerna/otp/totp"
)

// WGEasyClient is the HTTP client for the wg-easy REST API.
type WGEasyClient struct {
	endpoint   string
	username   string
	password   string
	totpSecret string // Base32 TOTP secret for 2FA logins (wg-easy >= 15.4.0); empty if 2FA is disabled.
	httpClient *http.Client
	loginMu    sync.Mutex // Serializes login attempts
	loggedIn   bool       // Tracks if we've successfully logged in
}

// NewWGEasyClient creates a new API client for wg-easy.
// totpSecret is the base32 TOTP secret used to satisfy two-factor
// authentication on wg-easy >= 15.4.0; pass "" when 2FA is not enabled.
func NewWGEasyClient(endpoint, username, password, totpSecret string) (*WGEasyClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	return &WGEasyClient{
		endpoint:   strings.TrimRight(endpoint, "/"),
		username:   username,
		password:   password,
		totpSecret: strings.ReplaceAll(totpSecret, " ", ""),
		httpClient: &http.Client{
			Jar: jar,
		},
	}, nil
}

// Login endpoints. wg-easy >= 15.4.0 moved password login from POST /api/session
// to POST /api/auth/password (Nuxt v4 rewrite that added OAuth/2FA). We try the
// new endpoint first and fall back to the legacy one for wg-easy < 15.4.0.
const (
	authPasswordPath  = "/api/auth/password" //#nosec G101 -- API route path, not a credential
	authVerify2FAPath = "/api/auth/verify-2fa"
	legacySessionPath = "/api/session"
)

// loginResponse is the JSON body returned by POST /api/auth/password on
// wg-easy >= 15.4.0. It returns HTTP 200 even when login is not complete
// (e.g. 2FA required), so the status field must be checked.
type loginResponse struct {
	Status string `json:"status"`
}

// login authenticates with the wg-easy API.
func (c *WGEasyClient) login() error {
	body, err := json.Marshal(map[string]interface{}{
		"username": c.username,
		"password": c.password,
		"remember": true,
	})
	if err != nil {
		return fmt.Errorf("marshaling login request: %w", err)
	}

	// Try the wg-easy >= 15.4.0 endpoint, falling back to the legacy path on 404.
	resp, err := c.postLogin(authPasswordPath, body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		resp, err = c.postLogin(legacySessionPath, body)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return &AuthenticationError{
			Message: fmt.Sprintf("status %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	// wg-easy >= 15.4.0 returns HTTP 200 with a status field even for
	// non-success outcomes (e.g. "TOTP_REQUIRED", "INVALID_TOTP_CODE").
	// The legacy endpoint has no status field, so only branch when it is
	// present and not "success".
	if len(respBody) > 0 {
		var lr loginResponse
		if err := json.Unmarshal(respBody, &lr); err == nil && lr.Status != "" && lr.Status != "success" {
			if lr.Status == "TOTP_REQUIRED" {
				return c.verify2FA()
			}
			return &AuthenticationError{
				Message: fmt.Sprintf("login not completed: %s", lr.Status),
			}
		}
	}

	return nil
}

// verify2FA completes a two-factor login by computing the current TOTP code
// from the configured secret and posting it to /api/auth/verify-2fa. The
// pending-login state is carried by the session cookie set during the
// password step.
func (c *WGEasyClient) verify2FA() error {
	if c.totpSecret == "" {
		return &AuthenticationError{
			Message: "two-factor authentication is required but no totp_secret was configured",
		}
	}

	code, err := totp.GenerateCode(c.totpSecret, time.Now())
	if err != nil {
		return &AuthenticationError{
			Message: fmt.Sprintf("generating TOTP code (check totp_secret is a valid base32 seed): %v", err),
		}
	}

	body, err := json.Marshal(map[string]interface{}{"totpCode": code})
	if err != nil {
		return fmt.Errorf("marshaling 2FA request: %w", err)
	}

	resp, err := c.postLogin(authVerify2FAPath, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return &AuthenticationError{
			Message: fmt.Sprintf("2FA verification failed: status %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	var lr loginResponse
	if err := json.Unmarshal(respBody, &lr); err == nil && lr.Status != "success" {
		return &AuthenticationError{
			Message: fmt.Sprintf("2FA verification not completed: %s", lr.Status),
		}
	}

	return nil
}

// postLogin issues a single login POST to the given path.
func (c *WGEasyClient) postLogin(path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, c.endpoint+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-wgeasy/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	return resp, nil
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
