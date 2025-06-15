package api

import (
	"Dowlet_projects/ecommerce/models"
	"log"
	"net/http"

	//"Dowlet_projects/ecommerce/services"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/gorilla/mux"
)


// updateProduct updates a product
// @Summary Update a product
// @Description Updates a product for the market admin's market, including its thumbnail image. Requires JWT authentication.
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param product_id path string true "Product ID"
// @Param category_id formData integer true "Category ID"
// @Param name formData string true "Product name"
// @Param name_ru formData string false "Product name (Russian)"
// @Param price formData number true "Product price"
// @Param discount formData number false "Discount percentage"
// @Param description formData string false "Product description"
// @Param description_ru formData string false "Product description (Russian)"
// @Param is_active formData boolean false "Product active status"
// @Param thumbnail formData file false "Thumbnail image"
// @Router /api/market/products/{product_id} [put]
func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
    // Verify market admin
    claims, ok := r.Context().Value("claims").(*models.Claims)
    if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
        respondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    // Get product ID from URL
    vars := mux.Vars(r)
    productIDStr, ok := vars["product_id"]
    if !ok {
		fmt.Println(ok)
        respondError(w, http.StatusBadRequest, "Missing product ID")
        return
    }
    productID, err := strconv.Atoi(productIDStr)
    if err != nil || productID < 1 {
		fmt.Println(err)
        respondError(w, http.StatusBadRequest, "Invalid product ID")
        return
    }

    // Parse multipart form (max 10MB)
    if err := r.ParseMultipartForm(10 << 20); err != nil {
		fmt.Println(err)
        respondError(w, http.StatusBadRequest, "Error parsing form")
        return
    }

    // Parse form fields
    categoryIDStr := r.FormValue("category_id")
    name := r.FormValue("name")
    nameRu := r.FormValue("name_ru")
    priceStr := r.FormValue("price")
    discountStr := r.FormValue("discount")
    description := r.FormValue("description")
    descriptionRu := r.FormValue("description_ru")
    isActiveStr := r.FormValue("is_active")

    // Validate required fields
    if name == "" || priceStr == "" {
		fmt.Println("name priceStr")
        respondError(w, http.StatusBadRequest, "Missing required fields: name and price")
        return
    }

    // Parse numeric fields
    categoryID, err := strconv.Atoi(categoryIDStr)
    if err != nil || categoryID < 1 {
		fmt.Println(err)
        respondError(w, http.StatusBadRequest, "Invalid category ID")
        return
    }
    price, err := strconv.ParseFloat(priceStr, 64)
    if err != nil || price <= 0 {
		fmt.Println(err)
        respondError(w, http.StatusBadRequest, "Invalid price")
        return
    }
    var discount float64
    if discountStr != "" {
        discount, err = strconv.ParseFloat(discountStr, 64)
        if err != nil || discount < 0 {
			fmt.Println(err)
            respondError(w, http.StatusBadRequest, "Invalid discount")
            return
        }
    }
    isActive := false
    if isActiveStr != "" {
        isActive, err = strconv.ParseBool(isActiveStr)
        if err != nil {
			fmt.Println(err)
            respondError(w, http.StatusBadRequest, "Invalid is_active value")
            return
        }
    }

    // Get upload directory
    uploadDir := os.Getenv("UPLOAD_DIR")
    if uploadDir == "" {
        uploadDir = "./Uploads/products"
    }

    // Ensure upload directory exists
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Println(err)
        log.Printf("Failed to create upload directory %s: %v", uploadDir, err)
        respondError(w, http.StatusInternalServerError, "Failed to create upload directory")
        return
    }

    // Handle file upload
    var imageURL string
    file, handler, err := r.FormFile("thumbnail")
    if err == nil {
		fmt.Println(err)
        defer file.Close()
        // Generate unique file name
        //ext := filepath.Ext(handler.Filename)
		imageURL = fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
        //imageURL = fmt.Sprintf("%d-%d%s", productID, time.Now().UnixNano(), ext)
        filePath := filepath.Join(uploadDir, imageURL)

        // Save file
        f, err := os.Create(filePath)
        if err != nil {
            log.Printf("Failed to create file %s: %v", filePath, err)
            respondError(w, http.StatusInternalServerError, "Failed to save image")
            return
        }
        defer f.Close()
        if _, err := io.Copy(f, file); err != nil {
            log.Printf("Failed to write file %s: %v", filePath, err)
            respondError(w, http.StatusInternalServerError, "Failed to save image")
            return
        }
        log.Printf("Saved new image: %s", filePath)
    } else if err != http.ErrMissingFile {
        respondError(w, http.StatusBadRequest, "Error processing image")
        return
    }

    // Update product and get old image URL
    oldImageURL, err := h.db.UpdateProduct(r.Context(), claims.MarketID, productID, categoryID, name, nameRu, price, discount, description, descriptionRu, isActive, imageURL)
    if err != nil {
        if err.Error() == "product not found or unauthorized" || err.Error() == "thumbnail not found" {
            fmt.Println(err)
			respondError(w, http.StatusNotFound, err.Error())
            return
        }
		fmt.Println(err)
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    // Delete old thumbnail file if it exists and a new image was uploaded
    if oldImageURL != "" && imageURL != "" {
        fileName := filepath.Base(oldImageURL)
        filePath := filepath.Join(uploadDir, fileName)
        if _, err := os.Stat(filePath); err == nil {
            if err := os.Remove(filePath); err != nil {
                log.Printf("Failed to delete old thumbnail at %s: %v", filePath, err)
                respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete old thumbnail: %v", err))
                return
            }
            log.Printf("Deleted old thumbnail: %s", filePath)
        } else if !os.IsNotExist(err) {
            log.Printf("Error checking old thumbnail at %s: %v", filePath, err)
            respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to check old thumbnail: %v", err))
            return
        }
    }

    respondJSON(w, http.StatusOK, map[string]string{"message": "Product updated successfully"})
}


