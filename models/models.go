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
	Phone         string  `json:"phone"`
	DeliveryPrice float64 `json:"delivery_price"`
	Name          string  `json:"name"`
	NameRu        string  `json:"name_ru"`
	Location      string  `json:"location"`
	LocationRu    string  `json:"location_ru"`
	ThumbnailURL  string  `json:"thumbnail_url"`
	IsNew         bool    `json:"is_new"`
	IsVIP         bool    `json:"is_vip"`
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
	ID             int         `json:"id"`
	MarketID       int         `json:"market_id"`
	MarketName     string      `json:"market_name"`
	MarketNameRu   string      `json:"market_name_ru"`
	CategoryID     int         `json:"category_id"`
	CategoryName   string      `json:"category_name"`
	CategoryNameRu string      `json:"category_name_ru"`
	ThumbnailURL   string      `json:"thumbnail_url"`
	Name           string      `json:"name"`
	NameRu         string      `json:"name_ru"`
	Price          float64     `json:"price"`
	Discount       float64     `json:"discount"`
	Description    string      `json:"description"`
	DescriptionRu  string      `json:"description_ru"`
	IsNew          bool        `json:"is_new"`
	FinalPrice     float64     `json:"final_price"`
	CreatedAt      string      `json:"created_at"`
	IsFavorite     bool        `json:"is_favorite"`
	Thumbnails     []Thumbnail `json:"thumbnails"`
}

type Thumbnail struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	Color     string `json:"color"`
	ColorRu   string `json:"color_ru"`
	ImageURL  string `json:"image_url"`
	Sizes     []Size `json:"sizes"`
}

