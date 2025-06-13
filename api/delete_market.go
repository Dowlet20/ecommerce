package api

import (
	"net/http"
	"Dowlet_projects/ecommerce/models"
	"github.com/gorilla/mux"
	"strconv"
)

// deleteProduct deletes a product
// @Summary Delete a product
// @Description Deletes a product and its thumbnails for the market admin's market. Requires JWT authentication.
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param product_id path string true "Product ID"
// @Router /api/market/products/{product_id} [delete]
func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	productIDStr := vars["product_id"]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	err = h.db.DeleteProduct(r.Context(), claims.MarketID, productID)
	if err != nil {
		if err.Error() == "product not found or unauthorized" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Product and thumbnails deleted successfully"})
}


// deleteSizeByID deletes a size
// @Summary Delete a size by ID
// @Description Deletes a size by its ID. Requires JWT authentication.
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param size_id path string true "Size ID"
// @Router /api/market/sizes/{size_id} [delete]
func (h *Handler) deleteSizeByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	sizeID := vars["size_id"]
	if sizeID == "" {
		respondError(w, http.StatusBadRequest, "Missing size ID")
		return
	}

	err := h.db.DeleteSizeByID(claims.MarketID, sizeID)
	if err != nil {
		if err.Error() == "size not found or unauthorized" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Size deleted successfully"})
}


// deleteOrderByID deletes a order
// @Summary Delete a order by ID
// @Description Deletes a order by its ID. Requires JWT authentication.
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param order_id path string true "Order ID"
// @Router /api/market/orders/{order_id} [delete]
func (h *Handler) deleteOrderByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	orderID := vars["order_id"]
	if orderID == "" {
		respondError(w, http.StatusBadRequest, "Missing order ID")
		return
	}

	err := h.db.DeleteOrderByID(claims.MarketID, orderID)
	if err != nil {
		if err.Error() == "order not found or unauthorized" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Order deleted successfully"})
}


// deleteThumbnail deletes a thumbnail
// @Summary Delete a thumbnail
// @Description Deletes a thumbnail by its ID. Requires JWT authentication.
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param thumbnail_id path string true "Thumbnail ID"
// @Router /api/market/thumbnails/{thumbnail_id} [delete]
func (h *Handler) deleteThumbnail(w http.ResponseWriter, r *http.Request) {
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

	err := h.db.DeleteThumbnail(claims.MarketID, thumbnailID)
	if err != nil {
		if err.Error() == "thumbnail not found or unauthorized" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Thumbnail deleted successfully"})
}