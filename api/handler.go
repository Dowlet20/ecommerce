package api

import (
	//"context"
	//"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	//"regexp"
	"strconv"
	"strings"
	"time"

	//"github.com/dgrijalva/jwt-go"
	"Dowlet_projects/ecommerce/models"
	"Dowlet_projects/ecommerce/services"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
	marketAdmin.HandleFunc("/orders/{cart_order_id}", h.getMarketAdminOrderByID).Methods("GET", "OPTIONS")

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
	router.HandleFunc("/markets/{id}", h.getMarketByID).Methods("GET", "OPTIONS")
	router.HandleFunc("/categories", h.getCategories).Methods("GET", "OPTIONS")
	// Wrap router with CORS
	router.Use(c.Handler)
}

// ProductRequest for creating/updating a product
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

// getMarkets returns all markets
// @Summary Get all markets
// @Description Retrieves a list of all markets. Requires JWT authentication.
// @Tags Markets
// @Accept json
// @Produce json
// @Router /markets [get]
func (h *Handler) getMarkets(w http.ResponseWriter, r *http.Request) {
	markets, err := h.db.GetMarkets()
	fmt.Println(err)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, markets)
}

// createMarket creates a new market with admin account and optional thumbnail
// @Summary Create a new market
// @Description Creates a new market with an admin account and optional thumbnail image. Requires superadmin JWT authentication.
// @Tags Markets
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Market name"
// @Param name_ru formData string true "Market name russian"
// @Param phone formData string true "Admin phone"
// @Param delivery_price formData float64 true "Delivery price"
// @Param password formData string true "Admin password"
// @Param thumbnail formData file false "Thumbnail image"
// @Security BearerAuth
// @Router /api/superadmin/markets [post]
func (h *Handler) createMarket(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.Role != "superadmin" {
		fmt.Println(claims)
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := r.ParseMultipartForm(10 << 20) // Max 10 MB
	if err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing form")
		return
	}

	name := r.FormValue("name")
	name_ru := r.FormValue("name_ru")
	phone := r.FormValue("phone")
	password := r.FormValue("password")
	// Get the form value as a string
	deliveryPriceStr := r.FormValue("delivery_price")

	// Convert the string to float64
	deliveryPrice, err := strconv.ParseFloat(deliveryPriceStr, 64)
	if err != nil {
		// Handle the error (e.g., invalid input)
		respondError(w, http.StatusInternalServerError, "error converting string to float64")
		return
	}

	if name == "" || name_ru == "" || phone == "" || password == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	var thumbnailURL string
	file, handler, err := r.FormFile("thumbnail")
	if err == nil {
		defer file.Close()

		uploadDir := os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			uploadDir = "./uploads"
		}
		marketsDir := filepath.Join(uploadDir, "markets")
		if err := os.MkdirAll(marketsDir, 0755); err != nil {
			respondError(w, http.StatusInternalServerError, "Error creating uploads directory")
			return
		}

		filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
		filePath := filepath.Join(marketsDir, filename)
		out, err := os.Create(filePath)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Error saving file")
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Error copying file")
			return
		}

		thumbnailURL = filepath.Join("/uploads/markets", filename)
		thumbnailURL = strings.ReplaceAll(thumbnailURL, string(filepath.Separator), "/")
	} else if err != http.ErrMissingFile {
		respondError(w, http.StatusBadRequest, "Error retrieving file")
		return
	}

	createdUsername, marketID, err := h.db.CreateMarket(name, name_ru, thumbnailURL, phone, password, deliveryPrice)
	if err != nil {
		if err.Error() == "username or phone already exists" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"username":  createdUsername,
		"password":  password,
		"market_id": marketID,
	})
}

// deleteMarket deletes a market
// @Summary Delete a market
// @Description Deletes a market, its products, and thumbnails. Requires superadmin JWT authentication.
// @Tags Markets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Market ID"
// @Router /api/superadmin/markets/{id} [delete]
func (h *Handler) deleteMarket(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.Role != "superadmin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	marketIDStr := vars["id"]
	marketID, err := strconv.Atoi(marketIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid market ID")
		return
	}

	err = h.db.DeleteMarket(marketID)
	if err != nil {
		if err.Error() == "market not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Market, products, and thumbnails deleted successfully"})
}

// getMarketProducts retrieves products for the market admin
// @Summary Get market products
// @Description Retrieves a paginated list of products for the market admin's market. Requires JWT authentication.
// @Tags Products
// @Produce json
// @Param page query integer false "Page number (default: 1)"
// @Param limit query integer false "Items per page (default: 10)"
// @Router /products [get]
func (h *Handler) getMarketProducts(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	products, err := h.db.GetMarketProducts(claims.MarketID, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, products)
}

