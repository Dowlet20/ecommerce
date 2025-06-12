package api

import (
	"Dowlet_projects/ecommerce/models"
	"net/http"
	"strconv"
	//"Dowlet_projects/ecommerce/services"
)

// getUserMessages retrieves all user messages
// @Summary Get all user messages
// @Description Retrieves all messages sent by users to superadmin. Requires superadmin JWT authentication.
// @Tags User Messages
// @Produce json
// @Security BearerAuth
// @Router /api/superadmin/user-messages [get]
func (h *Handler) getUserMessages(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.Role != "superadmin" {
		if !ok {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	messages, err := h.db.GetAllUserMessages()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, messages)
}


// getMarketMessages retrieves all market messages
// @Summary Get all market messages
// @Description Retrieves all messages sent by markets to superadmin. Requires superadmin JWT authentication.
// @Tags Market Messages
// @Produce json
// @Security BearerAuth
// @Router /api/superadmin/market-messages [get]
func (h *Handler) getMarketMessages(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.Role != "superadmin" {
		if !ok {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
		} else {
			respondError(w, http.StatusForbidden, "Forbidden")
		}
		return
	}

	messages, err := h.db.GetAllMarketMessages()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, messages)
}
// getUsers retrieves paginated users for superadmin
// @Summary     Get all users
// @Description Retrieves a paginated list of users with optional search by full_name, phone, or ID. Requires superadmin authentication.
// @Tags        Superadmin
// @Produce     json
// @Param       page  query integer false "Page number (default: 1)"
// @Param       limit query integer false "Items per page (default: 10)"
// @Param       search query string false "Search term to filter by full_name, phone, or ID"
// @Security    BearerAuth
// @Router      /api/superadmin/users [get]
func (h *Handler) getUsers(w http.ResponseWriter, r *http.Request) {
    // Verify superadmin
    claims, ok := r.Context().Value("claims").(*models.Claims)
    if !ok || claims.Role != "superadmin" {
        respondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    // Parse query parameters
    pageStr := r.URL.Query().Get("page")
    limitStr := r.URL.Query().Get("limit")
    search := r.URL.Query().Get("search")

    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        page = 1
    }

    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit < 1 {
        limit = 10
    }

    // Fetch users
    users, err := h.db.GetUsers(r.Context(), page, limit, search)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondJSON(w, http.StatusOK, users)
}