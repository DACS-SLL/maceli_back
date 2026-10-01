package handlers

import (
	"maceli-backend/internal/models"
	"testing"
	"time"
)

func TestWeeklyValidityUsesLimaCalendar(t *testing.T) {
	menu := models.WeeklyMenu{Start: "2026-09-28", End: "2026-10-04", Pages: []models.SiteImage{{URL: "image", Alt: "Carta"}}}
	for _, tc := range []struct{ at, want string }{
		{"2026-09-28T04:59:59Z", "upcoming"},
		{"2026-09-28T05:00:00Z", "current"},
		{"2026-10-05T04:59:59Z", "current"},
		{"2026-10-05T05:00:00Z", "expired"},
	} {
		at, _ := time.Parse(time.RFC3339, tc.at)
		if got := menuStatus(menu, at); got != tc.want {
			t.Errorf("%s: got %s, want %s", tc.at, got, tc.want)
		}
	}
	menu.Pages = nil
	if menuStatus(menu, time.Now()) != "empty" {
		t.Fatal("withdrawn menus must be empty")
	}
}

func TestContentRequiresWeeklyDatesAndAccessibleImages(t *testing.T) {
	content := models.SiteContent{Menu: models.WeeklyMenu{Title: "Carta semanal", Start: "2026-09-28", End: "2026-10-04", Pages: []models.SiteImage{{URL: "image", Alt: "Opciones de la semana"}}}}
	if err := validateContent(content); err != nil {
		t.Fatal(err)
	}
	content.Menu.End = "2026-09-28"
	if validateContent(content) == nil {
		t.Fatal("daily menus must be rejected")
	}
	content.Menu.End = "2026-10-04"
	content.Menu.Pages[0].Alt = " "
	if validateContent(content) == nil {
		t.Fatal("missing image descriptions must be rejected")
	}
	content.Menu.Pages = nil
	content.Hero = make([]models.SiteImage, 3)
	if validateContent(content) == nil {
		t.Fatal("too many hero images accepted")
	}
}

func TestCredentials(t *testing.T) {
	if validCredentials("admin@example.com", "short") {
		t.Fatal("short password accepted")
	}
	if validCredentials("Person <admin@example.com>", "long-enough-password") {
		t.Fatal("display name accepted as login email")
	}
	if !validCredentials("admin@example.com", "long-enough-password") {
		t.Fatal("valid credentials rejected")
	}
	if hashToken("one") == hashToken("two") || len(hashToken("one")) != 64 {
		t.Fatal("invalid session hashing")
	}
}