// getMarketByID retrieves a market and its products
// @Summary Get market by ID
// @Description Retrieves market details and its products by market ID.
// @Tags Markets
// @Produce json
// @Param id path string true "Market ID"
// @Router /markets/{id} [get]
func (h *Handler) getMarketByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketIDStr := vars["id"]
	marketID, err := strconv.Atoi(marketIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid market ID")
		return
	}

	market, products, err := h.db.GetMarketByID(marketID)
	if err != nil {
		if err.Error() == "market not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"market":   market,
		"products": products,
	})
}

// getAllProducts retrieves all paginated products with optional category filter
// @Summary Get all products
// @Description Retrieves paginated products with optional category, name search, or random selection for homepage
// @Tags Products
// @Produce json
// @Param category_id query string false "Category ID"
// @Param duration query integer false "Duration day for new (default: 7)"
// @Param page query integer false "Page number (default: 1, ignored if random=true)"
// @Param limit query integer false "Items per page (default: 10, ignored if random=true)"
// @Param search query string false "Search by product name"
// @Param random query boolean false "Return 10 random products (default: false)"
// @Router /products [get]
func (h *Handler) getAllProducts(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := r.URL.Query().Get("category_id")
	var categoryID int
	if categoryIDStr != "" {
		var err error
		categoryID, err = strconv.Atoi(categoryIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid category ID")
			return
		}
	}

	durationStr := r.URL.Query().Get("duration")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")
	randomStr := r.URL.Query().Get("random")

	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 {
		duration = 7
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	random, err := strconv.ParseBool(randomStr)
	if err != nil {
		random = false
	}

	if random {
		page = 1
		limit = 10
	}

	products, err := h.db.GetPaginatedProducts(categoryID, duration, page, limit, search, random)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, products)
}

// getProduct returns a single product by ID
// @Summary Get a product by ID
// @Description Retrieves a single product by its ID. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Router /products/{id} [get]
func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	product, err := h.db.GetProduct(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, product)
}

// // registerUser registers a user
// // @Summary Register user
// // @Description Registers a user and sends OTP.
// // @Tags Authentication
// // @Accept json
// // @Produce json
// // @Param user body UserRegisterRequest true "User details"
// // @Router /users/register [post]
// func (h *Handler) registerUser(w http.ResponseWriter, r *http.Request) {
// 	var req UserRegisterRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
// 		return
// 	}

// 	if req.FullName == "" || req.Phone == "" {
// 		respondError(w, http.StatusBadRequest, "Missing required fields")
// 		return
// 	}

// 	verifID, otp, err := h.db.RegisterUser(req.FullName, req.Phone)
// 	if err != nil {
// 		if err.Error() == "phone number already registered" {
// 			respondError(w, http.StatusConflict, err.Error())
// 			return
// 		}
// 		respondError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	// TODO: Send OTP via SMS
// 	fmt.Printf("OTP for %s: %s\n", req.Phone, otp)

// 	respondJSON(w, http.StatusOK, map[string]string{"message": "OTP sent", "verification_id": strconv.Itoa(verifID)})
// }

// // verifyUserOTP verifies OTP
// // @Summary Verify user OTP
// // @Description Verifies OTP and logs in the user with a JWT token.
// // @Tags Authentication
// // @Accept json
// // @Produce json
// // @Param otp body UserOTPRequest true "Phone and OTP"
// // @Router /users/verify [post]
// func (h *Handler) verifyUserOTP(w http.ResponseWriter, r *http.Request) {
// 	var req UserOTPRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
// 		return
// 	}

// 	if req.Phone == "" || req.OTP == "" {
// 		respondError(w, http.StatusBadRequest, "Missing required fields")
// 		return
// 	}

// 	userID, err := h.db.VerifyUserOTP(req.Phone, req.OTP)
// 	if err != nil {
// 		respondError(w, http.StatusUnauthorized, err.Error())
// 		return
// 	}

// 	token, err := h.generateJWT(userID, 0, "user")
// 	if err != nil {
// 		respondError(w, http.StatusInternalServerError, "Failed to generate token")
// 		return
// 	}

// 	respondJSON(w, http.StatusOK, map[string]string{"token": token})
// }

// // authMiddleware authenticates JWT
// func (h *Handler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		tokenStr := r.Header.Get("Authorization")
// 		if len(tokenStr) <= 7 || tokenStr[:7] != "Bearer " {
// 			respondError(w, http.StatusUnauthorized, "Missing or invalid token")
// 			return
// 		}
// 		tokenStr = tokenStr[7:]

