package api


import (
	"encoding/json"
	"net/http"
	"fmt"
	"strconv"
	"io"
	"os"
	"path/filepath"
	"Dowlet_projects/ecommerce/models"
	"time"
	"strings"
	"Dowlet_projects/ecommerce/services"
	"github.com/gorilla/mux"
)


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
		filePath = filepath.Join("uploads", "products",  filename)
		urlPath = fmt.Sprintf("/uploads/products/%s", filename)

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
	productID, err := h.db.CreateProduct(r.Context(), claims.MarketID, req.CategoryID, req.Name, req.NameRu,
		req.Price, req.Discount, req.Description, req.DescriptionRu, req.IsActive, urlPath,
		filePath, filename)

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"product_id": productID})
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
	if err := os.MkdirAll("uploads/products", 0755); err != nil {
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
		filePath := filepath.Join("uploads/products", filename)
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
		ImageURL := filepath.Join("/uploads/products", filename)
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




// createMarketMessage creates a new market message for superadmin
// @Summary Create a new market message
// @Description Creates a new message from a market to superadmin. Requires market admin JWT authentication.
// @Tags Market Messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.CreateMarketMessageRequest true "Message details"
// @Router /api/market/messages [post]
func (h *Handler) createMarketMessage(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		if !ok {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	// Parse request body
	var req models.CreateMarketMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	// Validate input
	if req.FullName == "" {
		respondError(w, http.StatusBadRequest, "Full name is required")
		return
	}
	if len(req.FullName) > 50 {
		respondError(w, http.StatusBadRequest, "Full name exceeds 50 characters")
		return
	}
	if req.Phone == "" {
		respondError(w, http.StatusBadRequest, "Phone is required")
		return
	}
	if len(req.Phone) > 20 {
		respondError(w, http.StatusBadRequest, "Phone exceeds 20 characters")
		return
	}
	if req.Message == "" {
		respondError(w, http.StatusBadRequest, "Message is required")
		return
	}

	// Create message in database
	messageID, err := h.db.CreateMarketMessage(claims.MarketID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":   "Market message sent successfully",
		"message_id": messageID,
	})
}
