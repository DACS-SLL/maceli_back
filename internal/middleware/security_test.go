package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitAndSpoofedProxyHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(RateLimit(2, 10, time.Minute))
	router.GET("/", func(c *gin.Context) { c.Status(200) })
	for i, want := range []int{200, 200, 429} {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		req.Header.Set("X-Forwarded-For", string(rune('a'+i)))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("request %d got %d want %d", i, rec.Code, want)
		}
		if want == 429 && rec.Header().Get("Retry-After") == "" {
			t.Fatal("missing retry header")
		}
	}
}

func TestOversizedRequestRejectedBeforeHandler(t *testing.T) {
	router := gin.New()
	router.Use(BodyLimit())
	called := false
	router.POST("/", func(c *gin.Context) { called = true; c.Status(200) })
	req := httptest.NewRequest("POST", "/", bytes.NewReader(make([]byte, 129<<10)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 413 || called {
		t.Fatal("oversized request reached the handler")
	}
}

func TestNonNumericIdentifierRejected(t *testing.T) {
	router := gin.New()
	router.Use(NumericID())
	router.GET("/plans/:id", func(c *gin.Context) { c.Status(200) })
	for _, path := range []string{"/plans/1=1", "/plans/0", "/plans/-1"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != 400 {
			t.Fatalf("unsafe identifier accepted: %s", path)
		}
	}
}
