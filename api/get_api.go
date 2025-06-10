package api

import (
	"net/http"
	"strconv"
	"fmt"
	"Dowlet_projects/ecommerce/models"
	"github.com/gorilla/mux"
	"encoding/json"
)


// getMarkets returns all markets
// @Summary Get all markets
// @Description Retrieves a list of all markets. Requires JWT authentication.
// @Tags Markets
// @Accept json
// @Produce json
// @Param duration query integer false "Duration day for new (default: 7)"
// @Param is_new query boolean false "Filter new markets (default: false)"
// @Param is_vip query boolean false "Filter vip markets (default: false)"
// @Router /markets [get]
func (h *Handler) getMarkets(w http.ResponseWriter, r *http.Request) {

	isNewStr := r.URL.Query().Get("is_new")

	isNew, err := strconv.ParseBool(isNewStr)
	if err != nil {
		isNew = false
	}

	isVipStr := r.URL.Query().Get("is_vip")

	isVip, err := strconv.ParseBool(isVipStr)
	if err != nil {
		isVip = false
	}

	durationStr := r.URL.Query().Get("duration")

	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 {
		duration = 7
	}

	markets, err := h.db.GetMarkets(r.Context(), isNew, isVip, duration)
	fmt.Println(err)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, markets)
}



// getMarketProducts retrieves products for the market admin
// @Summary Get market products
// @Description Retrieves a paginated list of products for the market admin's market. Requires JWT authentication.
// @Tags Products
// @Produce json
// @Param page query integer false "Page number (default: 1)"
// @Param limit query integer false "Items per page (default: 10)"
// @Router /products [get]
func (h *Handler) getMarketProducts(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*models.Claims)
	if !ok || claims.MarketID == 0 || claims.Role != "market_admin" {
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

	products, err := h.db.GetMarketProducts(r.Context(), claims.MarketID, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, products)
}



// getAllProducts retrieves all paginated products with optional filters and sorting
// @Summary Get all products
// @Description Retrieves paginated products with optional category, price range, discount, new status, sorting, name search, or random selection for homepage
// @Tags Products
// @Produce json
// @Param category_id query string false "Category ID"
// @Param duration query integer false "Duration day for new (default: 7)"
// @Param page query integer false "Page number (default: 1, ignored if random=true)"
// @Param limit query integer false "Items per page (default: 10, ignored if random=true)"
// @Param search query string false "Search by product name"
// @Param random query boolean false "Return 10 random products (default: false)"
// @Param start_price query number false "Minimum final price"
// @Param end_price query number false "Maximum final price"
// @Param sort query string false "Sort order: cheap_to_expensive, expensive_to_cheap (default: by ID)"
// @Param has_discount query boolean false "Filter products with discount (default: false)"
// @Param is_new query boolean false "Filter new products (default: false)"
// @Router /products [get]
func (h *Handler)  getAllProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	categoryIDStr := r.URL.Query().Get("category_id")
	var categoryID int
	if categoryIDStr != "" {
		var err error
		categoryID, err = strconv.Atoi(categoryIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid category ID")
			return
		}
	}

	durationStr := r.URL.Query().Get("duration")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")
	randomStr := r.URL.Query().Get("random")
	startPriceStr := r.URL.Query().Get("start_price")
	endPriceStr := r.URL.Query().Get("end_price")
	sort := r.URL.Query().Get("sort")
	hasDiscountStr := r.URL.Query().Get("has_discount")
	isNewStr := r.URL.Query().Get("is_new")

	// Validate duration
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 {
		duration = 7
	}

	// Validate page
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// Validate limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Validate random
	random, err := strconv.ParseBool(randomStr)
	if err != nil {
		random = false
	}
	if random {
		page = 1
		limit = 10
	}

	// Validate start_price
	var startPrice float64
	if startPriceStr != "" {
		startPrice, err = strconv.ParseFloat(startPriceStr, 64)
		if err != nil || startPrice < 0 {
			respondError(w, http.StatusBadRequest, "Invalid start price")
			return
		}
	}

	// Validate end_price
	var endPrice float64
	if endPriceStr != "" {
		endPrice, err = strconv.ParseFloat(endPriceStr, 64)
		if err != nil || endPrice < 0 {
			respondError(w, http.StatusBadRequest, "Invalid end price")
			return
		}
	}

	// Validate start_price <= end_price
	if startPriceStr != "" && endPriceStr != "" && startPrice > endPrice {
		respondError(w, http.StatusBadRequest, "Start price must be less than or equal to end price")
		return
	}

	// Validate sort
	if sort != "" && sort != "cheap_to_expensive" && sort != "expensive_to_cheap" {
		respondError(w, http.StatusBadRequest, "Invalid sort value; must be cheap_to_expensive or expensive_to_cheap")
		return
	}

	// Validate has_discount
	hasDiscount, err := strconv.ParseBool(hasDiscountStr)
	if err != nil {
		hasDiscount = false
	}

	// Validate is_new
	isNew, err := strconv.ParseBool(isNewStr)
	if err != nil {
		isNew = false
	}

	products, err := h.db.GetPaginatedProducts(r.Context(), categoryID, duration, page, limit, search, random, startPrice, endPrice, sort, hasDiscount, isNew)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, products)
}


