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

	// User protected routes
	userProtected := router.PathPrefix("/api").Subrouter()
	userProtected.Use(h.authMiddleware)
	userProtected.HandleFunc("/favorites", h.getUserFavorites).Methods("GET", "OPTIONS")
	userProtected.HandleFunc("/favorites", h.toggleFavorite).Methods("POST", "OPTIONS")
	userProtected.HandleFunc("/cart", h.addToCart).Methods("POST", "OPTIONS")
	userProtected.HandleFunc("/cart", h.getCart).Methods("GET", "OPTIONS")

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
    CategoryID  int    `json:"category_id"`
    Name        string `json:"name"`
    Price       float64 `json:"price"`
    Discount    float64 `json:"discount,omitempty"`
    Description string `json:"description,omitempty"`
}

// SizeRequest for adding a size
type SizeRequest struct {
	Size  string `json:"size"`
	Count int    `json:"count"`
}

// FavoriteRequest for toggling favorite
type FavoriteRequest struct {
	ProductID int `json:"product_id"`
}

// MarketRequest for market login
type MarketRequest struct {
	Username string `json:"username"`
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
// @Param username formData string true "Admin username"
// @Param full_name formData string true "Admin full name"
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
	username := r.FormValue("username")
	fullName := r.FormValue("full_name")
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

	if name == "" || username == "" || fullName == "" || phone == "" || password == "" {
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
	} else if err != http.ErrMissingFile {
		respondError(w, http.StatusBadRequest, "Error retrieving file")
		return
	}

	createdUsername, marketID, err := h.db.CreateMarket(name, thumbnailURL, username, fullName, phone, password, deliveryPrice)
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
// @Description Retrieves paginated products with optional category and name search
// @Tags Products
// @Produce json
// @Param category_id query string false "Category ID"
// @Param page query integer false "Page number (default: 1)"
// @Param limit query integer false "Items per page (default: 10)"
// @Param search query string false "Search by product name"
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

	products, err := h.db.GetPaginatedProducts(categoryID, page, limit, search)
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


// registerUser registers a user
// @Summary Register user
// @Description Registers a user and sends OTP.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body UserRegisterRequest true "User details"
// @Router /users/register [post]
func (h *Handler) registerUser(w http.ResponseWriter, r *http.Request) {
	var req UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.FullName == "" || req.Phone == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	verifID, otp, err := h.db.RegisterUser(req.FullName, req.Phone)
	if err != nil {
		if err.Error() == "phone number already registered" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO: Send OTP via SMS
	fmt.Printf("OTP for %s: %s\n", req.Phone, otp)

	respondJSON(w, http.StatusOK, map[string]string{"message": "OTP sent", "verification_id": strconv.Itoa(verifID)})
}

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
// type ProductRequest struct {
// 	MarketID    string  `json:"market_id"`
// 	Name        string  `json:"name"`
// 	Price       string  `json:"price"`
// 	Discount    string  `json:"discount,omitempty"`
// 	Description string  `json:"description,omitempty"`
// }

// createProduct creates a new product
// @Summary Create a new product
// @Description Creates a new product for the market admin's market. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body ProductRequest true "Product details"
// @Router /api/market/products [post]
func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
    claims, ok := r.Context().Value("claims").(*models.Claims)
    if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
        respondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    var req ProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Error parsing JSON body")
        return
    }

    if req.Name == "" || req.Price == 0 || req.CategoryID == 0 {
        respondError(w, http.StatusBadRequest, "Missing required fields")
        return
    }

    productID, err := h.db.CreateProduct(claims.MarketID, req.CategoryID, req.Name, req.Price, req.Discount, req.Description)
    if err != nil {
        if err.Error() == "market not found" || err.Error() == "category not found" {
            respondError(w, http.StatusBadRequest, err.Error())
            return
        }
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
	if colors == "" {
		respondError(w, http.StatusBadRequest, "Colors are required")
		return
	}
	colorList := strings.Split(colors, ",")
	files := r.MultipartForm.File["thumbnails"]
	if len(files) == 0 {
		respondError(w, http.StatusBadRequest, "At least one thumbnail is required")
		return
	}
	if len(files) != len(colorList) {
		respondError(w, http.StatusBadRequest, "Number of thumbnails must match number of colors")
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

		thumbnails = append(thumbnails, services.ThumbnailData{
			ProductID: productID,
			Color:     colorList[i],
			ImageURL:  "/uploads/products/"+id + "/" + filename,
		})
	}

	// Save thumbnails to database
	err = h.db.CreateThumbnails(thumbnails)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Thumbnails added successfully"})
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

	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Missing credentials")
		return
	}

	userID, marketID, err := h.db.AuthenticateMarket(req.Username, req.Password)
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

	err := h.db.CreateSizeByThumbnailID(claims.MarketID, thumbnailID, req.Size, req.Count)
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
    } else if err != http.ErrMissingFile {
        respondError(w, http.StatusBadRequest, "Error retrieving file")
        return
    }

    categoryID, err := h.db.CreateCategory(name, thumbnailURL)
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


// CartRequest for adding a product to the cart
type CartRequest struct {
    MarketID    int `json:"market_id"`
    ProductID   int `json:"product_id"`
    ThumbnailID int `json:"thumbnail_id"`
    SizeID      int `json:"size_id"`
    Count       int `json:"count"`
}

// addToCart adds a product to the user's cart
// @Summary Add to cart
// @Description Adds or updates a product in the user's cart. Requires user JWT authentication.
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cart body CartRequest true "Cart entry details"
// @Router /api/cart [post]
func (h *Handler) addToCart(w http.ResponseWriter, r *http.Request) {
    claims, ok := r.Context().Value("claims").(*models.Claims)
    if !ok || claims.UserID == 0 || claims.Role != "user" {
        respondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    var req CartRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Error parsing JSON body")
        return
    }

    if req.MarketID <= 0 || req.ProductID <= 0 || req.ThumbnailID <= 0 || req.SizeID <= 0 || req.Count <= 0 {
        respondError(w, http.StatusBadRequest, "Invalid or missing fields")
        return
    }

    cartID, err := h.db.AddToCart(claims.UserID, req.MarketID, req.ProductID, req.ThumbnailID, req.SizeID, req.Count)
    if err != nil {
        if err.Error() == "invalid market, product, thumbnail, or size" {
            respondError(w, http.StatusBadRequest, err.Error())
            return
        }
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondJSON(w, http.StatusOK, map[string]int{"cart_id": cartID})
}


// getCart retrieves the user's cart
// @Summary Get user cart
// @Description Retrieves the user's cart grouped by markets. Requires user JWT authentication.
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
