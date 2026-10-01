package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"maceli-backend/internal/database"
	"maceli-backend/internal/models"
)

// CI supplies a disposable PostgreSQL service. Never uses DATABASE_URL/production.
func TestPublishAndSessionIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a disposable PostgreSQL database")
	}
	root, err := database.Connect(dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("maceli_test_%d", time.Now().UnixNano())
	if err := root.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	defer root.Exec("DROP SCHEMA " + schema + " CASCADE")
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := database.Connect(parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := InitSite(db); err != nil {
		t.Fatal(err)
	}
	if err := BootstrapAdmin(db, "qa@example.com", "test-only-long-password"); err != nil {
		t.Fatal(err)
	}
	auth := NewAuthHandler(db)
	site := NewSiteHandler(db, &testUploader{})
	router := gin.New()
	router.POST("/login", auth.Login)
	router.GET("/public", site.Public)
	admin := router.Group("/admin")
	admin.Use(auth.RequireSession())
	admin.PUT("/site", site.Save)
	admin.POST("/publish", site.Publish)
	admin.POST("/logout", auth.Logout)
	token := ""
	request := func(method, path string, body interface{}, want int) map[string]interface{} {
		t.Helper()
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, rec.Code, want, rec.Body.String())
		}
		var data map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &data)
		return data
	}
	request("PUT", "/admin/site", map[string]interface{}{}, 401)
	request("POST", "/login", map[string]string{"email": "qa@example.com", "password": "incorrect"}, 401)
	result := request("POST", "/login", map[string]string{"email": "qa@example.com", "password": "test-only-long-password"}, 200)
	token = result["token"].(string)
	asset := models.MediaAsset{URL: "https://res.cloudinary.com/test/image/upload/menu.png"}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatal(err)
	}
	start := time.Now().In(time.FixedZone("Lima", -5*3600))
	content := models.SiteContent{Menu: models.WeeklyMenu{Title: "Carta de prueba", Start: start.Format("2006-01-02"), End: start.AddDate(0, 0, 6).Format("2006-01-02"), Pages: []models.SiteImage{{URL: asset.URL, Alt: "Opciones semanales"}}}}
	request("PUT", "/admin/site", map[string]interface{}{"content": content, "version": 1}, 200)
	if request("GET", "/public", nil, 200)["menu_status"] != "empty" {
		t.Fatal("draft leaked publicly")
	}
	request("PUT", "/admin/site", map[string]interface{}{"content": content, "version": 1}, 409)
	request("POST", "/admin/publish", map[string]int{"version": 1}, 409)
	request("POST", "/admin/publish", map[string]int{"version": 2}, 200)
	if request("GET", "/public", nil, 200)["menu_status"] != "current" {
		t.Fatal("published menu missing")
	}
	content.Menu.Pages[0].URL = "https://untrusted.example/image.png"
	request("PUT", "/admin/site", map[string]interface{}{"content": content, "version": 3}, 400)
	if request("GET", "/public", nil, 200)["menu_status"] != "current" {
		t.Fatal("failed save changed publication")
	}
	request("POST", "/admin/logout", nil, 204)
	request("PUT", "/admin/site", map[string]interface{}{}, 401)
	result = request("POST", "/login", map[string]string{"email": "qa@example.com", "password": "test-only-long-password"}, 200)
	token = result["token"].(string)
	if err := db.Model(&models.AdminSession{}).Where("token_hash = ?", hashToken(token)).Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	request("PUT", "/admin/site", map[string]interface{}{}, 401)
	result = request("POST", "/login", map[string]string{"email": "qa@example.com", "password": "test-only-long-password"}, 200)
	token = result["token"].(string)
	if err := db.Model(&models.AdminUser{}).Where("email = ?", "qa@example.com").Update("active", false).Error; err != nil {
		t.Fatal(err)
	}
	request("PUT", "/admin/site", map[string]interface{}{}, 401)
}
