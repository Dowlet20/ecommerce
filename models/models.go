package models

import "time"

// Product represents a product in the e-commerce system
type Product struct {
	ID          int         `json:"id"`
	MarketID    int         `json:"market_id"`
	Name        string      `json:"name"`
	Price       float64     `json:"price"`
	Discount    float64     `json:"discount"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"created_at"`
	IsFavorite  bool        `json:"is_favorite"`
	Thumbnails  []Thumbnail `json:"thumbnails"`
	Sizes       []Size      `json:"sizes"`
}

// Thumbnail represents a product thumbnail
type Thumbnail struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	Color     string `json:"color"`
	ImageURL  string `json:"image_url"`
}

// Size represents a product size related to thumbnail color
type Size struct {
	ID          int    `json:"id"`
	ThumbnailID int    `json:"thumbnail_id"`
	Size        string `json:"size"`
}

// Market represents a market in the e-commerce system
type Market struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// User represents a user in the system
type User struct {
	ID        int    `json:"id"`
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
	Verified  bool   `json:"verified"`
}

// OTP represents an OTP entry
type OTP struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Phone     string    `json:"phone"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}