package api

import (
	"Dowlet_projects/ecommerce/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	//"Dowlet_projects/ecommerce/services"
)

// getMarketAdminOrders retrieves orders for a market admin's market
// @Summary Get market admin orders
// @Description Retrieves all orders for the market admin's market, optionally filtered by status. Requires market admin JWT authentication.
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param status query string false "Order status (pending, delivered, cancelled)"
// @Router /api/market/orders [get]
func (h *Handler) getMarketAdminOrders(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
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

	orders, err := h.db.GetMarketAdminOrders(claims.MarketID, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, orders)

}


// getMarketAdminOrderByID retrieves a specific order by cart_order_id for a market admin
// @Summary Get market admin order by ID
// @Description Retrieves detailed order information for a specific cart_order_id. Requires market admin JWT authentication.
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param cart_order_id path string true "Cart Order ID"
// @Router /api/market/orders/{cart_order_id} [get]
func (h *Handler) getMarketAdminOrderByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
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

	order, err := h.db.GetMarketAdminOrderByID(claims.MarketID, cartOrderID)
	if err != nil {
		if err.Error() == "order not found or not for this market" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, order)
}


// getMarketProfile retrieves the authenticated market admin's market profile
// @Summary Get market profile
// @Description Retrieves the market profile data for the authenticated market admin. Requires market admin JWT authentication.
// @Tags Market
// @Produce json
// @Security BearerAuth
// @Router /api/market/profile [get]
func (h *Handler) getMarketProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	profile, err := h.db.GetMarketProfile(claims.MarketID)
	if err != nil {
		if err.Error() == "market not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, profile)
}



// getMarketByID retrieves a market and its products
// @Summary Get market by ID
// @Description Retrieves market details and its products by market ID with pagination.
// @Tags Markets
// @Produce json
// @Security BearerAuth
// @Param limit query integer false "Number of products per page (default: 20, min: 1, max: 100)"
// @Param page query integer false "Page number (default: 1, min: 1)"
// @Router /api/market/markets [get]
func (h *Handler) getMarketByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		if !ok {
			fmt.Println(claims)
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")

	// Default values
	limit := 20
	page := 1

	// Parse limit
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			respondError(w, http.StatusBadRequest, "Invalid limit; must be an integer between 1 and 100")
			return
		}
	}

	// Parse page
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			respondError(w, http.StatusBadRequest, "Invalid page; must be an integer >= 1")
			return
		}
	}

	marketID := claims.MarketID

	market, products, totalCount, err := h.db.GetMarketByID(r.Context(), marketID, page, limit)
	if err != nil {
		if err.Error() == "market not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"market": market,
		"products": map[string]interface{}{
			"items":       products,
			"total_count": totalCount,
			"page":        page,
			"limit":       limit,
		},
	})
}