package api

import (
	//"Dowlet_projects/ecommerce/models"
	"Dowlet_projects/ecommerce/services"
	"github.com/rs/cors"
	"github.com/gorilla/mux"
)


// Handler holds dependencies for API routes
type Handler struct {
	db *services.DBService
}

// NewHandler creates a new API handler
func NewHandler(db *services.DBService) *Handler {
	return &Handler{db: db}
}

// SetupRoutes configures API routes
func (h *Handler) SetupRoutes(router *mux.Router) {
	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// Superadmin-only routes
	superadmin := router.PathPrefix("/api/superadmin").Subrouter()
	superadmin.Use(h.authMiddleware)
	superadmin.HandleFunc("/markets", h.createMarket).Methods("POST", "OPTIONS")
	superadmin.HandleFunc("/markets/{id}", h.deleteMarket).Methods("DELETE", "OPTIONS")
	superadmin.HandleFunc("/categories", h.createCategory).Methods("POST", "OPTIONS")
	superadmin.HandleFunc("/categories/{category_id}", h.deleteCategory).Methods("DELETE", "OPTIONS")
	superadmin.HandleFunc("/banners", h.createBanner).Methods("POST", "OPTIONS")
	superadmin.HandleFunc("/banners/{id}", h.deleteBanner).Methods("DELETE", "OPTIONS")
	superadmin.HandleFunc("/user-messages", h.getUserMessages).Methods("GET", "OPTIONS")
	superadmin.HandleFunc("/user-messages/{id}", h.deleteUserMessage).Methods("DELETE", "OPTIONS")
	superadmin.HandleFunc("/market-messages", h.getMarketMessages).Methods("GET", "OPTIONS")
	superadmin.HandleFunc("/market-messages/{id}", h.deleteMarketMessage).Methods("DELETE", "OPTIONS")
	superadmin.HandleFunc("/users", h.getUsers).Methods("GET", "OPTIONS")
	superadmin.HandleFunc("/users/{id}", h.deleteUser).Methods("DELETE", "OPTIONS")
	superadmin.HandleFunc("/users/{id}", h.updateUserVerified).Methods("PUT")
	superadmin.HandleFunc("/markets/{id}", h.updateMarket).Methods("PUT", "OPTIONS")

	// Market admin routes
	marketAdmin := router.PathPrefix("/api/market").Subrouter()
	marketAdmin.Use(h.authMiddleware)
	marketAdmin.HandleFunc("/products", h.createProduct).Methods("POST", "OPTIONS")
	marketAdmin.HandleFunc("/products/{product_id}", h.updateProduct).Methods("PUT", "OPTIONS")
	marketAdmin.HandleFunc("/products/{product_id}", h.deleteProduct).Methods("DELETE", "OPTIONS")
	marketAdmin.HandleFunc("/products/{id}/thumbnails", h.addProductThumbnails).Methods("POST", "OPTIONS")
	marketAdmin.HandleFunc("/thumbnails/{thumbnail_id}/size", h.addSizeByThumbnail).Methods("POST", "OPTIONS")
	marketAdmin.HandleFunc("/sizes/{size_id}", h.deleteSizeByID).Methods("DELETE", "OPTIONS")
	marketAdmin.HandleFunc("/thumbnails/{thumbnail_id}", h.deleteThumbnail).Methods("DELETE", "OPTIONS")
	marketAdmin.HandleFunc("/markets/{id}/thumbnail", h.uploadMarketThumbnail).Methods("POST", "OPTIONS")
	marketAdmin.HandleFunc("/orders", h.getMarketAdminOrders).Methods("GET", "OPTIONS")
	marketAdmin.HandleFunc("/orders/{order_id}", h.getMarketAdminOrderByID).Methods("GET", "OPTIONS")
	marketAdmin.HandleFunc("/profile", h.getMarketProfile).Methods("GET", "OPTIONS")
	marketAdmin.HandleFunc("/profile", h.updateMarketProfile).Methods("PUT", "OPTIONS")
	marketAdmin.HandleFunc("/orders/{order_id}", h.updateOrderStatus).Methods("PUT", "OPTIONS")
	marketAdmin.HandleFunc("/sizes/{size_id}", h.updateSize).Methods("PUT", "OPTIONS")
	marketAdmin.HandleFunc("/thumbnails/{thumbnail_id}", h.updateThumbnail).Methods("PUT","OPITONS")
	marketAdmin.HandleFunc("/markets", h.getMarketByID).Methods("GET", "OPTIONS")
	marketAdmin.HandleFunc("/messages", h.createMarketMessage).Methods("POST", "OPTIONS")
	marketAdmin.HandleFunc("/orders/{order_id}", h.deleteOrderByID).Methods("DELETE", "OPTIONS")
	// User protected routes
	userProtected := router.PathPrefix("/api").Subrouter()
	userProtected.Use(h.authMiddleware)
	userProtected.HandleFunc("/favorites", h.getUserFavorites).Methods("GET", "OPTIONS")
	userProtected.HandleFunc("/favorites", h.toggleFavorite).Methods("POST", "OPTIONS")
	userProtected.HandleFunc("/cart", h.addToCart).Methods("POST", "OPTIONS")
	userProtected.HandleFunc("/cart", h.getCart).Methods("GET", "OPTIONS")
	userProtected.HandleFunc("/cart/{cart_order_id}", h.deleteCart).Methods("DELETE", "OPTIONS")
	userProtected.HandleFunc("/locations", h.addLocation).Methods("POST", "OPTIONS")
	userProtected.HandleFunc("/locations", h.getLocations).Methods("GET", "OPTIONS")
	userProtected.HandleFunc("/cart/{cart_order_id}/order", h.createOrder).Methods("POST", "OPTIONS")
	userProtected.HandleFunc("/locations/{location_id}", h.deleteLocation).Methods("DELETE", "OPTIONS")
	userProtected.HandleFunc("/cart", h.clearCart).Methods("DELETE", "OPTIONS")
	userProtected.HandleFunc("/cart-product/{size_id}", h.deleteCartBySizeID).Methods("DELETE", "OPTIONS")
	userProtected.HandleFunc("/cart/{size_id}", h.updateCartCountBySizeID).Methods("PUT", "OPTIONS")
	userProtected.HandleFunc("/locations/{location_id}", h.updateLocationByID).Methods("PUT", "OPTIONS")
	userProtected.HandleFunc("/profile", h.getProfile).Methods("GET", "OPTIONS")
	userProtected.HandleFunc("/profile", h.updateProfile).Methods("PUT", "OPTIONS")
	userProtected.HandleFunc("/user-orders", h.getUserOrders).Methods("GET", "OPTIONS")
	userProtected.HandleFunc("/user-orders/{order_id}", h.deleteUserHistory).Methods("PUT", "OPTIONS")
	userProtected.HandleFunc("/messages", h.createMessage).Methods("POST", "OPTIONS")


	// Public routes
	router.HandleFunc("/superadmin/register", h.registerSuperadmin).Methods("POST", "OPTIONS")
	router.HandleFunc("/register", h.register).Methods("POST", "OPTIONS")
	router.HandleFunc("/login", h.login).Methods("POST", "OPTIONS")
	router.HandleFunc("/verify", h.verifyCode).Methods("POST", "OPTIONS")
	router.HandleFunc("/market/login", h.loginMarket).Methods("POST", "OPTIONS")
	router.HandleFunc("/superadmin/login", h.loginSuperadmin).Methods("POST", "OPTIONS")
	router.HandleFunc("/markets", h.getMarkets).Methods("GET", "OPTIONS")
	router.HandleFunc("/markets/{id}/products", h.getMarketProducts).Methods("GET", "OPTIONS")
	router.HandleFunc("/products", h.getAllProducts).Methods("GET", "OPTIONS")
	router.HandleFunc("/products/{id}", h.getProduct).Methods("GET", "OPTIONS")
	router.HandleFunc("/thumbnails", h.getAllThumbnails).Methods("GET", "OPTIONS")
	router.HandleFunc("/categories", h.getCategories).Methods("GET", "OPTIONS")
	router.HandleFunc("/banners", h.getBanners).Methods("GET", "OPTIONS")
	router.HandleFunc("/markets/{id}", h.getMarketByIDALL).Methods("GET", "OPTIONS")
	// Wrap router with CORS
	router.Use(c.Handler)
}