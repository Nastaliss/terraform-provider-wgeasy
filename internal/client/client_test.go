package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *WGEasyClient) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewWGEasyClient(server.URL, "admin", "secret")
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	return server, client
}

func TestLogin(t *testing.T) {
	_, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" && r.Method == http.MethodPost {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if body["username"] == "admin" && body["password"] == "secret" && body["remember"] == true {
				http.SetCookie(w, &http.Cookie{Name: "session", Value: "test-session"})
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	err := client.login()
	if err != nil {
		t.Fatalf("expected successful login, got: %v", err)
	}
}

func TestLoginFailure(t *testing.T) {
	_, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid credentials"))
	})

	err := client.login()
	if err == nil {
		t.Fatal("expected error on failed login")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Fatalf("expected AuthenticationError, got: %T", err)
	}
}

func TestGetClients(t *testing.T) {
	authenticated := false
	_, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" && r.Method == http.MethodPost {
			authenticated = true
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "test"})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/client" && r.Method == http.MethodGet {
			if !authenticated {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			clients := []Client{
				{
					ID:          "abc-123",
					Name:        "test-client",
					Enabled:     true,
					IPv4Address: "10.8.0.2",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(clients)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	clients, err := client.GetClients()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}
	if clients[0].Name != "test-client" {
		t.Errorf("expected name 'test-client', got '%s'", clients[0].Name)
	}
}

func TestGetClient(t *testing.T) {
	_, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "test"})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/client" && r.Method == http.MethodGet {
			clients := []Client{
				{ID: "abc-123", Name: "client-1", IPv4Address: "10.8.0.2"},
				{ID: "def-456", Name: "client-2", IPv4Address: "10.8.0.3"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(clients)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	c, err := client.GetClient("def-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "client-2" {
		t.Errorf("expected 'client-2', got '%s'", c.Name)
	}
}

func TestGetClientNotFound(t *testing.T) {
	_, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "test"})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/client" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Client{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.GetClient("nonexistent")
	if err == nil {
		t.Fatal("expected NotFoundError")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("expected NotFoundError, got: %T", err)
	}
}

func TestCreateClient(t *testing.T) {
	_, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "test"})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/client" && r.Method == http.MethodPost {
			var req CreateClientRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := CreateClientResponse{
				Status:   "success",
				ClientID: "new-uuid-789",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	clientID, err := client.CreateClient(CreateClientRequest{Name: "new-client"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clientID != "new-uuid-789" {
		t.Errorf("expected 'new-uuid-789', got '%s'", clientID)
	}
}

func TestUpdateClient(t *testing.T) {
	_, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "test"})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/client/abc-123" && r.Method == http.MethodPost {
			var req UpdateClientRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "updated-name" || !req.Enabled || req.MTU != 1420 {
				t.Errorf("unexpected update request: %+v", req)
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/client" && r.Method == http.MethodGet {
			clients := []Client{
				{
					ID:          "abc-123",
					Name:        "updated-name",
					Enabled:     true,
					IPv4Address: "10.8.0.2",
					MTU:         1420,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(clients)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	updateReq := UpdateClientRequest{
		Name:    "updated-name",
		Enabled: true,
		MTU:     1420,
	}
	updated, err := client.UpdateClient("abc-123", updateReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "updated-name" {
		t.Errorf("expected 'updated-name', got '%s'", updated.Name)
	}
	if updated.MTU != 1420 {
		t.Errorf("expected MTU 1420, got %d", updated.MTU)
	}
}

func TestDeleteClient(t *testing.T) {
	_, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "test"})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/client/abc-123" && r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	err := client.DeleteClient("abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteClientNotFound(t *testing.T) {
	_, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "test"})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/client/nonexistent" && r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	err := client.DeleteClient("nonexistent")
	if err != nil {
		t.Fatalf("expected idempotent delete, got: %v", err)
	}
}

func TestAutoReloginOn401(t *testing.T) {
	callCount := 0
	loginCount := 0
	_, client := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" {
			loginCount++
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "test"})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/client" && r.Method == http.MethodGet {
			callCount++
			if callCount == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Client{{ID: "abc-123", Name: "test", IPv4Address: "10.8.0.2"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	clients, err := client.GetClients()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}
	if loginCount != 1 {
		t.Errorf("expected 1 login attempt, got %d", loginCount)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (1 failed + 1 retry), got %d", callCount)
	}
}
