package api

import (
   "net/http"
   "Dowlet_projects/ecommerce/models"
	//"Dowlet_projects/ecommerce/services"
	"fmt"
	"strconv"
	"io"
	"os"
	"path/filepath"
	"time"
	"strings"
)



// createMarket creates a new market with admin account and optional thumbnail
// @Summary Create a new market
// @Description Creates a new market with an admin account and optional thumbnail image. Requires superadmin JWT authentication.
// @Tags Markets
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Market name"
// @Param name_ru formData string true "Market name russian"
// @Param location formData string true "Market location"
// @Param location_ru formData string true "Market location russian"
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
	location := r.FormValue("location")
	location_ru := r.FormValue("location_ru")
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

	createdUsername, marketID, err := h.db.CreateMarket(name, name_ru, location, location_ru, thumbnailURL, phone, password, deliveryPrice)
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


// createBanner creates a new banner with an uploaded thumbnail
// @Summary Create a new banner
// @Description Creates a new banner with a description and thumbnail image. Requires superadmin JWT authentication.
// @Tags Banners
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param description formData string false "Banner description (max 255 characters)"
// @Param thumbnail formData file true "Thumbnail image (PNG/JPEG, max 5MB)"
// @Router /api/superadmin/banners [post]
func (h *Handler) createBanner(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.Role != "superadmin" {
		if !ok {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get description
	req := models.CreateBannerRequest{
		Description: r.FormValue("description"),
	}

	// Validate description length
	if len(req.Description) > 255 {
		respondError(w, http.StatusBadRequest, "Description exceeds 255 characters")
		return
	}

	// Handle thumbnail upload
	file, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Thumbnail file is required")
		return
	}
	defer file.Close()

	// Validate file type (PNG/JPEG)
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType != "image/png" && mimeType != "image/jpeg" {
		respondError(w, http.StatusBadRequest, "Invalid file type; only PNG or JPEG allowed")
		return
	}

	// Validate file size (max 5MB)
	if fileHeader.Size > 5<<20 {
		respondError(w, http.StatusBadRequest, "File size exceeds 5MB limit")
		return
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), strings.TrimSuffix(fileHeader.Filename, ext)+ext)
	thumbnailURL := filepath.Join("/uploads/banners", filename)
	dstPath := filepath.Join(".", thumbnailURL)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create upload directory")
		return
	}

	// Save file
	dst, err := os.Create(dstPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save thumbnail")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save thumbnail")
		return
	}


	thumbnailURL1 := filepath.Join("/uploads/banners", filename)
	thumbnailURL1 = strings.ReplaceAll(thumbnailURL1, string(filepath.Separator), "/")
	// Create banner in database
	banner, err := h.db.CreateBanner(req, thumbnailURL1)
	if err != nil {
		// Clean up uploaded file if database insert fails
		os.Remove(dstPath)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Banner created successfully",
		"banner":  banner,
	})
}