// getProduct returns a single product by ID
// @Summary Get a product by ID
// @Description Retrieves a single product by its ID. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Router /products/{id} [get]
func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	product, err := h.db.GetProduct(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, product)
}



// getAllThumbnails retrieves all thumbnails with product information
// @Summary Get all thumbnails with product details
// @Description Retrieves a list of all thumbnails with associated product information. Requires JWT authentication.
// @Tags Products
// @Accept json
// @Produce json
// @Router /thumbnails [get]
func (h *Handler) getAllThumbnails(w http.ResponseWriter, r *http.Request) {
	thumbnails, err := h.db.GetAllThumbnailsWithProducts()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, thumbnails)
}



// getCategories retrieves paginated categories with search
// @Summary Get categories
// @Description Retrieves a paginated list of categories with optional name search. Requires superadmin JWT authentication.
// @Tags Categories
// @Produce json
// @Param page query integer false "Page number (default: 1)"
// @Param limit query integer false "Items per page (default: 10)"
// @Param search query string false "Search by category name"
// @Router /categories [get]
func (h *Handler) getCategories(w http.ResponseWriter, r *http.Request) {
	// claims, ok := r.Context().Value("claims").(*models.Claims)
	// if !ok || claims.Role != "superadmin" {
	// 	respondError(w, http.StatusUnauthorized, "Unauthorized")
	// 	return
	// }

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

	categories, err := h.db.GetCategories(page, limit, search)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, categories)
}



// getBanners retrieves all banners
// @Summary Get all banners
// @Description Retrieves all banners with their ID, description, and thumbnail URL. Public endpoint, no authentication required.
// @Tags Banners
// @Produce json
// @Router /banners [get]
func (h *Handler) getBanners(w http.ResponseWriter, r *http.Request) {
	banners, err := h.db.GetAllBanners()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, banners)
}



// getMarketByIDALL retrieves a market and its products by ID
// @Summary Get market by ID
// @Description Retrieves market details and its products by market ID with pagination.
// @Tags Markets
// @Produce json
// @Param id path string true "Market ID"
// @Param limit query integer false "Number of products per page (default: 20, min: 1, max: 100)"
// @Param page query integer false "Page number (default: 1, min: 1)"
// @Router /markets/{id} [get]
func (h *Handler) getMarketByIDALL(w http.ResponseWriter, r *http.Request) {
	// Get market ID from URL
	vars := mux.Vars(r)
	marketIDStr := vars["id"]
	marketID, err := strconv.Atoi(marketIDStr)
	if err != nil || marketID < 1 {
		respondError(w, http.StatusBadRequest, "Invalid market ID")
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
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			respondError(w, http.StatusBadRequest, "Invalid limit; must be an integer between 1 and 100")
			return
		}
	}

	// Parse page
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			respondError(w, http.StatusBadRequest, "Invalid page; must be an integer >= 1")
			return
		}
	}

	market, products, totalCount, err := h.db.GetMarketByID(r.Context(),marketID, page, limit)
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


func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Printf("Failed to encode JSON: %v\n", err)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}