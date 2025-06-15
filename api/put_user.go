package api

import (
	"Dowlet_projects/ecommerce/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

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


// updateUserVerified updates a user's verified status for superadmin
// @Summary     Update user verified status
// @Description Updates the verified status of a user by ID. Requires superadmin authentication.
// @Tags        Superadmin
// @Accept      json
// @Produce     json
// @Param       id   path integer true "User ID"
// @Param       body body models.UpdateUserVerifiedRequest true "Verified status"
// @Security    BearerAuth
// @Router      /api/superadmin/users/{id} [put]
func (h *Handler) updateUserVerified(w http.ResponseWriter, r *http.Request) {
    // Verify superadmin
    claims, ok := r.Context().Value("claims").(*models.Claims)
    if !ok || claims.Role != "superadmin" {
        respondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    // Extract user ID from URL
    vars := mux.Vars(r)
    userIDStr, ok := vars["id"]
    if !ok {
        respondError(w, http.StatusBadRequest, "Missing user ID")
        return
    }

    userID, err := strconv.Atoi(userIDStr)
    if err != nil || userID < 1 {
        respondError(w, http.StatusBadRequest, "Invalid user ID")
        return
    }

    // Parse request body
    var req models.UpdateUserVerifiedRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Update verified status
    err = h.db.UpdateUserVerified(r.Context(), userID, req.Verified)
    if err != nil {
        if err.Error() == "user not found" {
            respondError(w, http.StatusNotFound, "User not found")
            return
        }
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    // Fetch updated user to return in response
    updatedUser, err := h.db.GetUsers(r.Context(), 1, 1, "") // Temporary: fetch user directly
    if err != nil || len(updatedUser) == 0 {
        respondError(w, http.StatusInternalServerError, "Failed to fetch updated user")
        return
    }

    respondJSON(w, http.StatusOK, updatedUser[0])
}


// updateMarket updates a market
// @Summary Update a market
// @Description Updates a market by ID, including its thumbnail image. Requires superadmin JWT authentication.
// @Tags Markets
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Market ID"
// @Param password formData string false "Market password"
// @Param delivery_price formData number false "Delivery price"
// @Param phone formData string false "Phone number"
// @Param name formData string false "Market name"
// @Param name_ru formData string false "Market name (Russian)"
// @Param location formData string false "Location"
// @Param location_ru formData string false "Location (Russian)"
// @Param isVIP formData boolean false "VIP status"
// @Param image formData file false "Thumbnail image"
// @Router /api/superadmin/markets/{id} [put]
func (h *Handler) updateMarket(w http.ResponseWriter, r *http.Request) {
    // Verify superadmin
    claims, ok := r.Context().Value("claims").(*models.Claims)
    if !ok || claims.Role != "superadmin" {
        if !ok {
            respondError(w, http.StatusUnauthorized, "Unauthorized")
        } else {
            respondError(w, http.StatusForbidden, "Forbidden")
        }
        return
    }

    // Get market ID from URL
    vars := mux.Vars(r)
    marketIDStr, ok := vars["id"]
    if !ok {
        respondError(w, http.StatusBadRequest, "Missing market ID")
        return
    }
    marketID, err := strconv.Atoi(marketIDStr)
    if err != nil || marketID < 1 {
        respondError(w, http.StatusBadRequest, "Invalid market ID")
        return
    }

    // Parse multipart form (max 10MB)
    if err := r.ParseMultipartForm(10 << 20); err != nil {
        respondError(w, http.StatusBadRequest, "Error parsing form")
        return
    }

    // Parse form fields
    password := r.FormValue("password")
    deliveryPriceStr := r.FormValue("delivery_price")
    phone := r.FormValue("phone")
    name := r.FormValue("name")
    nameRu := r.FormValue("name_ru")
    location := r.FormValue("location")
    locationRu := r.FormValue("location_ru")
    isVIPStr := r.FormValue("isVIP")

    // Validate and parse numeric/boolean fields
    var deliveryPrice *float64
    if deliveryPriceStr != "" {
        dp, err := strconv.ParseFloat(deliveryPriceStr, 64)
        if err != nil || dp < 0 {
            respondError(w, http.StatusBadRequest, "Invalid delivery price")
            return
        }
        deliveryPrice = &dp
    }
    var isVIP *bool
    if isVIPStr != "" {
        vip, err := strconv.ParseBool(isVIPStr)
        if err != nil {
            respondError(w, http.StatusBadRequest, "Invalid isVIP value")
            return
        }
        isVIP = &vip
    }

    // Validate phone if provided
    if phone != "" {
        if len(phone) > 20 || !strings.HasPrefix(phone, "+") {
            respondError(w, http.StatusBadRequest, "Invalid phone format")
            return
        }
    }

    // Get upload directory
    uploadDir := os.Getenv("UPLOAD_DIR")
    if uploadDir == "" {
        uploadDir = "./Uploads/markets"
    }

    // Ensure upload directory exists
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        log.Printf("Failed to create upload directory %s: %v", uploadDir, err)
        respondError(w, http.StatusInternalServerError, "Failed to create upload directory")
        return
    }

    // Handle file upload
    var imageURL string
    file, handler, err := r.FormFile("image")
    if err == nil {
        defer file.Close()
        // Validate file type
        ext := strings.ToLower(filepath.Ext(handler.Filename))
        if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
            respondError(w, http.StatusBadRequest, "Invalid image format; use JPG or PNG")
            return
        }
        // Generate unique file name
        imageURL = fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
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

    // Update market and get old thumbnail URL
    oldThumbnailURL, newThumbnailURL, err := h.db.UpdateMarket(r.Context(), marketID, password, deliveryPrice, phone, name, nameRu, location, locationRu, isVIP, imageURL)
    if err != nil {
        if err.Error() == "market not found" {
            respondError(w, http.StatusNotFound, err.Error())
            return
        }
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    // Delete old thumbnail file if a new image was uploaded
    if oldThumbnailURL != "" && imageURL != "" {
        fileName := filepath.Base(oldThumbnailURL)
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

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "Market updated successfully",
        "thumbnail_url":newThumbnailURL,
    })
}

// updateSuperadmin updates a superadmin
// @Summary Update a superadmin
// @Description Updates a superadmin's details. Requires superadmin JWT authentication.
// @Tags Superadmins
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param superadmin body models.SuperadminRequest true "Superadmin details"
// @Router /api/superadmin/superadmin-update [put]
func (h *Handler) updateSuperadmin(w http.ResponseWriter, r *http.Request) {
    // Verify superadmin
    claims, ok := r.Context().Value("claims").(*models.Claims)
    if !ok || claims.Role != "superadmin" {
        if !ok {
            respondError(w, http.StatusUnauthorized, "Unauthorized")
        } else {
            respondError(w, http.StatusForbidden, "Forbidden")
        }
        return
    }

    // Parse JSON body
    var req models.SuperadminRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Error parsing JSON body")
        return
    }

    // Validate inputs
    if req.Phone != nil {
        phone := strings.TrimSpace(*req.Phone)
        if len(phone) > 20 || (phone != "" && !strings.HasPrefix(phone, "+")) {
            respondError(w, http.StatusBadRequest, "Invalid phone format")
            return
        }
        req.Phone = &phone
    }
    if req.FullName != "" {
        fullName := strings.TrimSpace(req.FullName)
        if len(fullName) > 255 {
            respondError(w, http.StatusBadRequest, "Full name too long")
            return
        }
        req.FullName = fullName
    }
    if req.Username != "" {
        username := strings.TrimSpace(req.Username)
        if len(username) > 50 || len(username) < 3 {
            respondError(w, http.StatusBadRequest, "Username must be 3-50 characters")
            return
        }
        req.Username = username
    }
    if req.Password != "" && len(req.Password) < 6 {
        respondError(w, http.StatusBadRequest, "Password must be at least 6 characters")
        return
    }

    // Update superadmin
    err := h.db.UpdateSuperadmin(r.Context(), int(claims.UserID), req.Phone, req.FullName, req.Username, req.Password)
    if err != nil {
        switch err.Error() {
        case "superadmin not found":
            respondError(w, http.StatusNotFound, err.Error())
        case "username already exists":
            respondError(w, http.StatusConflict, err.Error())
        default:
            respondError(w, http.StatusInternalServerError, err.Error())
        }
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{"message": "Superadmin updated successfully"})
}
