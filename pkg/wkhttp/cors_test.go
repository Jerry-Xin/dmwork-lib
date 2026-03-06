package wkhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORSMiddlewareWithOrigins_AllowAll(t *testing.T) {
	wk := New()
	wk.Use(CORSMiddlewareWithOrigins(nil))
	wk.GET("/test", func(c *Context) {
		c.ResponseOK()
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	wk.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Allow-Origin *, got %s", got)
	}
	// When Allow-Origin is *, Credentials should not be set
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("expected no Credentials header when origin is *, got %s", got)
	}
}

func TestCORSMiddlewareWithOrigins_Whitelist(t *testing.T) {
	allowedOrigins := []string{"https://example.com", "https://app.example.com"}
	wk := New()
	wk.Use(CORSMiddlewareWithOrigins(allowedOrigins))
	wk.GET("/test", func(c *Context) {
		c.ResponseOK()
	})

	// Test allowed origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	wk.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("expected Allow-Origin https://example.com, got %s", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected Allow-Credentials true, got %s", got)
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Errorf("expected Vary Origin, got %s", got)
	}
}

func TestCORSMiddlewareWithOrigins_BlockedOrigin(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	wk := New()
	wk.Use(CORSMiddlewareWithOrigins(allowedOrigins))
	wk.GET("/test", func(c *Context) {
		c.ResponseOK()
	})

	// Test blocked origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	wk.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	// CORS headers should not be set for blocked origins
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no Allow-Origin header for blocked origin, got %s", got)
	}
}

func TestCORSMiddlewareWithOrigins_PreflightAllowed(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	wk := New()
	wk.Use(CORSMiddlewareWithOrigins(allowedOrigins))
	wk.GET("/test", func(c *Context) {
		c.ResponseOK()
	})

	// Test OPTIONS preflight for allowed origin
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	wk.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("expected Allow-Origin https://example.com, got %s", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Allow-Methods header to be set")
	}
}

func TestCORSMiddlewareWithOrigins_PreflightBlocked(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	wk := New()
	wk.Use(CORSMiddlewareWithOrigins(allowedOrigins))
	wk.GET("/test", func(c *Context) {
		c.ResponseOK()
	})

	// Test OPTIONS preflight for blocked origin
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	wk.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d for blocked preflight, got %d", http.StatusForbidden, w.Code)
	}
}

func TestCORSMiddlewareWithOrigins_WildcardInList(t *testing.T) {
	allowedOrigins := []string{"*"}
	wk := New()
	wk.Use(CORSMiddlewareWithOrigins(allowedOrigins))
	wk.GET("/test", func(c *Context) {
		c.ResponseOK()
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	w := httptest.NewRecorder()
	wk.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Allow-Origin *, got %s", got)
	}
}

func TestCORSMiddleware_Deprecated(t *testing.T) {
	// Test that the deprecated CORSMiddleware still works (backwards compatibility)
	wk := New()
	wk.Use(CORSMiddleware())
	wk.GET("/test", func(c *Context) {
		c.ResponseOK()
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	wk.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Allow-Origin *, got %s", got)
	}
}

func TestCORSMiddlewareWithOrigins_NoOriginHeader(t *testing.T) {
	allowedOrigins := []string{"https://example.com"}
	wk := New()
	wk.Use(CORSMiddlewareWithOrigins(allowedOrigins))
	wk.GET("/test", func(c *Context) {
		c.ResponseOK()
	})

	// Request without Origin header (same-origin request)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wk.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
