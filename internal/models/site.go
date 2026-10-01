package models

import "time"

type AdminUser struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `json:"-"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

type AdminSession struct {
	TokenHash string    `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
}

type SiteImage struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}

type WeeklyMenu struct {
	Title       string      `json:"title"`
	Start       string      `json:"start"`
	End         string      `json:"end"`
	Description string      `json:"description"`
	Pages       []SiteImage `json:"pages"`
}

type SiteContent struct {
	Hero   []SiteImage `json:"hero"`
	Dishes []SiteImage `json:"dishes"`
	Plans  []SiteImage `json:"plans"`
	About  []SiteImage `json:"about"`
	Menu   WeeklyMenu  `json:"menu"`
}

// A single row holds both versions so publishing is an atomic database update.
type SiteState struct {
	ID          uint   `gorm:"primaryKey"`
	Version     uint   `gorm:"not null"`
	Draft       string `gorm:"type:text;not null"`
	Published   string `gorm:"type:text;not null"`
	PublishedAt *time.Time
	UpdatedAt   time.Time
	UpdatedBy   uint
}

type MediaAsset struct {
	ID         uint   `gorm:"primaryKey"`
	URL        string `gorm:"uniqueIndex;not null"`
	CreatedAt  time.Time
	UploadedBy uint
}
