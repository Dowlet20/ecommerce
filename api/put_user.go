package api

import (
	"net/http"
	"Dowlet_projects/ecommerce/models"
	"strconv"
	"github.com/gorilla/mux"
	"encoding/json"
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