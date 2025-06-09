package api

import (
	//"context"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	//"io"
	"net/http"
	//"os"
	//"path/filepath"
	"regexp"
	//"strings"
	"time"
	//"strconv"

	"github.com/dgrijalva/jwt-go"
	//"github.com/gorilla/mux"
	//"github.com/rs/cors"
	"Dowlet_projects/ecommerce/models"
	"Dowlet_projects/ecommerce/services"
)

// RegisterRequest defines the request body for user registration
type RegisterRequest struct {
	FullName string `json:"full_name" example:"John Doe" description:"Full name of the user" validate:"required"`
	Phone    string `json:"phone" example:"+12345678901" description:"Phone number in international format" validate:"required"`
}

// LoginRequest defines the request body for login
type LoginRequest struct {
	Phone string `json:"phone" example:"+12345678901" description:"Phone number in international format" validate:"required"`
}

// VerifyCodeRequest defines the request body for verification
type VerifyCodeRequest struct {
	Phone string `json:"phone" example:"+12345678901" description:"Phone number in international format" validate:"required"`
	Code  string `json:"code" example:"1234" description:"4-digit verification code" validate:"required,len=4"`
}


// register handles user registration
// @Summary Register a new user
// @Description Registers a new user with a full name and phone number, sending a verification code. Validates input, checks for duplicate phone numbers, and logs the code (replace with SMS in production).
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "User registration details"
// @Router /register [post]
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.FullName == "" {
		respondError(w, http.StatusBadRequest, "Full name is required")
		return
	}
	if !validatePhone(req.Phone) {
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Check if phone exists
	_, err := h.db.GetUserByPhone(req.Phone)
	if err == nil {
		respondError(w, http.StatusConflict, "Phone number already registered")
		return
	}
	if err != sql.ErrNoRows {
		fmt.Println(err)
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Generate verification code
	code, err := services.GenerateVerificationCode(h.db, req.Phone, req.FullName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate verification code")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Verification code generated",
		"code":    code, // Remove in production; use SMS
	})
}

// login initiates user login
// @Summary Initiate user login
// @Description Initiates login by sending a verification code to the user's phone number. Checks if the phone is registered and logs the code.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body LoginRequest true "User login details"
// @Router /login [post]
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if !validatePhone(req.Phone) {
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}

	// Check if phone exists
	userID, err := h.db.GetUserByPhone(req.Phone)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Phone number not registered")
		return
	}
	if err != nil {
		fmt.Println(err)
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Generate verification code
	code, err := services.GenerateVerificationCode(h.db, req.Phone, "")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate verification code")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Verification code generated",
		"code":    code, // Remove in production; use SMS
		"user_id": userID,
	})
}


// loginSuperadmin authenticates a superadmin
// @Summary Login superadmin
// @Description Authenticates a superadmin and returns a JWT token.
// @Tags Superadmin
// @Accept json
// @Produce json
// @Param credentials body SuperadminRequest true "Superadmin credentials"
// @Router /superadmin/login [post]
func (h *Handler) loginSuperadmin(w http.ResponseWriter, r *http.Request) {
	var req SuperadminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Missing credentials")
		return
	}

	userID, err := h.db.AuthenticateSuperadmin(req.Username, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := h.generateJWT(userID, 0, "superadmin")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"token": token})
}

// loginMarket authenticates a market admin
// @Summary Login market admin
// @Description Authenticates a market admin and returns a JWT token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body MarketRequest true "Market admin credentials"
// @Router /market/login [post]
func (h *Handler) loginMarket(w http.ResponseWriter, r *http.Request) {
	var req MarketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.Phone == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Missing credentials")
		return
	}

	userID, marketID, err := h.db.AuthenticateMarket(req.Phone, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := h.generateJWT(userID, marketID, "market_admin")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"token": token})
}


// registerSuperadmin registers a new superadmin
// @Summary Register a new superadmin
// @Description Registers a new superadmin with username, full name, phone, and password.
// @Tags Superadmin
// @Accept json
// @Produce json
// @Param superadmin body models.SuperadminRegisterRequest true "Superadmin details"
// @Router /superadmin/register [post]
func (h *Handler) registerSuperadmin(w http.ResponseWriter, r *http.Request) {
	var req models.SuperadminRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Error parsing JSON body")
		return
	}

	if req.Username == "" || req.FullName == "" || req.Phone == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	superadminID, err := h.db.RegisterSuperadmin(req.Username, req.FullName, req.Phone, req.Password)
	if err != nil {
		if err.Error() == "username or phone already exists" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"superadmin_id": superadminID})
}

// verifyCode verifies the code and completes registration/login
// @Summary Verify phone number
// @Description Verifies a phone number using a 4-digit code, issuing a JWT token upon successful verification. Deletes the used code.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body VerifyCodeRequest true "Verification details"
// @Router /verify [post]
func (h *Handler) verifyCode(w http.ResponseWriter, r *http.Request) {
	var req VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if !validatePhone(req.Phone) {
		respondError(w, http.StatusBadRequest, "Invalid phone number format")
		return
	}
	if len(req.Code) != 4 || !regexp.MustCompile(`^\d{4}$`).MatchString(req.Code) {
		respondError(w, http.StatusBadRequest, "Invalid verification code")
		return
	}

	// Validate code
	storedCode, expiresAt, fullName, err := h.db.GetVerificationCode(req.Phone)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusBadRequest, "No verification code found")
		return
	}
	if err != nil {
		fmt.Println(err)
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if time.Now().After(expiresAt) {
		respondError(w, http.StatusBadRequest, "Verification code expired")
		return
	}

	if storedCode != req.Code {
		respondError(w, http.StatusBadRequest, "Invalid verification code")
		return
	}

	// Check if user is already registered
	userID, err := h.db.GetUserByPhone(req.Phone)
	var isRegistered bool
	if err == nil {
		isRegistered = true
	} else if err != sql.ErrNoRows {
		fmt.Println(err)
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// If not registered, complete registration
	if !isRegistered {
		if fullName == "" {
			respondError(w, http.StatusBadRequest, "Missing registration data")
			return
		}
		userID64, err := h.db.SaveUser(fullName, req.Phone)
		if err != nil {
			fmt.Println(err)
			respondError(w, http.StatusInternalServerError, "Failed to register user")
			return
		}
		userID = int(userID64)
	}

	// Delete used code
	if err := h.db.DeleteVerificationCode(req.Phone); err != nil {
		fmt.Printf("Failed to delete verification code: %v\n", err.Error())
	}

	// Generate JWT
	token, err := h.generateJWT(userID, 0, "user")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"token":   token,
	})
}

// authMiddleware authenticates JWT tokens
func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			respondError(w, http.StatusUnauthorized, "Authorization header missing")
			return
		}
		// Remove "Bearer " prefix
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("your-secret-key"), nil
		})
		if err != nil || !token.Valid {
			respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// // Add claims to context duzedildi.....
		// type contextKey string
		// const claimsKey contextKey = "claims"
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


// validatePhone checks if the phone number is valid
func validatePhone(phone string) bool {
	re := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return re.MatchString(phone)
}

// generateJWT creates a JWT token
func generateJWT(userID int) (string, error) {
	claims := &models.Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}
	return signedToken, nil
}


// generateJWT creates a JWT
func (h *Handler) generateJWT(userID, marketID int, role string) (string, error) {
	claims := &models.Claims{
		UserID:   userID,
		MarketID: marketID,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-secret-key"))
}




