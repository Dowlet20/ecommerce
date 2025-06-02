package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"strconv"


	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"Dowlet_projects/ecommerce/models"
	"Dowlet_projects/ecommerce/services"
)

// Handler holds dependencies for API routes
type Handler struct {
	db *services.DBService
}

// NewHandler creates a new API handler
func NewHandler(db *services.DBService) *Handler {
	return &Handler{db: db}
}

// RegisterRequest defines the request body for user registration
type RegisterRequest struct {
	FullName string `json:"full_name" example:"John Doe" description:"Full name of the user" validate:"required"`
	Phone    string `json:"phone" example:"+12345678901" description:"Phone number in international format" validate:"required"`
}

// LoginRequest defines the request body for login
type LoginRequest struct {
	Phone string `json:"phone" example:"+12345678901" description:"Phone number in international format" validate:"required"`
}

// VerifyCodeRequest defines the request body for verification
type VerifyCodeRequest struct {
	Phone string `json:"phone" example:"+12345678901" description:"Phone number in international format" validate:"required"`
	Code  string `json:"code" example:"1234" description:"4-digit verification code" validate:"required,len=4"`
}

// SetupRoutes configures API routes
func (h *Handler) SetupRoutes(router *mux.Router) {
	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// Public routes
	router.HandleFunc("/register", h.register).Methods("POST", "OPTIONS")
	router.HandleFunc("/login", h.login).Methods("POST", "OPTIONS")
	router.HandleFunc("/verify", h.verifyCode).Methods("POST", "OPTIONS")
	router.HandleFunc("/markets", h.getMarkets).Methods("GET", "OPTIONS")
	router.HandleFunc("/markets/{id}/products", h.getMarketProducts).Methods("GET", "OPTIONS")
	router.HandleFunc("/products", h.getAllProducts).Methods("GET", "OPTIONS")
	router.HandleFunc("/products/{id}", h.getProduct).Methods("GET", "OPTIONS")
	router.HandleFunc("/products", h.createProduct).Methods("POST", "OPTIONS")
	router.HandleFunc("/markets/{id}/thumbnail", h.uploadMarketThumbnail).Methods("POST", "OPTIONS")
	router.HandleFunc("/products/{id}", h.deleteProduct).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/markets", h.createMarket).Methods("POST", "OPTIONS")
	router.HandleFunc("/markets/{id}", h.deleteMarket).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/products/{id}/thumbnails", h.addProductThumbnails).Methods("POST", "OPTIONS")
	router.HandleFunc("/thumbnails/{thumbnail_id}", h.deleteThumbnail).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/thumbnails", h.getAllThumbnails).Methods("GET", "OPTIONS")
	router.HandleFunc("/thumbnails/{thumbnail_id}/size", h.addSizeByThumbnail).Methods("POST", "OPTIONS")
	router.HandleFunc("/sizes/{size_id}", h.deleteSizeByID).Methods("DELETE", "OPTIONS")
	// Protected routes under /protected
	// protected := router.PathPrefix("/api").Subrouter()
	// router.Use(h.authMiddleware)

	// Wrap router with CORS
	router.Use(c.Handler)
}

// register handles user registration
// @Summary Register a new user
// @Description Registers a new user with a full name and phone number, sending a verification code. Validates input, checks for duplicate phone numbers, and logs the code (replace with SMS in production).
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "User registration details"
// @Router /register [post]
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.FullName == "" {
		respondError(w, http.StatusBadRequest, "Full name is required")
		return
	}
	if !validatePhone(req.Phone) {
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Check if phone exists
	_, err := h.db.GetUserByPhone(req.Phone)
	if err == nil {
		respondError(w, http.StatusConflict, "Phone number already registered")
		return
	}
	if err != sql.ErrNoRows {
		fmt.Println(err)
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Generate verification code
	code, err := services.GenerateVerificationCode(h.db, req.Phone, req.FullName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate verification code")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Verification code generated",
		"code":    code, // Remove in production; use SMS
	})
}

// login initiates user login
// @Summary Initiate user login
// @Description Initiates login by sending a verification code to the user's phone number. Checks if the phone is registered and logs the code.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body LoginRequest true "User login details"
// @Router /login [post]
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if !validatePhone(req.Phone) {
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Check if phone exists
	userID, err := h.db.GetUserByPhone(req.Phone)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Phone number not registered")
		return
	}
	if err != nil {
		fmt.Println(err)
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Generate verification code
	code, err := services.GenerateVerificationCode(h.db, req.Phone, "")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate verification code")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Verification code generated",
		"code":    code, // Remove in production; use SMS
		"user_id": userID,
	})
}