// 		claims := &models.Claims{}
// 		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
// 			return []byte("Salam@123"), nil
// 		})
// 		if err != nil || !token.Valid {
// 			respondError(w, http.StatusUnauthorized, "Invalid token")
// 			return
// 		}

// 		ctx := context.WithValue(r.Context(), "claims", claims)
// 		next(w, r.WithContext(ctx))
// 	}
// }

// generateJWT creates a JWT
func (h *Handler) generateJWT(userID, marketID int, role string) (string, error) {
	claims := &models.Claims{
		UserID:   userID,
		MarketID: marketID,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-secret-key"))
}

// ProductRequest represents the JSON request body for creating a product
//
//	type ProductRequest struct {
//		MarketID    string  `json:"market_id"`
//		Name        string  `json:"name"`
//		Price       string  `json:"price"`
//		Discount    string  `json:"discount,omitempty"`
//		Description string  `json:"description,omitempty"`
//	}
//
// createProduct creates a new product
// @Summary Create a new product
// @Description Creates a new product for the market admin's market with an optional thumbnail. Requires JWT authentication.
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Product name"
// @Param name_ru formData string false "Product name in Russian"
// @Param price formData number true "Product price"
// @Param discount formData number false "Product discount (0-100)"
// @Param description formData string false "Product description"
// @Param description_ru formData string false "Product description in Russian"
// @Param category_id formData int true "Category ID"
// @Param is_active formData boolean false "Is product active"
// @Param thumbnail formData file false "Thumbnail image"
// @Router /api/market/products [post]
func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing form data")
		return
	}

	// Get form fields
	req := ProductRequest{
		Name:          r.FormValue("name"),
		NameRu:        r.FormValue("name_ru"),
		Description:   r.FormValue("description"),
		DescriptionRu: r.FormValue("description_ru"),
	}
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid price")
		return
	}
	req.Price = price

	discount, err := strconv.ParseFloat(r.FormValue("discount"), 64)
	if err != nil && r.FormValue("discount") != "" {
		respondError(w, http.StatusBadRequest, "Invalid discount")
		return
	}
	req.Discount = discount

	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}
	req.CategoryID = categoryID

	isActive, err := strconv.ParseBool(r.FormValue("is_active"))
	if err != nil && r.FormValue("is_active") != "" {
		respondError(w, http.StatusBadRequest, "Invalid is_active value")
		return
	}
	req.IsActive = isActive

	// Validate required fields
	if req.Name == "" || req.Price == 0 || req.CategoryID == 0 {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	var urlPath string
	var filePath string
	var filename string
	file, fileHeader, err := r.FormFile("thumbnail")
	if err == nil {
		defer file.Close()

		// Validate file type
		mimeType := fileHeader.Header.Get("Content-Type")
		if mimeType != "image/jpeg" && mimeType != "image/png" {
			respondError(w, http.StatusBadRequest, "Only JPEG or PNG images are allowed")
			return
		}

		// Generate unique filename
		timestamp := time.Now().UnixNano()
		filename = fmt.Sprintf("%d-%s", timestamp, strings.ReplaceAll(fileHeader.Filename, " ", "_"))
		filePath = filepath.Join("uploads", "products", "main", filename)
		urlPath = fmt.Sprintf("/uploads/products/main/%s", filename)

		// Create directories if they don't exist
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create upload directory")
			return
		}

		// Save file
		dst, err := os.Create(filePath)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to save thumbnail")
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to save thumbnail")
			return
		}
	} else if err != http.ErrMissingFile {
		respondError(w, http.StatusBadRequest, "Error accessing thumbnail")
		return
	}

	// Create product
	productID, err := h.db.CreateProduct(claims.MarketID, req.CategoryID, req.Name, req.NameRu,
		req.Price, req.Discount, req.Description, req.DescriptionRu, req.IsActive, urlPath,
		filePath, filename)

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"product_id": productID})
}

// deleteSizeByID deletes a size
// @Summary Delete a size by ID
// @Description Deletes a size by its ID. Requires JWT authentication.
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param size_id path string true "Size ID"
// @Router /api/market/sizes/{size_id} [delete]
func (h *Handler) deleteSizeByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sizeID := vars["size_id"]
	if sizeID == "" {
		respondError(w, http.StatusBadRequest, "Missing size ID")
		return
	}

	err := h.db.DeleteSizeByID(claims.MarketID, sizeID)
	if err != nil {
		if err.Error() == "size not found or unauthorized" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Size deleted successfully"})
}

