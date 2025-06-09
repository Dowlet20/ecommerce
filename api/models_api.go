package api

import (
	"Dowlet_projects/ecommerce/models"
)

type CartRequest struct {
	MarketID int                     `json:"market_id"`
	Products []models.CartProductReq `json:"products"`
}

type LocationRequest struct {
	LocationName      string `json:"location_name"`
	LocationNameRu    string `json:"location_name_ru"`
	LocationAddress   string `json:"location_address"`
	LocationAddressRu string `json:"location_address_ru"`
}

// OrderRequest for submitting an order
type OrderRequest struct {
	LocationID int    `json:"location_id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Notes      string `json:"notes"`
}

type ProductRequest struct {
	CategoryID    int     `json:"category_id"`
	Name          string  `json:"name"`
	NameRu        string  `json:"name_ru"`
	Price         float64 `json:"price"`
	Discount      float64 `json:"discount,omitempty"`
	Description   string  `json:"description,omitempty"`
	DescriptionRu string  `json:"description_ru,omitempty"`
	IsActive      bool    `json:"is_active"`
}

// SizeRequest for adding a size
type SizeRequest struct {
	Size  string  `json:"size"`
	Count int     `json:"count"`
	Price float64 `json:"price"`
}

// FavoriteRequest for toggling favorite
type FavoriteRequest struct {
	ProductID int `json:"product_id"`
}

// MarketRequest for market login
type MarketRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// SuperadminRequest for superadmin login
type SuperadminRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserRegisterRequest for user registration
type UserRegisterRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

// UserOTPRequest for OTP verification
type UserOTPRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

// CreateMarketRequest for superadmin creating market
type CreateMarketRequest struct {
	Name         string `json:"name"`
	ThumbnailURL string `json:"thumbnail_url"`
	Username     string `json:"username"`
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	Password     string `json:"password"`
}
