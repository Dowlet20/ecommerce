package models

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Product represents a product in the e-commerce system
type Product struct {
	ID          int         `json:"id"`
	MarketID    int         `json:"market_id"`
	MarketName  string      `json:"market_name"`
	Name        string      `json:"name"`
	Price       float64     `json:"price"`
	Discount    float64     `json:"discount"`
	Description string      `json:"description"`
	CreatedAt   string      `json:"created_at"`
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
	Count       int    `json:"count"`
}

// Market represents a market in the e-commerce system
type Market struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

// VerificationCode represents a verification code entry
type VerificationCode struct {
	Phone     string    `json:"phone"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	FullName  string    `json:"full_name"`
}

// Claims for JWT
type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

type AuthClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}