// uploadMarketThumbnail uploads a thumbnail for a market
// @Summary Upload market thumbnail
// @Description Uploads a thumbnail image for a specific market. Requires JWT authentication.
// @Tags Markets
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Market ID"
// @Param thumbnail formData file true "Thumbnail image"
// @Router /api/market/markets/{id}/thumbnail [post]
func (h *Handler) uploadMarketThumbnail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing form")
		return
	}

	file, handler, err := r.FormFile("thumbnail")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Error retrieving file")
		return
	}
	defer file.Close()

	if err := os.MkdirAll("uploads/markets", 0755); err != nil {
		respondError(w, http.StatusInternalServerError, "Error creating uploads directory")
		return
	}

	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
	filePath := filepath.Join("uploads/markets", filename)
	out, err := os.Create(filePath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error saving file")
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error copying file")
		return
	}

	thumbnailURL := "/uploads/markets" + filename
	err = h.db.UpdateMarketThumbnail(id, thumbnailURL)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Thumbnail uploaded successfully"})
}

// updateProduct updates a product
// @Summary Update a product
// @Description Updates a product for the market admin's market. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product_id path string true "Product ID"
// @Param product body ProductRequest true "Product details"
// @Router /api/market/products/{product_id} [put]
func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	productIDStr := vars["product_id"]
	fmt.Println(productIDStr)
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		fmt.Println(err)
		respondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.Name == "" || req.Price == 0 {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	err = h.db.UpdateProduct(claims.MarketID, productID, req.Name, req.Price, req.Discount, req.Description)
	if err != nil {
		if err.Error() == "product not found or unauthorized" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Product updated successfully"})
}

// addProductThumbnails adds multiple thumbnails to a product
// @Summary Add thumbnails to a product
// @Description Adds multiple thumbnail images with associated colors to a product by ID. Requires JWT authentication.
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param colors formData string true "Comma-separated list of colors for thumbnails"
// @Param colors_ru formData string true "Comma-separated list of colors_ru for thumbnails"
// @Param thumbnails formData file true "Thumbnail images"
// @Router /api/market/products/{id}/thumbnails [post]
func (h *Handler) addProductThumbnails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := r.ParseMultipartForm(10 << 20) // Max 10 MB
	if err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing form")
		return
	}

	colors := r.FormValue("colors")
	colors_ru := r.FormValue("colors_ru")
	if colors == "" {
		respondError(w, http.StatusBadRequest, "Colors are required")
		return
	}
	if colors_ru == "" {
		respondError(w, http.StatusBadRequest, "Colors_ru are required")
		return
	}
	colorList := strings.Split(colors, ",")
	colorList_ru := strings.Split(colors_ru, ",")
	files := r.MultipartForm.File["thumbnails"]
	if len(files) == 0 {
		respondError(w, http.StatusBadRequest, "At least one thumbnail is required")
		return
	}
	if len(files) != len(colorList) {
		respondError(w, http.StatusBadRequest, "Number of thumbnails must match number of colors")
		return
	}
	if len(files) != len(colorList_ru) {
		respondError(w, http.StatusBadRequest, "Number of thumbnails must match number of colors_ru")
		return
	}

	// Create uploads directory
	if err := os.MkdirAll("uploads/products/"+id, 0755); err != nil {
		respondError(w, http.StatusInternalServerError, "Error creating directory")
		return
	}

	var thumbnails []services.ThumbnailData
	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Error opening file")
			return
		}
		defer file.Close()

		filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), fileHeader.Filename)
		filePath := filepath.Join("uploads/products/"+id, filename)
		out, err := os.Create(filePath)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Error saving file")
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Error copying file")
			return
		}

		// Parse product ID
		productID, err := strconv.Atoi(id)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid product ID")
			return
		}
		ImageURL := filepath.Join("/uploads/products/"+id, filename)
		ImageURL = strings.ReplaceAll(ImageURL, string(filepath.Separator), "/")

		thumbnails = append(thumbnails, services.ThumbnailData{
			ProductID: productID,
			Color:     colorList[i],
			ColorRu:   colorList_ru[i],
			ImageURL:  ImageURL,
		})
	}

	// Save thumbnails to database
	thumbnail_id, err := h.db.CreateThumbnails(thumbnails)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int64 {
		"thumbnail_id":thumbnail_id,
	})
}

// deleteThumbnail deletes a thumbnail
// @Summary Delete a thumbnail
// @Description Deletes a thumbnail by its ID. Requires JWT authentication.
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param thumbnail_id path string true "Thumbnail ID"
// @Router /api/market/thumbnails/{thumbnail_id} [delete]
func (h *Handler) deleteThumbnail(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	thumbnailID := vars["thumbnail_id"]
	if thumbnailID == "" {
		respondError(w, http.StatusBadRequest, "Missing thumbnail ID")
		return
	}

	err := h.db.DeleteThumbnail(claims.MarketID, thumbnailID)
	if err != nil {
		if err.Error() == "thumbnail not found or unauthorized" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Thumbnail deleted successfully"})
}

// getAllThumbnails retrieves all thumbnails with product information
// @Summary Get all thumbnails with product details
// @Description Retrieves a list of all thumbnails with associated product information. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Router /thumbnails [get]
func (h *Handler) getAllThumbnails(w http.ResponseWriter, r *http.Request) {
	thumbnails, err := h.db.GetAllThumbnailsWithProducts()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, thumbnails)
}