type Size struct {
	ID          int     `json:"id"`
	ThumbnailID int     `json:"thumbnail_id"`
	Size        string  `json:"size"`
	Count       int     `json:"count"`
	Price       float64 `json:"price"`
}
type SizeUpdate struct {
	ID          int     `json:"id"`
	Size        string  `json:"size"`
	Count       int     `json:"count"`
	Price       float64 `json:"price"`
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
	NameRu       string `json:"name_ru"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// CartMarket represents a market in a user's cart
type CartMarket struct {
	MarketID      int           `json:"market_id"`
	MarketName    string        `json:"market_name"`
	MarketNameRu  string        `json:"market_name_ru"`
	DeliveryPrice float64       `json:"delivery_price"`
	UserID        int           `json:"user_id"`
	CartID        int           `json:"cart_id"`
	Products      []CartProduct `json:"products"`
}

// CartProduct represents a product in a cart
type CartProduct struct {
	SizeID       int     `json:"size_id"`
	ProductID    int     `json:"product_id"`
	ThumbnailURL string  `json:"thumbnail_url"`
	Name         string  `json:"name"`
	NameRu       string  `json:"name_ru"`
	Price        float64 `json:"price"`
	Discount     float64 `json:"discount"`
	Color        string  `json:"color"`
	ColorRu      string  `json:"color_ru"`
	Size         string  `json:"size"`
	SizePrice    float64 `json:"size_price"`
	Sum          float64 `json:"sum"`
	Count        int     `json:"count"`
}

// CartProductReq represents a product to add to the cart
type CartProductReq struct {
	ProductID   int `json:"product_id"`
	ThumbnailID int `json:"thumbnail_id"`
	SizeID      int `json:"size_id"`
	Count       int `json:"count"`
}

// Location represents a user's saved location
type Location struct {
	ID                int    `json:"id"`
	UserID            int    `json:"user_id"`
	LocationName      string `json:"location_name"`
	LocationNameRu    string `json:"location_name_ru"`
	LocationAddress   string `json:"location_address"`
	LocationAddressRu string `json:"location_address_ru"`
}

// Order represents a user order
type Order struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	CartOrderID int    `json:"cart_order_id"`
	LocationID  int    `json:"location_id"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Notes       string `json:"notes"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type MarketAdminOrder struct {
	ID                int     `json:"id"`
	CartOrderID       int     `json:"cart_order_id"`
	LocationAddress   string  `json:"location_address"`
	LocationAddressRu string  `json:"location_address_ru"`
	Status            string  `json:"status"`
	Name              string  `json:"name"`
	CreatedAt         string  `json:"created_at"`
	Sum               float64 `json:"sum"`
}

// MarketAdminOrderDetail represents a detailed order view for market admins
type MarketAdminOrderDetail struct {
	ID                int                       `json:"id"`
	CartOrderID       int                       `json:"cart_order_id"`
	Name              string                    `json:"name"`
	Phone             string                    `json:"phone"`
	Status            string                    `json:"status"`
	LocationAddress   string                    `json:"location_address"`
	LocationAddressRu string                    `json:"location_address_ru"`
	CreatedAt         string                    `json:"created_at"`
	Sum               float64                   `json:"sum"`
	Products          []MarketAdminOrderProduct `json:"products"`
}

// MarketAdminOrderProduct represents a product in an order for market admins
type MarketAdminOrderProduct struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	NameRu    string  `json:"name_ru"`
	Price     float64 `json:"price"`
	ImageURL  string  `json:"image_url"`
	Discount  float64 `json:"discount"`
	CreatedAt string  `json:"created_at"`
	Size      string  `json:"size"`
	SizePrice float64 `json:"size_price"`
	Count     int64   `json:"count"`
	Sum       float64 `json:"sum"`
}

type CartRequest struct {
	ProductID   int `json:"product_id"`
	ThumbnailID int `json:"thumbnail_id"`
	SizeID      int `json:"size_id"`
	Count       int `json:"count"`
}

// UpdateCartRequest for updating cart entry count
type UpdateCartRequest struct {
	CountChange int `json:"count_change"`
}

// UpdateLocationRequest for updating a location entry
type UpdateLocationRequest struct {
	LocationName      string `json:"location_name,omitempty"`
	LocationNameRu    string `json:"location_name_ru,omitempty"`
	LocationAddress   string `json:"location_address,omitempty"`
	LocationAddressRu string `json:"location_address_ru,omitempty"`
}

type UserProfile struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

type UpdateProfileRequest struct {
	FullName string `json:"full_name,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

type MarketProfile struct {
	ID            int     `json:"id"`
	Phone         string  `json:"phone"`
	DeliveryPrice float64 `json:"delivery_price"`
	Name          string  `json:"name"`
	NameRu        string  `json:"name_ru"`
	Location      string  `json:"location"`
	LocationRu    string  `json:"location_ru"`
	ThumbnailURL  string  `json:"thumbnail_url"`
}

type UpdateMarketProfileRequest struct {
	DeliveryPrice float64 `json:"delivery_price,omitempty"`
	Name          string  `json:"name,omitempty"`
	NameRu        string  `json:"name_ru,omitempty"`
	Location      string  `json:"location,omitempty"`
	LocationRu    string  `json:"location_ru,omitempty"`
}

type Banner struct {
	ID           int    `json:"id"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// CreateBannerRequest for creating a new banner
type CreateBannerRequest struct {
	Description string `json:"description,omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

type OrderUpdate struct {
	ID       int    `json:"id"`
	MarketID int    `json:"market_id"`
	Status   string `json:"status"`
}

type UserOrder struct {
	ID            int     `json:"id"`
	CartOrderID   int     `json:"cart_order_id"`
	MarketName    string  `json:"market_name"`
	MarketNameRu  string  `json:"market_name_ru"`
	ProductName   string  `json:"product_name"`
	ProductNameRu string  `json:"product_name_ru"`
	Size          string  `json:"size"`
	Color         string  `json:"color"`
	ColorRu       string  `json:"color_ru"`
	ImageURL      string  `json:"image_url"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	Sum           float64 `json:"sum"`
	ProductImageURL string `json:"product_image_url"`
}

// CreateMessageRequest for creating a new message
type CreateMessageRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Message  string `json:"message"`
}

// UserMessage represents a user message
type UserMessage struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Message  string `json:"message"`
}


type UpdateSizeRequest struct {
	Count int     `json:"count"`
	Price float64 `json:"price"`
	Size  string  `json:"size"`
}


// CreateMarketMessageRequest for creating a new market message
type CreateMarketMessageRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Message  string `json:"message"`
}

// MarketMessage represents a market message
type MarketMessage struct {
	ID       int    `json:"id"`
	MarketID int    `json:"market_id"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Message  string `json:"message"`
}



// UpdateUserVerifiedRequest represents the request body for updating a user's verified status
type UpdateUserVerifiedRequest struct {
    Verified bool `json:"verified"`
}