// verifyCode verifies the code and completes registration/login
// @Summary Verify phone number
// @Description Verifies a phone number using a 4-digit code, issuing a JWT token upon successful verification. Deletes the used code.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body VerifyCodeRequest true "Verification details"
// @Router /verify [post]
func (h *Handler) verifyCode(w http.ResponseWriter, r *http.Request) {
	var req VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if !validatePhone(req.Phone) {
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}
	if len(req.Code) != 4 || !regexp.MustCompile(`^\d{4}$`).MatchString(req.Code) {
		respondError(w, http.StatusBadRequest, "Invalid verification code")
		return
	}

	// Validate code
	storedCode, expiresAt, fullName, err := h.db.GetVerificationCode(req.Phone)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusBadRequest, "No verification code found")
		return
	}
	if err != nil {
		fmt.Println(err)
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if time.Now().After(expiresAt) {
		respondError(w, http.StatusBadRequest, "Verification code expired")
		return
	}

	if storedCode != req.Code {
		respondError(w, http.StatusBadRequest, "Invalid verification code")
		return
	}

	// Check if user is already registered
	userID, err := h.db.GetUserByPhone(req.Phone)
	var isRegistered bool
	if err == nil {
		isRegistered = true
	} else if err != sql.ErrNoRows {
		fmt.Println(err)
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// If not registered, complete registration
	if !isRegistered {
		if fullName == "" {
			respondError(w, http.StatusBadRequest, "Missing registration data")
			return
		}
		userID64, err := h.db.SaveUser(fullName, req.Phone)
		if err != nil {
			fmt.Println(err)
			respondError(w, http.StatusInternalServerError, "Failed to register user")
			return
		}
		userID = int(userID64)
	}

	// Delete used code
	if err := h.db.DeleteVerificationCode(req.Phone); err != nil {
		fmt.Printf("Failed to delete verification code: %v\n", err.Error())
	}

	// Generate JWT
	token, err := generateJWT(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"token":   token,
	})
}

// authMiddleware authenticates JWT tokens
func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			respondError(w, http.StatusUnauthorized, "Authorization header missing")
			return
		}
		// Remove "Bearer " prefix
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("your-secret-key"), nil
		})
		if err != nil || !token.Valid {
			respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// // authMiddleware authenticates JWT tokens
// func (h *Handler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		tokenStr := r.Header.Get("Authorization")
// 		if tokenStr == "" {
// 			respondError(w, http.StatusUnauthorized, "Authorization header missing")
// 			return
// 		}
// 		// Remove "Bearer " prefix
// 		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

// 		claims := &models.Claims{}
// 		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
// 			return []byte("your-secret-key"), nil
// 		})
// 		if err != nil || !token.Valid {
// 			respondError(w, http.StatusUnauthorized, "Invalid token")
// 			return
// 		}

// 		// Add claims to context
// 		ctx := context.WithValue(r.Context(), "claims", claims)
// 		next(w, r.WithContext(ctx))
// 	}
// }

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

// createMarket creates a new market with a name and optional thumbnail
// @Summary Create a new market
// @Description Creates a new market with a name and an optional thumbnail image. Requires JWT authentication.
// @Tags Markets
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Market name"
// @Param thumbnail formData file false "Thumbnail image"
// @Router /markets [post]
func (h *Handler) createMarket(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // Max 10 MB
	if err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing form")
		return
	}

	name := r.FormValue("name")
	if name == "" {
		respondError(w, http.StatusBadRequest, "Market name is required")
		return
	}

	var thumbnailURL string
	file, handler, err := r.FormFile("thumbnail")
	if err == nil {
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

		thumbnailURL = "/uploads/markets" + filename
	} else if err != http.ErrMissingFile {
		respondError(w, http.StatusBadRequest, "Error retrieving file")
		return
	}

	// Save market to database using DBService method
	marketID, err := h.db.CreateMarket(name, thumbnailURL)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create market")
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"market_id": int(marketID)})
}


// deleteMarket deletes a market and its associated thumbnail
// @Summary Delete a market
// @Description Deletes a market by ID, including its thumbnail file and associated products. Requires JWT authentication.
// @Tags Markets
// @Accept json
// @Produce json
// @Param id path string true "Market ID"
// @Router /markets/{id} [delete]
func (h *Handler) deleteMarket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Delete market using DBService
	err := h.db.DeleteMarket(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Market deleted successfully"})
}

// getMarketProducts retrieves products for a specific market
// @Summary Get products for a market
// @Description Retrieves all products for a specific market by ID. Requires JWT authentication.
// @Tags Markets
// @Accept json
// @Produce json
// @Param id path string true "Market ID"
// @Router /markets/{id}/products [get]
func (h *Handler) getMarketProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	products, err := h.db.GetMarketProducts(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, products)
}