// // SizeRequest represents the JSON request body for adding a size
// type SizeRequest struct {
// 	Size  string `json:"size"`
// 	Count int    `json:"count"`
// }

// getUserFavorites returns favorite products
// @Summary Get user's favorite products
// @Description Retrieves a paginated list of favorite products. Requires JWT authentication.
// @Tags Favorites
// @Produce json
// @Security BearerAuth
// @Param page query integer false "Page number (default: 1)"
// @Param limit query integer false "Items per page (default: 10)"
// @Router /api/favorites [get]
func (h *Handler) getUserFavorites(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	products, err := h.db.GetUserFavoriteProducts(claims.UserID, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, products)
}

// toggleFavorite toggles a favorite product
// @Summary Toggle favorite product
// @Description Adds or removes a product from favorites. Requires JWT authentication.
// @Tags Favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param favorite body FavoriteRequest true "Product ID"
// @Router /api/favorites [post]
func (h *Handler) toggleFavorite(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.ProductID <= 0 {
		respondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	isFavorite, err := h.db.ToggleFavoriteProduct(claims.UserID, req.ProductID)
	if err != nil {
		if err.Error() == "product not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"is_favorite": isFavorite})
}

// loginMarket authenticates a market admin
// @Summary Login market admin
// @Description Authenticates a market admin and returns a JWT token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body MarketRequest true "Market admin credentials"
// @Router /market/login [post]
func (h *Handler) loginMarket(w http.ResponseWriter, r *http.Request) {
	var req MarketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.Phone == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Missing credentials")
		return
	}

	userID, marketID, err := h.db.AuthenticateMarket(req.Phone, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := h.generateJWT(userID, marketID, "market_admin")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"token": token})
}

// loginSuperadmin authenticates a superadmin
// @Summary Login superadmin
// @Description Authenticates a superadmin and returns a JWT token.
// @Tags Superadmin
// @Accept json
// @Produce json
// @Param credentials body SuperadminRequest true "Superadmin credentials"
// @Router /superadmin/login [post]
func (h *Handler) loginSuperadmin(w http.ResponseWriter, r *http.Request) {
	var req SuperadminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Missing credentials")
		return
	}

	userID, err := h.db.AuthenticateSuperadmin(req.Username, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := h.generateJWT(userID, 0, "superadmin")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"token": token})
}

