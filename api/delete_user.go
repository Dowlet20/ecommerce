package api

import (
	"net/http"
	"Dowlet_projects/ecommerce/models"
	"strconv"
	"github.com/gorilla/mux"
)

// deleteCart deletes a user's cart order
// @Summary Delete cart
// @Description Deletes all entries for a user's cart order by cart_order_id. Requires user JWT authentication.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Param cart_order_id path string true "Cart Order ID"
// @Router /api/cart/{cart_order_id} [delete]
func (h *Handler) deleteCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	cartOrderIDStr := vars["cart_order_id"]
	cartOrderID, err := strconv.Atoi(cartOrderIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid cart order ID")
		return
	}

	err = h.db.DeleteCart(claims.UserID, cartOrderID)
	if err != nil {
		if err.Error() == "cart not found or not owned by user" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Cart deleted successfully"})
}


// deleteLocation deletes a user's location
// @Summary Delete location
// @Description Deletes a specific location for the authenticated user if no orders reference it. Requires user JWT authentication.
// @Tags Locations
// @Produce json
// @Security BearerAuth
// @Param location_id path string true "Location ID"
// @Router /api/locations/{location_id} [delete]
func (h *Handler) deleteLocation(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	locationIDStr := vars["location_id"]
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid location ID")
		return
	}

	err = h.db.DeleteLocation(claims.UserID, locationID)
	if err != nil {
		if err.Error() == "location not found or not owned by user" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "location is referenced by orders and cannot be deleted" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Location deleted successfully"})
}



// clearCart clears all cart entries for the authenticated user
// @Summary Clear cart
// @Description Deletes all cart entries for the authenticated user. Requires user JWT authentication.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Router /api/cart [delete]
func (h *Handler) clearCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.db.ClearCart(claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Cart cleared successfully"})
}



// deleteCartBySizeID deletes a cart entry for the authenticated user based on size_id
// @Summary Delete cart entry by size_id
// @Description Deletes a specific cart entry for the authenticated user based on size_id. Requires user JWT authentication.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Param size_id path int true "Size ID"
// @Router /api/cart-product/{size_id} [delete]
func (h *Handler) deleteCartBySizeID(w http.ResponseWriter, r *http.Request) {
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

	err = h.db.DeleteCartBySizeID(claims.UserID, sizeID)
	if err != nil {
		if err.Error() == "cart entry not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Cart entry deleted successfully"})
}