// getAllProducts returns all products across all markets with pagination and search
// @Summary Get all products with pagination and search
// @Description Retrieves a paginated list of products across all markets, including favorite status, with optional name search. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Param page query integer false "Page number (default: 1)"
// @Param limit query integer false "Items per page (default: 10)"
// @Param search query string false "Search term for product name"
// @Router /products [get]
func (h *Handler) getAllProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
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

	// Fetch paginated and searched products
	products, err := h.db.GetPaginatedProducts(page, limit, search)
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
// ProductRequest represents the JSON request body for creating a product
type ProductRequest struct {
	MarketID    string  `json:"market_id"`
	Name        string  `json:"name"`
	Price       string  `json:"price"`
	Discount    string  `json:"discount,omitempty"`
	Description string  `json:"description,omitempty"`
}

// createProduct creates a new product
// @Summary Create a new product
// @Description Creates a new product without thumbnails or sizes using JSON body. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Param product body ProductRequest true "Product details"
// @Router /products [post]
func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.MarketID == "" || req.Name == "" || req.Price == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	productID, err := h.db.CreateProduct(req.MarketID, req.Name, req.Price, req.Discount, req.Description)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"product_id": productID})
}


// deleteSizeByID deletes a size by its ID
// @Summary Delete a size by ID
// @Description Deletes a size by its ID from the sizes table. Requires JWT authentication.
// @Tags Products
// @Produce json
// @Param size_id path string true "Size ID"
// @Router /sizes/{size_id} [delete]
func (h *Handler) deleteSizeByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sizeID := vars["size_id"]

	if sizeID == "" {
		respondError(w, http.StatusBadRequest, "Missing size ID")
		return
	}

	err := h.db.DeleteSizeByID(sizeID)
	if err != nil {
		if err.Error() == "size not found" {
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
// @Param id path string true "Market ID"
// @Param thumbnail formData file true "Thumbnail image"
// @Router /markets/{id}/thumbnail [post]
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


// addProductThumbnails adds multiple thumbnails to a product
// @Summary Add thumbnails to a product
// @Description Adds multiple thumbnail images with associated colors to a product by ID. Requires JWT authentication.
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Product ID"
// @Param colors formData string true "Comma-separated list of colors for thumbnails"
// @Param thumbnails formData file true "Thumbnail images"
// @Router /products/{id}/thumbnails [post]
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
			ImageURL:  "/uploads/products"+id + "/" + filename,
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

// deleteThumbnail deletes a thumbnail by its ID
// @Summary Delete a thumbnail
// @Description Deletes a thumbnail by its ID, including its file from uploads. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Param thumbnail_id path string true "Thumbnail ID"
// @Router /thumbnails/{thumbnail_id} [delete]
func (h *Handler) deleteThumbnail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	thumbnailID := vars["thumbnail_id"]

	if thumbnailID == "" {
		respondError(w, http.StatusBadRequest, "Missing thumbnail ID")
		return
	}

	// Delete thumbnail using DBService
	err := h.db.DeleteThumbnail(thumbnailID)
	if err != nil {
		if err.Error() == "thumbnail not found" {
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


// SizeRequest represents the JSON request body for adding a size
type SizeRequest struct {
	Size  string `json:"size"`
	Count int    `json:"count"`
}

// addSizeByThumbnail adds a single size with count linked to a thumbnail
// @Summary Add a size with count by thumbnail ID
// @Description Adds a single size with its count for a product associated with a specific thumbnail ID using JSON body. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Param thumbnail_id path string true "Thumbnail ID"
// @Param size body SizeRequest true "Size and count"
// @Router /thumbnails/{thumbnail_id}/size [post]
func (h *Handler) addSizeByThumbnail(w http.ResponseWriter, r *http.Request) {
	var req SizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	vars := mux.Vars(r)
	thumbnailID := vars["thumbnail_id"]

	if thumbnailID == "" || req.Size == "" || req.Count < 0 {
		respondError(w, http.StatusBadRequest, "Missing or invalid fields")
		return
	}

	err := h.db.CreateSizeByThumbnailID(thumbnailID, req.Size, req.Count)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Size added successfully"})
}
// deleteProduct deletes a product and its associated files
// @Summary Delete a product
// @Description Deletes a product by ID, including its thumbnails and associated files. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Router /products/{id} [delete]
func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get thumbnail URLs to delete files
	thumbnails, err := h.db.GetProductThumbnails(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Delete product from database
	err = h.db.DeleteProduct(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Delete thumbnail files
	for _, thumb := range thumbnails {
		filePath := filepath.Join(".", thumb.ImageURL[1:]) // Remove leading '/'
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Error deleting file %s: %v\n", filePath, err)
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Product deleted successfully"})
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

// validatePhone checks if the phone number is valid
func validatePhone(phone string) bool {
	re := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return re.MatchString(phone)
}

// generateJWT creates a JWT token
func generateJWT(userID int) (string, error) {
	claims := &models.Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}
	return signedToken, nil
}