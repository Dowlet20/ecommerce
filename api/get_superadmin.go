package api

import (

	"net/http"
	"Dowlet_projects/ecommerce/models"
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