// updateMarketProfile updates the authenticated market admin's market profile
// @Summary Update market profile
// @Description Updates the market profile data (except phone) for the authenticated market admin. Supports thumbnail file upload. Requires market admin JWT authentication.
// @Tags Market
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param delivery_price formData number false "Delivery price"
// @Param name formData string false "Market name"
// @Param name_ru formData string false "Market name in Russian"
// @Param location formData string false "Market location"
// @Param location_ru formData string false "Market location in Russian"
// @Param thumbnail formData file false "Thumbnail image (PNG/JPEG, max 5MB)"
// @Router /api/market/profile [put]
func (h *Handler) updateMarketProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get form fields
	req := models.UpdateMarketProfileRequest{
		DeliveryPrice: parseFloat(r.FormValue("delivery_price")),
		Name:          r.FormValue("name"),
		NameRu:        r.FormValue("name_ru"),
		Location:      r.FormValue("location"),
		LocationRu:    r.FormValue("location_ru"),
	}

	// Check if any text fields are provided
	hasTextFields := req.DeliveryPrice != 0 || req.Name != "" || req.NameRu != "" || req.Location != "" || req.LocationRu != ""

	// Handle thumbnail upload
	var thumbnailURL string
	file, fileHeader, err := r.FormFile("thumbnail")
	var filename string
	if err == nil {
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
		filename = fmt.Sprintf("%d-%s", time.Now().UnixNano(), strings.TrimSuffix(fileHeader.Filename, ext)+ext)
		thumbnailURL = filepath.Join("/uploads/markets", filename)
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
	} else if err != http.ErrMissingFile {
		respondError(w, http.StatusBadRequest, "Error processing thumbnail file")
		return
	}

	// Check if any fields are provided
	if !hasTextFields && thumbnailURL == "" {
		respondError(w, http.StatusBadRequest, "No fields provided to update")
		return
	}

	thumbnailURL1 := filepath.Join("/uploads/markets", filename)
	thumbnailURL1 = strings.ReplaceAll(thumbnailURL1, string(filepath.Separator), "/")

	updatedProfile, err := h.db.UpdateMarketProfile(claims.MarketID, req, thumbnailURL1)
	if err != nil {
		// Clean up uploaded file if database update fails
		if thumbnailURL != "" {
			os.Remove(filepath.Join(".", thumbnailURL))
		}
		if err.Error() == "market not found" || err.Error() == "market not found after update" {
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
		"message": "Market profile updated successfully",
		"profile": updatedProfile,
	})
}



