//go:build test

package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"url-shortener-api/internal/server/http/router"

	"github.com/stretchr/testify/require"
)

// setupE2EServer creates a test server with all dependencies
func setupE2EServer(t *testing.T) *httptest.Server {
	t.Helper()

	// Use test database connection
	db := pg_test.TestDB

	// Create router with all real dependencies
	r := router.NewRouter(db)

	return httptest.NewServer(r)
}

// TestE2E_UserRegistration_LoginFlow tests complete user registration and login
func TestE2E_UserRegistration_LoginFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	server := setupE2EServer(t)
	defer server.Close()

	email := "testuser@example.com"
	password := "SecurePass123!"

	t.Run("Register new user", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    email,
			"password": password,
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(server.URL+"/api/auth/register", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Contains(t, result, "user_id")
	})

	t.Run("Login with registered user", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    email,
			"password": password,
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(server.URL+"/api/auth/login", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Contains(t, result, "access_token")
		require.Contains(t, result, "refresh_token")
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    email,
			"password": "WrongPassword123!",
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(server.URL+"/api/auth/login", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestE2E_URLShortening_RedirectFlow tests complete URL shortening workflow
func TestE2E_URLShortening_RedirectFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	server := setupE2EServer(t)
	defer server.Close()

	var accessToken string
	var shortCode string

	t.Run("Register and login", func(t *testing.T) {
		// Register
		registerPayload := map[string]interface{}{
			"email":    "shortener@example.com",
			"password": "SecurePass123!",
		}
		body, _ := json.Marshal(registerPayload)
		resp, _ := http.Post(server.URL+"/api/auth/register", "application/json", bytes.NewReader(body))
		resp.Body.Close()

		// Login
		loginPayload := map[string]interface{}{
			"email":    "shortener@example.com",
			"password": "SecurePass123!",
		}
		body, _ = json.Marshal(loginPayload)
		resp, err := http.Post(server.URL+"/api/auth/login", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		accessToken = result["access_token"].(string)
	})

	t.Run("Create short URL", func(t *testing.T) {
		payload := map[string]interface{}{
			"url": "https://www.example.com/very/long/path/to/resource",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", server.URL+"/api/urls/shorten", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Contains(t, result, "short_code")
		shortCode = result["short_code"].(string)
	})

	t.Run("Redirect using short code", func(t *testing.T) {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		}

		resp, err := client.Get(server.URL + "/" + shortCode)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusFound, resp.StatusCode)
		location := resp.Header.Get("Location")
		require.Equal(t, "https://www.example.com/very/long/path/to/resource", location)
	})

	t.Run("Verify click count incremented", func(t *testing.T) {
		// Note: This would require a "get URL details" endpoint
		// For now, just verify redirect worked
		t.Skip("Requires GET /api/urls/:shortCode endpoint")
	})
}

// TestE2E_SessionManagement tests session lifecycle
func TestE2E_SessionManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	server := setupE2EServer(t)
	defer server.Close()

	var accessToken string
	var refreshToken string

	t.Run("Register and login", func(t *testing.T) {
		// Register
		registerPayload := map[string]interface{}{
			"email":    "session@example.com",
			"password": "SecurePass123!",
		}
		body, _ := json.Marshal(registerPayload)
		resp, _ := http.Post(server.URL+"/api/auth/register", "application/json", bytes.NewReader(body))
		resp.Body.Close()

		// Login
		loginPayload := map[string]interface{}{
			"email":    "session@example.com",
			"password": "SecurePass123!",
		}
		body, _ = json.Marshal(loginPayload)
		resp, err := http.Post(server.URL+"/api/auth/login", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		accessToken = result["access_token"].(string)
		refreshToken = result["refresh_token"].(string)
	})

	t.Run("List active sessions", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Contains(t, result, "sessions")
	})

	t.Run("Refresh access token", func(t *testing.T) {
		payload := map[string]interface{}{
			"refresh_token": refreshToken,
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(server.URL+"/api/auth/refresh", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Contains(t, result, "access_token")
	})

	t.Run("Logout", func(t *testing.T) {
		req, _ := http.NewRequest("POST", server.URL+"/api/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Cannot use token after logout", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/sessions", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestE2E_ErrorHandling tests API error responses
func TestE2E_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	server := setupE2EServer(t)
	defer server.Close()

	t.Run("Invalid JSON payload", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/api/auth/register", "application/json", bytes.NewReader([]byte("invalid json")))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		payload := map[string]interface{}{
			"email": "test@example.com",
			// Missing password
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(server.URL+"/api/auth/register", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid email format", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    "not-an-email",
			"password": "SecurePass123!",
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(server.URL+"/api/auth/register", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Non-existent endpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/nonexistent")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Unauthorized access", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/sessions", nil)
		// No Authorization header

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
