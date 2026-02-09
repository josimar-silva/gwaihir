package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware(t *testing.T) {
	middleware := RequestIDMiddleware()
	if middleware == nil {
		t.Fatal("Expected non-nil middleware")
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Create a simple handler to check the request ID
	handlerCalled := false
	handler := func(_ *gin.Context) {
		handlerCalled = true
	}

	middleware(c)
	handler(c)

	if !handlerCalled {
		t.Fatal("Handler was not called")
	}

	// Check that X-Request-ID header is set
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Fatal("Expected X-Request-ID header to be set")
	}
}

func TestRequestLoggingMiddleware(t *testing.T) {
	middleware := RequestLoggingMiddleware()
	if middleware == nil {
		t.Fatal("Expected non-nil middleware")
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Create a simple handler to check the duration
	handlerCalled := false
	handler := func(_ *gin.Context) {
		handlerCalled = true
	}

	middleware(c)
	handler(c)

	if !handlerCalled {
		t.Fatal("Handler was not called")
	}

	// Check that duration is set
	_, exists := c.Get("duration")
	if !exists {
		t.Fatal("Expected duration to be set in context")
	}
}

func TestGetRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(requestIDKey, "test-request-id-123")

	id := GetRequestID(c)
	if id != "test-request-id-123" {
		t.Errorf("Expected request ID 'test-request-id-123', got %s", id)
	}
}

func TestGetRequestIDMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set request ID

	id := GetRequestID(c)
	if id == "" {
		t.Fatal("Expected non-empty request ID for missing case")
	}
}

func TestContextWithRequestID(t *testing.T) {
	ctx := context.Background()
	newCtx := ContextWithRequestID(ctx, "test-id-456")
	if newCtx == nil {
		t.Fatal("Expected non-nil context")
	}

	value := newCtx.Value(contextRequestIDKey)
	if value == nil {
		t.Fatal("Expected request ID to be in context")
	}

	if value != "test-id-456" {
		t.Errorf("Expected request ID 'test-id-456', got %v", value)
	}
}

func TestMiddlewareChain(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(RequestLoggingMiddleware())

	router.GET("/test", func(c *gin.Context) {
		id := GetRequestID(c)
		if id == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no request id"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"request_id": id})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	// Verify X-Request-ID header is set
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("Expected X-Request-ID header to be set")
	}
}