// updateOrderStatus updates the status of an order
// @Summary Update order status
// @Description Updates the status of an order to 'canceled' or 'delivered'. Requires market admin JWT authentication.
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cart_order_id path integer true "Cart Order ID"
// @Param user_id path integer true "User ID"
// @Param body body models.UpdateOrderStatusRequest true "Order status update"
// @Router /api/market/orders/{cart_order_id}/{user_id} [put]
func (h *Handler) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		if !ok {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	// Get order_id from URL
	vars := mux.Vars(r)
	orderIDStr := vars["cart_order_id"]
	cartOrderID, err := strconv.Atoi(orderIDStr)
	if err != nil || cartOrderID < 1 {
		respondError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	userIDStr := vars["user_id"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID < 1 {
		respondError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	// Parse request body
	var req models.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	// Validate status
	if req.Status == "" {
		respondError(w, http.StatusBadRequest, "Status is required")
		return
	}

	// Update order status
	if err := h.db.UpdateOrderStatus(cartOrderID, userID, claims.MarketID, req.Status); err != nil {
		if err.Error() == "order not found" || err.Error() == "order not found or not associated with this market" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "invalid status; must be 'canceled' or 'delivered'" {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Order status updated successfully",
	})
}



// deleteUserHistory updates the is_active of orders' order
// @Summary Update order is_active
// @Description Updates the is_active of an order to false
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order_id path integer true "Order ID"
// @Router /api/user-orders/{order_id} [put]
func (h *Handler) deleteUserHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok  || claims.Role != "user" {
		if !ok {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	// Get order_id from URL
	vars := mux.Vars(r)
	orderIDStr := vars["order_id"]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil || orderID < 1 {
		respondError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}


	// Update order status
	if err := h.db.DeleteUserHistory(orderID, claims.UserID); err != nil {
		if err.Error() == "order not found" || err.Error() == "order not found or not associated with this market" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Order status updated successfully",
	})
}


// updateSize updates a size entry
// @Summary Update a size
// @Description Updates the count, price, and size of a size entry. Requires market admin JWT authentication.
// @Tags Markets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param size_id path integer true "Size ID"
// @Param body body models.UpdateSizeRequest true "Size update details"
// @Router /api/market/sizes/{size_id} [put]
func (h *Handler) updateSize(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		if !ok {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	// Get size_id from URL
	vars := mux.Vars(r)
	sizeIDStr := vars["size_id"]
	sizeID, err := strconv.Atoi(sizeIDStr)
	if err != nil || sizeID < 1 {
		respondError(w, http.StatusBadRequest, "Invalid size ID")
		return
	}

	// Parse request body
	var req models.UpdateSizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	// Validate input
	if req.Count < 0 {
		respondError(w, http.StatusBadRequest, "Count must be non-negative")
		return
	}
	if req.Price < 0 {
		respondError(w, http.StatusBadRequest, "Price must be non-negative")
		return
	}
	if req.Size == "" {
		respondError(w, http.StatusBadRequest, "Size is required")
		return
	}
	if len(req.Size) > 50 {
		respondError(w, http.StatusBadRequest, "Size exceeds 50 characters")
		return
	}

	// Update size
	updatedSize, err := h.db.UpdateSize(sizeID, req)
	if err != nil {
		if err.Error() == "size not found or not associated with this market" || err.Error() == "size not found after update" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Size updated successfully",
		"size":    updatedSize,
	})
}


// updateThumbnail updates a thumbnail
// @Summary Update a thumbnail
// @Description Updates a thumbnail’s color, color_ru, and image. Deletes the old image and uploads a new one. Requires market admin JWT authentication.
// @Tags Markets
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param thumbnail_id path integer true "Thumbnail ID"
// @Param color formData string true "Color for thumbnail"
// @Param color_ru formData string true "Color_ru for thumbnail"
// @Param thumbnail formData file true "New thumbnail image"
// @Router /api/market/thumbnails/{thumbnail_id} [put]
func (h *Handler) updateThumbnail(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		if !ok {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	// Get thumbnail_id from URL
	vars := mux.Vars(r)
	thumbnailIDStr := vars["thumbnail_id"]
	thumbnailID, err := strconv.Atoi(thumbnailIDStr)
	if err != nil || thumbnailID < 1 {
		respondError(w, http.StatusBadRequest, "Invalid thumbnail ID")
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // Max 10 MB
	if err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing form")
		return
	}

	// Get form values
	color := r.FormValue("color")
	colorRu := r.FormValue("color_ru")
	if color == "" {
		respondError(w, http.StatusBadRequest, "Color is required")
		return
	}
	if len(color) > 50 {
		respondError(w, http.StatusBadRequest, "Color exceeds 50 characters")
		return
	}
	if colorRu == "" {
		respondError(w, http.StatusBadRequest, "Color_ru is required")
		return
	}
	if len(colorRu) > 50 {
		respondError(w, http.StatusBadRequest, "Color_ru exceeds 50 characters")
		return
	}

	// Get thumbnail file
	files := r.MultipartForm.File["thumbnail"]
	if len(files) != 1 {
		respondError(w, http.StatusBadRequest, "Exactly one thumbnail file is required")
		return
	}

	// Fetch product_id to construct file path
	oldImageURL, filePath, imageCreated, err := h.db.UpdateThumbnail(thumbnailID, color, colorRu, claims.MarketID, files)

	if err != nil {
		if imageCreated != "" {
			os.Remove(filePath)
		}
		if err.Error() == "thumbnail not found or not associated with this market" || err.Error() == "thumbnail not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Delete old image file if it exists
	if oldImageURL != "" {
		oldFilePath := filepath.Join(".", oldImageURL)
		if err := os.Remove(oldFilePath); err != nil && !os.IsNotExist(err) {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete old thumbnail: %v", err))
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]int64{
		"thumbnail_id": int64(thumbnailID),
	})
}

// parseFloat parses a string to float64, returning 0 if empty or invalid
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
