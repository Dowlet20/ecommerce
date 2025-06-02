package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"io"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/yourusername/ecommerce/models"
	"github.com/yourusername/ecommerce/services"
	"log"
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
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// Public routes
	router.HandleFunc("/register", h.register).Methods("POST")
	router.HandleFunc("/verify-otp", h.verifyOTP).Methods("POST")
	router.HandleFunc("/login", h.login).Methods("POST")

	// Protected routes
	router.HandleFunc("/markets", h.authMiddleware(h.getMarkets)).Methods("GET")
	router.HandleFunc("/markets/{id}/products", h.authMiddleware(h.getMarketProducts)).Methods("GET")
	router.HandleFunc("/products", h.authMiddleware(h.getAllProducts)).Methods("GET")
	router.HandleFunc("/products/{id}", h.authMiddleware(h.getProduct)).Methods("GET")
	router.HandleFunc("/products", h.authMiddleware(h.createProduct)).Methods("POST")
	router.HandleFunc("/markets/{id}/thumbnail", h.authMiddleware(h.uploadMarketThumbnail)).Methods("POST")
	router.HandleFunc("/products/{id}", h.authMiddleware(h.deleteProduct)).Methods("DELETE")

	// Wrap router with CORS
	router.Use(c.Handler)
}

// register handles user registration
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	otp := services.GenerateOTP()
	// In production, send OTP via SMS (e.g., using Twilio)
	// Placeholder: fmt.Println("Sending OTP", otp, "to", user.Phone)

	err := h.db.SaveUserAndOTP(user, otp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered. Please verify OTP."})
}

// verifyOTP verifies the OTP for phone verification
func (h *Handler) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		OTP   string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.db.VerifyOTP(req.Phone, req.OTP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Phone verified successfully"})
}

// login handles user authentication
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.db.AuthenticateUser(req.Phone, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials or unverified phone", http.StatusUnauthorized)
		return
	}

	claims := &jwt.MapClaims{
		"full_name": user.FullName,
		"phone":     user.Phone,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// authMiddleware validates JWT token
func (h *Handler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("your-secret-key"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// getMarkets returns all markets
func (h *Handler) getMarkets(w http.ResponseWriter, r *http.Request) {
	markets, err := h.db.GetMarkets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(markets)
}

// getMarketProducts returns products for a specific market
func (h *Handler) getMarketProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	products, err := h.db.GetMarketProducts(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// getAllProducts returns all products across all markets
func (h *Handler) getAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.db.GetAllProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// getProduct returns a single product by ID
func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	product, err := h.db.GetProduct(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// createProduct creates a new product with thumbnails and sizes
func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	marketID := r.FormValue("market_id")
	name := r.FormValue("name")
	price := r.FormValue("price")
	discount := r.FormValue("discount")
	description := r.FormValue("description")
	colors := r.FormValue("colors")
	sizes := r.FormValue("sizes")

	if marketID == "" || name == "" || price == "" || colors == "" || sizes == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("uploads", 0755); err != nil {
		http.Error(w, "Error creating uploads directory", http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["thumbnails"]
	var thumbnailURLs []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), fileHeader.Filename)
		filePath := filepath.Join("uploads", filename)
		out, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Error copying file", http.StatusInternalServerError)
			return
		}

		thumbnailURLs = append(thumbnailURLs, "/uploads/"+filename)
	}

	productID, err := h.db.CreateProduct(marketID, name, price, discount, description, thumbnailURLs, colors, sizes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"product_id": productID})
}

// uploadMarketThumbnail uploads a thumbnail for a market
func (h *Handler) uploadMarketThumbnail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("thumbnail")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := os.MkdirAll("uploads", 0755); err != nil {
		http.Error(w, "Error creating uploads directory", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
	filePath := filepath.Join("uploads", filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	thumbnailURL := "/uploads/" + filename
	err = h.db.UpdateMarketThumbnail(id, thumbnailURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Thumbnail uploaded successfully"})
}

// deleteProduct deletes a product and its associated files
func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get thumbnail URLs to delete files
	thumbnails, err := h.db.GetProductThumbnails(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete product from database
	err = h.db.DeleteProduct(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete thumbnail files
	for _, thumb := range thumbnails {
		filePath := filepath.Join(".", thumb.ImageURL[1:]) // Remove leading '/'
		if err := os.Remove(filePath); err != nil {
			log.Printf("Error deleting file %s: %v", filePath, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Product deleted successfully"})
}