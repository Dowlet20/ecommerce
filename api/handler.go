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
	"strings"
	"time"
	"strconv"


	//"github.com/dgrijalva/jwt-go"
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
	// Protected routes under /api
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(h.authMiddleware)
	protected.HandleFunc("/favorites", h.getUserFavorites).Methods("GET", "OPTIONS")
	protected.HandleFunc("/favorites", h.toggleFavorite).Methods("POST", "OPTIONS")
	// Wrap router with CORS
	router.Use(c.Handler)
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


// getUserFavorites returns the authenticated user's favorite products with pagination
// @Summary Get user's favorite products
// @Description Retrieves a paginated list of the authenticated user's favorite products. Requires JWT authentication.
// @Tags Favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query integer false "Page number (default: 1)"
// @Param limit query integer false "Items per page (default: 10)"
// @Router /api/favorites [get]
func (h *Handler) getUserFavorites(w http.ResponseWriter, r *http.Request) {
	// Extract claims from JWT (assumed to be set in authMiddleware)
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
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

	// Fetch paginated favorite products
	products, err := h.db.GetUserFavoriteProducts(claims.UserID, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, products)
}



// FavoriteRequest represents the JSON request body for toggling a favorite
type FavoriteRequest struct {
	ProductID int `json:"product_id"`
}

// toggleFavorite adds or removes a product from the user's favorites
// @Summary Toggle favorite product
// @Description Adds a product to the user's favorites if not already favorited, or removes it if it is. Requires JWT authentication.
// @Tags Favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param favorite body FavoriteRequest true "Product ID"
// @Router /api/favorites [post]
func (h *Handler) toggleFavorite(w http.ResponseWriter, r *http.Request) {
	// Extract claims from JWT
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse JSON body
	var req FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.ProductID <= 0 {
		respondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Toggle favorite status
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