// addSizeByThumbnail adds a size with count
// @Summary Add a size by thumbnail ID
// @Description Adds a size with count linked to a thumbnail. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param thumbnail_id path string true "Thumbnail ID"
// @Param size body SizeRequest true "Size and count"
// @Router /api/market/thumbnails/{thumbnail_id}/size [post]
func (h *Handler) addSizeByThumbnail(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	thumbnailID := vars["thumbnail_id"]
	if thumbnailID == "" {
		respondError(w, http.StatusBadRequest, "Missing thumbnail ID")
		return
	}

	var req SizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.Size == "" || req.Count <= 0 {
		respondError(w, http.StatusBadRequest, "Missing or invalid fields")
		return
	}

	err := h.db.CreateSizeByThumbnailID(claims.MarketID, thumbnailID, req.Size, req.Count, req.Price)
	if err != nil {
		if err.Error() == "thumbnail not found or unauthorized" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Size added successfully"})
}

// deleteProduct deletes a product
// @Summary Delete a product
// @Description Deletes a product and its thumbnails for the market admin's market. Requires JWT authentication.
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param product_id path string true "Product ID"
// @Router /api/market/products/{product_id} [delete]
func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	productIDStr := vars["product_id"]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	err = h.db.DeleteProduct(claims.MarketID, productID)
	if err != nil {
		if err.Error() == "product not found or unauthorized" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Product and thumbnails deleted successfully"})
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Printf("Failed to encode JSON: %v\n", err)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// registerSuperadmin registers a new superadmin
// @Summary Register a new superadmin
// @Description Registers a new superadmin with username, full name, phone, and password.
// @Tags Superadmin
// @Accept json
// @Produce json
// @Param superadmin body models.SuperadminRegisterRequest true "Superadmin details"
// @Router /superadmin/register [post]
func (h *Handler) registerSuperadmin(w http.ResponseWriter, r *http.Request) {
	var req models.SuperadminRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.Username == "" || req.FullName == "" || req.Phone == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	superadminID, err := h.db.RegisterSuperadmin(req.Username, req.FullName, req.Phone, req.Password)
	if err != nil {
		if err.Error() == "username or phone already exists" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"superadmin_id": superadminID})
}

// createCategory creates a new category with optional thumbnail
// @Summary Create a new category
// @Description Creates a category with name and optional thumbnail image. Requires superadmin JWT authentication.
// @Tags Categories
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Category name"
// @Param name_ru formData string true "Category name_ru"
// @Param thumbnail formData file false "Thumbnail image"
// @Router /api/superadmin/categories [post]
func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.Role != "superadmin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := r.ParseMultipartForm(10 << 20) // Max 10 MB
	if err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing form")
		return
	}

	name := r.FormValue("name")
	name_ru := r.FormValue("name_ru")
	if name == "" {
		respondError(w, http.StatusBadRequest, "Category name is required")
		return
	}

	var thumbnailURL string
	file, handler, err := r.FormFile("thumbnail")
	if err == nil {
		defer file.Close()

		uploadDir := os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			uploadDir = "./Uploads"
		}
		categoriesDir := filepath.Join(uploadDir, "categories")
		if err := os.MkdirAll(categoriesDir, 0755); err != nil {
			respondError(w, http.StatusInternalServerError, "Error creating uploads directory")
			return
		}

		filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
		filePath := filepath.Join(categoriesDir, filename)
		out, err := os.Create(filePath)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Error saving file")
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Error copying file")
			return
		}

		thumbnailURL = filepath.Join("/uploads/categories", filename)
		thumbnailURL = strings.ReplaceAll(thumbnailURL, string(filepath.Separator), "/")
	} else if err != http.ErrMissingFile {
		respondError(w, http.StatusBadRequest, "Error retrieving file")
		return
	}

	categoryID, err := h.db.CreateCategory(name, name_ru, thumbnailURL)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			respondError(w, http.StatusConflict, "Category name already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"category_id": categoryID})
}

// deleteCategory deletes a category and its thumbnail
// @Summary Delete a category
// @Description Deletes a category and its thumbnail image. Requires superadmin JWT authentication.
// @Tags Categories
// @Produce json
// @Security BearerAuth
// @Param category_id path string true "Category ID"
// @Router /api/superadmin/categories/{category_id} [delete]
func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.Role != "superadmin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	categoryIDStr := vars["category_id"]
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	err = h.db.DeleteCategory(categoryID)
	if err != nil {
		if err.Error() == "category not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			respondError(w, http.StatusConflict, "Category has associated products")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Category and thumbnail deleted successfully"})
}

// getCategories retrieves paginated categories with search
// @Summary Get categories
// @Description Retrieves a paginated list of categories with optional name search. Requires superadmin JWT authentication.
// @Tags Categories
// @Produce json
// @Param page query integer false "Page number (default: 1)"
// @Param limit query integer false "Items per page (default: 10)"
// @Param search query string false "Search by category name"
// @Router /categories [get]
func (h *Handler) getCategories(w http.ResponseWriter, r *http.Request) {
	// claims, ok := r.Context().Value("claims").(*models.Claims)
	// if !ok || claims.Role != "superadmin" {
	// 	respondError(w, http.StatusUnauthorized, "Unauthorized")
	// 	return
	// }

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	categories, err := h.db.GetCategories(page, limit, search)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, categories)
}

// // CartRequest for adding a product to the cart
// type CartRequest struct {
//     MarketID    int `json:"market_id"`
//     ProductID   int `json:"product_id"`
//     ThumbnailID int `json:"thumbnail_id"`
//     SizeID      int `json:"size_id"`
//     Count       int `json:"count"`
// }

// CartRequest for adding products to the cart
type CartRequest struct {
	MarketID int                     `json:"market_id"`
	Products []models.CartProductReq `json:"products"`
}

// addToCart adds a product to the user's cart
// @Summary Add to cart
// @Description Adds or updates a product in the user's cart for a market under a single cart order. Requires user JWT authentication.
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cart body models.CartRequest true "Cart entry details"
// @Router /api/cart [post]
func (h *Handler) addToCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.ProductID <= 0 || req.ThumbnailID <= 0 || req.SizeID <= 0 || req.Count <= 0 {
		respondError(w, http.StatusBadRequest, "Invalid or missing product_id, thumbnail_id, size_id, or count")
		return
	}

	cartOrderID, err := h.db.AddToCart(claims.UserID, req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"cart_order_id": cartOrderID})
}

