package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewWithToken(t *testing.T) {
	c := NewWithToken("https://example.com", "my-token")

	if c.baseURL != "https://example.com" {
		t.Errorf("expected baseURL='https://example.com', got '%s'", c.baseURL)
	}
	if c.authToken != "my-token" {
		t.Errorf("expected authToken='my-token', got '%s'", c.authToken)
	}
	if c.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

func TestGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		resp := APIResponse{Success: true, Data: json.RawMessage(`{"id":"123","name":"test"}`)}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewWithToken(server.URL, "test-token")
	resp, err := c.Get("/api/v1/items", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected Success=true")
	}

	var data map[string]string
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to parse Data: %v", err)
	}
	if data["id"] != "123" {
		t.Errorf("expected id='123', got '%s'", data["id"])
	}
	if data["name"] != "test" {
		t.Errorf("expected name='test', got '%s'", data["name"])
	}
}

func TestGet_WithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("page") != "2" {
			t.Errorf("expected query param page=2, got '%s'", q.Get("page"))
		}
		if q.Get("limit") != "10" {
			t.Errorf("expected query param limit=10, got '%s'", q.Get("limit"))
		}
		resp := APIResponse{Success: true, Data: json.RawMessage(`[]`)}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewWithToken(server.URL, "test-token")
	query := map[string]string{"page": "2", "limit": "10"}
	_, err := c.Get("/api/v1/items", query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPost_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type=application/json, got '%s'", ct)
		}

		body, _ := io.ReadAll(r.Body)
		var parsed map[string]string
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if parsed["name"] != "test-item" {
			t.Errorf("expected body name='test-item', got '%s'", parsed["name"])
		}

		resp := APIResponse{Success: true, Data: json.RawMessage(`{"id":"new-123"}`)}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewWithToken(server.URL, "test-token")
	resp, err := c.Post("/api/v1/items", map[string]string{"name": "test-item"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected Success=true")
	}
}

func TestPatch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var parsed map[string]string
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if parsed["name"] != "updated-name" {
			t.Errorf("expected body name='updated-name', got '%s'", parsed["name"])
		}

		resp := APIResponse{Success: true, Data: json.RawMessage(`{"id":"123","name":"updated-name"}`)}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewWithToken(server.URL, "test-token")
	resp, err := c.Patch("/api/v1/items/123", map[string]string{"name": "updated-name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected Success=true")
	}
}

func TestDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/items/123") {
			t.Errorf("expected path to end with /items/123, got '%s'", r.URL.Path)
		}

		resp := APIResponse{Success: true}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewWithToken(server.URL, "test-token")
	resp, err := c.Delete("/api/v1/items/123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected Success=true")
	}
}

func TestRequest_AuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-secret-token" {
			t.Errorf("expected Authorization='Bearer my-secret-token', got '%s'", auth)
		}
		resp := APIResponse{Success: true}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewWithToken(server.URL, "my-secret-token")
	_, err := c.Get("/api/v1/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequest_NoAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("expected no Authorization header, got '%s'", auth)
		}
		resp := APIResponse{Success: true}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewWithToken(server.URL, "")
	_, err := c.Get("/api/v1/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequest_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			Success: false,
			Error:   &APIError{Code: "NOT_FOUND", Message: "Resource not found"},
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewWithToken(server.URL, "test-token")
	_, err := c.Get("/api/v1/missing", nil)
	if err == nil {
		t.Fatal("expected error for unsuccessful response")
	}
}

func TestRequest_APIErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			Success: false,
			Error:   &APIError{Code: "FORBIDDEN", Message: "Access denied"},
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewWithToken(server.URL, "test-token")
	_, err := c.Get("/api/v1/protected", nil)
	if err == nil {
		t.Fatal("expected error for unsuccessful response")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "FORBIDDEN") {
		t.Errorf("expected error to contain 'FORBIDDEN', got '%s'", errMsg)
	}
	if !strings.Contains(errMsg, "Access denied") {
		t.Errorf("expected error to contain 'Access denied', got '%s'", errMsg)
	}
}

func TestRequest_NetworkError(t *testing.T) {
	// Use an unreachable server URL to simulate a network error
	c := NewWithToken("http://127.0.0.1:1", "test-token")
	_, err := c.Get("/api/v1/test", nil)
	if err == nil {
		t.Fatal("expected error for network failure")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("expected error to contain 'request failed', got '%s'", err.Error())
	}
}
