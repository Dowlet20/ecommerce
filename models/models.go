package models

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	MarketID int    `json:"market_id,omitempty"` // For market admins
	Role     string `json:"role"`                // superadmin, market_admin, user
	jwt.StandardClaims
}

type Market struct {
	ID            int     `json:"id"`
	Username      string  `json:"username"`
	FullName      string  `json:"full_name"`
	Phone         string  `json:"phone"`
	DeliveryPrice float64 `json:"delivery_price"`
	Name          string  `json:"name"`
	ThumbnailURL  string  `json:"thumbnail_url"`
}

type User struct {
	ID        int       `json:"id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
}

type Superadmin struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

type VerificationCode struct {
	Phone     string    `json:"phone"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	FullName  string    `json:"full_name"`
}

type Product struct {
	ID          int         `json:"id"`
	MarketID    int         `json:"market_id"`
	MarketName  string      `json:"market_name"`
	CategoryID  int         `json:"category_id"`
	Name        string      `json:"name"`
	Price       float64     `json:"price"`
	Discount    float64     `json:"discount"`
	Description string      `json:"description"`
	CreatedAt   string      `json:"created_at"`
	IsFavorite  bool        `json:"is_favorite"`
	Thumbnails  []Thumbnail `json:"thumbnails"`
}

type Thumbnail struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	Color     string `json:"color"`
	ImageURL  string `json:"image_url"`
	Sizes     []Size `json:"sizes"`
}

type Size struct {
	ID          int    `json:"id"`
	ThumbnailID int    `json:"thumbnail_id"`
	Size        string `json:"size"`
	Count       int    `json:"count"`
}

type AuthClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

// SuperadminRegisterRequest for superadmin registration
type SuperadminRegisterRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type Category struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// CartMarket represents a market in a user's cart
type CartMarket struct {
	MarketName    string        `json:"market_name"`
	DeliveryPrice float64       `json:"delivery_price"`
	UserID        int           `json:"user_id"`
	CartID        int           `json:"cart_id"`
	Products      []CartProduct `json:"products"`
}

// CartProduct represents a product in a cart
type CartProduct struct {
	ThumbnailURL string  `json:"thumbnail_url"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Discount     float64 `json:"discount"`
	Color        string  `json:"color"`
	Size         string  `json:"size"`
	Count        int     `json:"count"`
}