// getCart retrieves the user's cart
// @Summary Get user cart
// @Description Retrieves the user's cart grouped by cart order and markets. Requires user JWT authentication.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Router /api/cart [get]
func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cart, err := h.db.GetUserCart(claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, cart)
}

// deleteCart deletes a user's cart order
// @Summary Delete cart
// @Description Deletes all entries for a user's cart order by cart_order_id. Requires user JWT authentication.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Param cart_order_id path string true "Cart Order ID"
// @Router /api/cart/{cart_order_id} [delete]
func (h *Handler) deleteCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	cartOrderIDStr := vars["cart_order_id"]
	cartOrderID, err := strconv.Atoi(cartOrderIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid cart order ID")
		return
	}

	err = h.db.DeleteCart(claims.UserID, cartOrderID)
	if err != nil {
		if err.Error() == "cart not found or not owned by user" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Cart deleted successfully"})
}

// clearCart clears all cart entries for the authenticated user
// @Summary Clear cart
// @Description Deletes all cart entries for the authenticated user. Requires user JWT authentication.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Router /api/cart [delete]
func (h *Handler) clearCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.db.ClearCart(claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Cart cleared successfully"})
}

// LocationRequest for adding a new location
type LocationRequest struct {
	LocationName       string `json:"location_name"`
	LocationNameRu     string `json:"location_name_ru"`
	LocationAddress    string `json:"location_address"`
	LocationAddressRu string `json:"location_address_ru"`
}

