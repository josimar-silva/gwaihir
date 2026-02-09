package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

const testAPIKey = "test-secret-key-123"

func TestAPIKeyAuthMiddleware_ValidKey(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	expectedKey := testAPIKey
	router.Use(APIKeyAuthMiddleware(expectedKey))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", expectedKey)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 with valid key, got %d", w.Code)
	}
}

func TestAPIKeyAuthMiddleware_MissingKey(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	expectedKey := testAPIKey
	router.Use(APIKeyAuthMiddleware(expectedKey))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Don't set X-API-Key header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 without key, got %d", w.Code)
	}
}

func TestAPIKeyAuthMiddleware_InvalidKey(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	expectedKey := testAPIKey
	router.Use(APIKeyAuthMiddleware(expectedKey))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with invalid key, got %d", w.Code)
	}
}

func TestAPIKeyAuthMiddleware_EmptyKey(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	expectedKey := testAPIKey
	router.Use(APIKeyAuthMiddleware(expectedKey))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with empty key, got %d", w.Code)
	}
}

func TestRouterWithAuth_ValidKey(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouterWithAuth(handler, testAPIKey)

	req := httptest.NewRequest(http.MethodGet, "/machines", nil)
	req.Header.Set("X-API-Key", testAPIKey)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 with valid key, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestRouterWithAuth_MissingKey(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouterWithAuth(handler, testAPIKey)

	req := httptest.NewRequest(http.MethodGet, "/machines", nil)
	// Don't set API key header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 without key, got %d", w.Code)
	}
}

func TestRouterWithAuth_InvalidKey(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouterWithAuth(handler, testAPIKey)

	req := httptest.NewRequest(http.MethodGet, "/machines", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with invalid key, got %d", w.Code)
	}
}

func TestRouterWithAuth_NoAuthRequired(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouterWithAuth(handler, testAPIKey)

	// Health endpoint should work without API key
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for health endpoint without key, got %d", w.Code)
	}
}

func TestRouterWithoutAuth(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	// Create router without API key
	router := NewRouterWithAuth(handler, "")

	// Protected endpoint should work without API key when no key is set
	req := httptest.NewRequest(http.MethodGet, "/machines", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 when no auth is required, got %d", w.Code)
	}
}

func TestRouterWithAuth_ProtectedWoLEndpoint(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouterWithAuth(handler, testAPIKey)

	// POST /wol without key should be rejected
	req := httptest.NewRequest(http.MethodPost, "/wol", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for /wol without key, got %d", w.Code)
	}
}
