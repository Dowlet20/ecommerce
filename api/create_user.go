package api

import (
	"net/http"
	"Dowlet_projects/ecommerce/models"
	"encoding/json"
	"strings"
	"strconv"
	"github.com/gorilla/mux"
)

// toggleFavorite toggles a favorite product
// @Summary Toggle favorite product
// @Description Adds or removes a product from favorites. Requires JWT authentication.
// @Tags Favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param favorite body FavoriteRequest true "Product ID"
// @Router /api/favorites [post]
func (h *Handler) toggleFavorite(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.ProductID <= 0 {
		respondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

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


// addToCart adds a product to the user's cart
// @Summary Add to cart
// @Description Adds or updates a product in the user's cart for a market under a single cart order. Requires user JWT authentication.
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cart body models.CartRequest true "Cart entry details"
// @Router /api/cart [post]
func (h *Handler) addToCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.CartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.ProductID <= 0 || req.ThumbnailID <= 0 || req.SizeID <= 0 || req.Count <= 0 {
		respondError(w, http.StatusBadRequest, "Invalid or missing product_id, thumbnail_id, size_id, or count")
		return
	}

	cartOrderID, err := h.db.AddToCart(claims.UserID, req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"cart_order_id": cartOrderID})
}



// addLocation adds a new location for the user
// @Summary Add location
// @Description Adds a new location with name and address for the authenticated user.
// @Tags Locations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param location body LocationRequest true "Location details"
// @Router /api/locations [post]
func (h *Handler) addLocation(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req LocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.LocationName == "" || req.LocationAddress == "" {
		respondError(w, http.StatusBadRequest, "Location name and address are required")
		return
	}

	locationID, err := h.db.CreateLocation(claims.UserID, req.LocationName, req.LocationNameRu, req.LocationAddress, req.LocationAddressRu)
	if err != nil {
		if err.Error() == "location name already exists for user" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"location_id": locationID})
}


// createOrder submits an order for a cart
// @Summary Create order
// @Description Submits an order for a cart order with location and user details. Requires user JWT authentication.
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cart_order_id path string true "Cart Order ID"
// @Param order body OrderRequest true "Order details"
// @Router /api/cart/{cart_order_id}/order [post]
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
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

	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.LocationID <= 0 || req.Name == "" || req.Phone == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	orderID, err := h.db.CreateOrder(claims.UserID, cartOrderID, req.LocationID, req.Name, req.Phone, req.Notes)
	if err != nil {
		if err.Error() == "cart not found or not owned by user" || err.Error() == "location not found or not owned by user" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "order already exists for this cart" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"order_id": orderID})
}


// createMessage creates a new user message for superadmin
// @Summary Create a new message
// @Description Creates a new message to superadmin. Requires user JWT authentication.
// @Tags User Messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.CreateMessageRequest true "Message details"
// @Router /api/messages [post]
func (h *Handler) createMessage(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.UserID == 0 || claims.Role != "user" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req models.CreateMessageRequest
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
	messageID, err := h.db.CreateMessage(claims.UserID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":   "Message sent successfully",
		"message_id": messageID,
	})
}