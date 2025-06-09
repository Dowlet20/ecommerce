package api

import (
	"net/http"
	"Dowlet_projects/ecommerce/models"
	"strconv"
)

// getUserFavorites returns favorite products
// @Summary Get user's favorite products
// @Description Retrieves a paginated list of favorite products. Requires JWT authentication.
// @Tags Favorites
// @Produce json
// @Security BearerAuth
// @Param page query integer false "Page number (default: 1)"
// @Param limit query integer false "Items per page (default: 10)"
// @Router /api/favorites [get]
func (h *Handler) getUserFavorites(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

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

	products, err := h.db.GetUserFavoriteProducts(claims.UserID, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, products)
}


// getCart retrieves the user's cart
// @Summary Get user cart
// @Description Retrieves the user's cart grouped by cart order and markets. Requires user JWT authentication.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Router /api/cart [get]
func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cart, err := h.db.GetUserCart(claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, cart)
}



// getLocations retrieves the user's locations
// @Summary Get user locations
// @Description Retrieves all locations for the authenticated user.
// @Tags Locations
// @Produce json
// @Security BearerAuth
// @Router /api/locations [get]
func (h *Handler) getLocations(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	locations, err := h.db.GetUserLocations(claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, locations)
}



// getProfile retrieves the authenticated user's profile
// @Summary Get user profile
// @Description Retrieves the full_name, phone, and id of the authenticated user. Requires user JWT authentication.
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Router /api/profile [get]
func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	profile, err := h.db.GetUserProfile(claims.UserID)
	if err != nil {
		if err.Error() == "user not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, profile)
}



// getMarketAdminOrders retrieves orders for a market admin's market
// @Summary Get market admin orders
// @Description Retrieves all orders for the market admin's market, optionally filtered by status. Requires market admin JWT authentication.
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param status query string false "Order status (pending, delivered, cancelled)"
// @Router /api/user-orders [get]
func (h *Handler) getUserOrders(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	status := r.URL.Query().Get("status")
	if status != "" {
		validStatuses := map[string]bool{
			"pending":    true,
			"delivered":  true,
			"cancelled":  true,
		}
		if !validStatuses[status] {
			respondError(w, http.StatusBadRequest, "Invalid status. Must be one of: pending, delivered, cancelled")
			return
		}
	}

	orders, err := h.db.GetUserOrders(claims.UserID, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, orders)
}