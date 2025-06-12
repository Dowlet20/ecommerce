package api

import (
	"Dowlet_projects/ecommerce/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	//"Dowlet_projects/ecommerce/services"
)

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

// deleteBanner deletes a banner and its thumbnail
// @Summary Delete a banner
// @Description Deletes a banner by ID and removes its thumbnail image. Requires superadmin JWT authentication.
// @Tags Banners
// @Produce json
// @Security BearerAuth
// @Param id path integer true "Banner ID"
// @Router /api/superadmin/banners/{id} [delete]
func (h *Handler) deleteBanner(w http.ResponseWriter, r *http.Request) {
    // Verify superadmin
    claims, ok := r.Context().Value("claims").(*models.Claims)
    if !ok || claims.Role != "superadmin" {
        if !ok {
            respondError(w, http.StatusUnauthorized, "Unauthorized")
            return
        }
        respondError(w, http.StatusForbidden, "Forbidden")
        return
    }

    // Get banner ID from URL
    vars := mux.Vars(r)
    bannerIDStr, ok := vars["id"]
    if !ok {
        respondError(w, http.StatusBadRequest, "Missing banner ID")
        return
    }
    bannerID, err := strconv.Atoi(bannerIDStr)
    if err != nil || bannerID < 1 {
        respondError(w, http.StatusBadRequest, "Invalid banner ID")
        return
    }

    // Delete banner and get thumbnail URL
    thumbnailURL, err := h.db.DeleteBanner(r.Context(), bannerID)
    if err != nil {
        if err.Error() == "banner not found" {
            respondError(w, http.StatusNotFound, err.Error())
            return
        }
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    // Get upload directory
    uploadDir := os.Getenv("UPLOAD_DIR")
    if uploadDir == "" {
        uploadDir = "./Uploads/banners"
    }

    // Ensure upload directory exists
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        log.Printf("Upload directory does not exist: %s", uploadDir)
        respondJSON(w, http.StatusOK, map[string]string{
            "message": "Banner deleted successfully, but upload directory not found",
        })
        return
    }

    // Delete thumbnail file if it exists
    if thumbnailURL != "" {
        // Extract base file name to prevent path traversal
        fileName := filepath.Base(thumbnailURL)
        filePath := filepath.Join(uploadDir, fileName)

        // Check if file exists
        if _, err := os.Stat(filePath); err == nil {
            if err := os.Remove(filePath); err != nil {
                log.Printf("Failed to delete thumbnail at %s: %v", filePath, err)
                respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete thumbnail: %v", err))
                return
            }
            log.Printf("Deleted thumbnail: %s", filePath)
        } else if !os.IsNotExist(err) {
            log.Printf("Error checking thumbnail at %s: %v", filePath, err)
            respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to check thumbnail: %v", err))
            return
        } else {
            log.Printf("Thumbnail does not exist: %s", filePath)
        }
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "Banner deleted successfully",
    })
}


// deleteUserMessage deletes a user message
// @Summary Delete a user message
// @Description Deletes a specific user message by ID. Requires superadmin JWT authentication.
// @Tags User Messages
// @Produce json
// @Security BearerAuth
// @Param id path integer true "Message ID"
// @Router /api/superadmin/user-messages/{id} [delete]
func (h *Handler) deleteUserMessage(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.Role != "superadmin" {
		if !ok {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	// Get message ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	// Delete message
	if err := h.db.DeleteUserMessage(id); err != nil {
		if err.Error() == "message not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "User message deleted successfully",
	})
}



// deleteMarketMessage deletes a market message
// @Summary Delete a market message
// @Description Deletes a specific market message by ID. Requires superadmin JWT authentication.
// @Tags Market Messages
// @Produce json
// @Security BearerAuth
// @Param id path integer true "Message ID"
// @Router /api/superadmin/market-messages/{id} [delete]
func (h *Handler) deleteMarketMessage(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.Role != "superadmin" {
		if !ok {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	// Get message ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	// Delete message
	if err := h.db.DeleteMarketMessage(id); err != nil {
		if err.Error() == "message not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Market message deleted successfully",
	})
}

// deleteUser deletes a user by ID for superadmin
// @Summary     Delete a user
// @Description Deletes a user by ID. Requires superadmin authentication.
// @Tags        Superadmin
// @Param       id path integer true "User ID"
// @Security    BearerAuth
// @Router      /api/superadmin/users/{id} [delete]
func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
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

    // Delete user
    err = h.db.DeleteUser(r.Context(), userID)
    if err != nil {
        if err.Error() == "user not found" {
            respondError(w, http.StatusNotFound, "User not found")
            return
        }
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    // Return 204 No Content
    w.WriteHeader(http.StatusNoContent)
}