// OrderRequest for submitting an order
type OrderRequest struct {
	LocationID int    `json:"location_id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Notes      string `json:"notes"`
}

// addLocation adds a new location for the user
// @Summary Add location
// @Description Adds a new location with name and address for the authenticated user.
// @Tags Locations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param location body LocationRequest true "Location details"
// @Router /api/locations [post]
func (h *Handler) addLocation(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req LocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.LocationName == "" || req.LocationAddress == "" {
		respondError(w, http.StatusBadRequest, "Location name and address are required")
		return
	}

	locationID, err := h.db.CreateLocation(claims.UserID, req.LocationName, req.LocationNameRu, req.LocationAddress, req.LocationAddressRu)
	if err != nil {
		if err.Error() == "location name already exists for user" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"location_id": locationID})
}

// getLocations retrieves the user's locations
// @Summary Get user locations
// @Description Retrieves all locations for the authenticated user.
// @Tags Locations
// @Produce json
// @Security BearerAuth
// @Router /api/locations [get]
func (h *Handler) getLocations(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	locations, err := h.db.GetUserLocations(claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, locations)
}

// createOrder submits an order for a cart
// @Summary Create order
// @Description Submits an order for a cart order with location and user details. Requires user JWT authentication.
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cart_order_id path string true "Cart Order ID"
// @Param order body OrderRequest true "Order details"
// @Router /api/cart/{cart_order_id}/order [post]
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	cartOrderIDStr := vars["cart_order_id"]
	cartOrderID, err := strconv.Atoi(cartOrderIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid cart order ID")
		return
	}

	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.LocationID <= 0 || req.Name == "" || req.Phone == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	orderID, err := h.db.CreateOrder(claims.UserID, cartOrderID, req.LocationID, req.Name, req.Phone, req.Notes)
	if err != nil {
		if err.Error() == "cart not found or not owned by user" || err.Error() == "location not found or not owned by user" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "order already exists for this cart" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"order_id": orderID})
}

// getMarketAdminOrders retrieves orders for a market admin's market
// @Summary Get market admin orders
// @Description Retrieves all orders for the market admin's market, optionally filtered by status. Requires market admin JWT authentication.
// @Tags Market Admin
// @Produce json
// @Security BearerAuth
// @Param status query string false "Order status (pending, processing, delivered, cancelled)"
// @Router /api/market/orders [get]
func (h *Handler) getMarketAdminOrders(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	status := r.URL.Query().Get("status")
	if status != "" {
		validStatuses := map[string]bool{
			"pending":    true,
			"processing": true,
			"delivered":  true,
			"cancelled":  true,
		}
		if !validStatuses[status] {
			respondError(w, http.StatusBadRequest, "Invalid status. Must be one of: pending, processing, delivered, cancelled")
			return
		}
	}

	orders, err := h.db.GetMarketAdminOrders(claims.MarketID, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, orders)
}

// getMarketAdminOrderByID retrieves a specific order by cart_order_id for a market admin
// @Summary Get market admin order by ID
// @Description Retrieves detailed order information for a specific cart_order_id. Requires market admin JWT authentication.
// @Tags Market Admin
// @Produce json
// @Security BearerAuth
// @Param cart_order_id path string true "Cart Order ID"
// @Router /api/market/orders/{cart_order_id} [get]
func (h *Handler) getMarketAdminOrderByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	cartOrderIDStr := vars["cart_order_id"]
	cartOrderID, err := strconv.Atoi(cartOrderIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid cart order ID")
		return
	}

	order, err := h.db.GetMarketAdminOrderByID(claims.MarketID, cartOrderID)
	if err != nil {
		if err.Error() == "order not found or not for this market" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, order)
}

// deleteLocation deletes a user's location
// @Summary Delete location
// @Description Deletes a specific location for the authenticated user if no orders reference it. Requires user JWT authentication.
// @Tags Locations
// @Produce json
// @Security BearerAuth
// @Param location_id path string true "Location ID"
// @Router /api/locations/{location_id} [delete]
func (h *Handler) deleteLocation(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	locationIDStr := vars["location_id"]
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid location ID")
		return
	}

	err = h.db.DeleteLocation(claims.UserID, locationID)
	if err != nil {
		if err.Error() == "location not found or not owned by user" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "location is referenced by orders and cannot be deleted" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Location deleted successfully"})
}

// deleteCartBySizeID deletes a cart entry for the authenticated user based on size_id
// @Summary Delete cart entry by size_id
// @Description Deletes a specific cart entry for the authenticated user based on size_id. Requires user JWT authentication.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Param size_id path int true "Size ID"
// @Router /api/cart-product/{size_id} [delete]
func (h *Handler) deleteCartBySizeID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sizeIDStr := vars["size_id"]
	sizeID, err := strconv.Atoi(sizeIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid size_id")
		return
	}

	err = h.db.DeleteCartBySizeID(claims.UserID, sizeID)
	if err != nil {
		if err.Error() == "cart entry not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Cart entry deleted successfully"})
}

// updateCartCountBySizeID updates the count of a cart entry for the authenticated user based on size_id
// @Summary Update cart entry count
// @Description Updates the count of a specific cart entry for the authenticated user based on size_id. Requires user JWT authentication.
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param size_id path int true "Size ID"
// @Param body body models.UpdateCartRequest true "Count change"
// @Router /api/cart/{size_id} [put]
func (h *Handler) updateCartCountBySizeID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sizeIDStr := vars["size_id"]
	sizeID, err := strconv.Atoi(sizeIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid size_id")
		return
	}

	var req models.UpdateCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.CountChange == 0 {
		respondError(w, http.StatusBadRequest, "Count change cannot be zero")
		return
	}

	newCount, err := h.db.UpdateCartCountBySizeID(claims.UserID, sizeID, req.CountChange)
	if err != nil {
		if err.Error() == "cart entry not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "count cannot be less than 1" {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Cart entry count updated successfully",
		"new_count": newCount,
	})
}

// updateLocationByID updates a location entry for the authenticated user based on location_id
// @Summary Update location
// @Description Updates the name and/or address of a specific location for the authenticated user. Requires user JWT authentication.
// @Tags Locations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param location_id path int true "Location ID"
// @Param body body models.UpdateLocationRequest true "Location update details"
// @Router /api/locations/{location_id} [put]
func (h *Handler) updateLocationByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	locationIDStr := vars["location_id"]
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid location_id")
		return
	}

	var req models.UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.LocationName == "" && req.LocationAddress == "" {
		respondError(w, http.StatusBadRequest, "No fields provided to update")
		return
	}

	updatedLocation, err := h.db.UpdateLocationByID(claims.UserID, locationID, req)
	if err != nil {
		if err.Error() == "location not found or unauthorized" || err.Error() == "location not found after update" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "no fields provided to update" {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Location updated successfully",
		"location": updatedLocation,
	})
}


// getProfile retrieves the authenticated user's profile
// @Summary Get user profile
// @Description Retrieves the full_name, phone, and id of the authenticated user. Requires user JWT authentication.
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Router /api/profile [get]
func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	profile, err := h.db.GetUserProfile(claims.UserID)
	if err != nil {
		if err.Error() == "user not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, profile)
}


// updateProfile updates the authenticated user's profile
// @Summary Update user profile
// @Description Updates the full_name and/or phone of the authenticated user. Requires user JWT authentication.
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.UpdateProfileRequest true "Profile update details"
// @Router /api/profile [put]
func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.FullName == "" && req.Phone == "" {
		respondError(w, http.StatusBadRequest, "No fields provided to update")
		return
	}

	updatedProfile, err := h.db.UpdateUserProfile(claims.UserID, req)
	if err != nil {
		if err.Error() == "user not found" || err.Error() == "user not found after update" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "no fields provided to update" {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Profile updated successfully",
		"profile": updatedProfile,
	})
}
  