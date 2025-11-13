package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/pkg/httputil"
	"github.com/rs/zerolog/log"
)

type AuthService interface {
	LoginViaPassword(ctx context.Context, nickname, password string) (string, error)
	RegisterUser(ctx context.Context, user *entities.User, password string) error
}

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{
		authService: service,
	}
}

// LoginRequest represents the expected payload for the login endpoint.
// swagger:model LoginRequest
type LoginRequest struct {
	Nickname string `json:"nickname" example:"johnny"`
	Password string `json:"password" example:"super-secret"`
}

// LoginResponse represents the successful login response.
// swagger:model LoginResponse
type LoginResponse struct {
	Token string `json:"token" example:"jwt-token"`
}

// Login godoc
// @Summary Login with nickname and password
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login payload"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500
// @Router /v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "invalid json"})
		return
	}

	if req.Nickname == "" || req.Password == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "nickname and password are required"})
		return
	}

	token, err := h.authService.LoginViaPassword(r.Context(), req.Nickname, req.Password)
	if err != nil {
		if errors.Is(err, entities.ErrInvalidCredentials) {
			httputil.JSON(w, http.StatusUnauthorized, ErrorResponse{Message: "invalid credentials"})
			return
		}

		log.Error().Err(err).Str("nickname", req.Nickname).Msg("login via password failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputil.JSON(w, http.StatusOK, LoginResponse{Token: token})
}

// RegisterUserRequest represents the expected payload for user registration.
// swagger:model RegisterUserRequest
type RegisterUserRequest struct {
	Nickname    string `json:"nickname"    example:"johnny"`
	Password    string `json:"password"    example:"super-secret"`
	PhoneNumber string `json:"phoneNumber" example:"+77010000000"`
	FirstName   string `json:"firstName"   example:"John"`
	// LastName of the registering user
	LastName string `json:"lastName"    example:"Doe"`
}

// RegisterUserResponse represents the response body returned after a successful registration.
// swagger:model RegisterUserResponse
type RegisterUserResponse struct {
	ID          string `json:"id"          example:"user-123"`
	Nickname    string `json:"nickname"    example:"johnny"`
	PhoneNumber string `json:"phoneNumber" example:"+77010000000"`
	FirstName   string `json:"firstName"   example:"John"`
	LastName    string `json:"lastName"    example:"Doe"`
}

// RegisterUser godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param register body RegisterUserRequest true "Registration payload"
// @Success 201 {object} RegisterUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500
// @Router /v1/auth/register [post]
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "invalid json"})
		return
	}

	if req.Nickname == "" || req.Password == "" || req.PhoneNumber == "" || req.FirstName == "" || req.LastName == "" {
		httputil.JSON(
			w,
			http.StatusBadRequest,
			ErrorResponse{Message: "nickname, password, phoneNumber, firstName and lastName are required"},
		)
		return
	}

	user := &entities.User{
		ID:          uuid.New().String(),
		Nickname:    req.Nickname,
		PhoneNumber: req.PhoneNumber,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
	}

	if err := h.authService.RegisterUser(r.Context(), user, req.Password); err != nil {
		log.Error().Err(err).Str("nickname", req.Nickname).Msg("register user failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputil.JSON(w, http.StatusCreated, RegisterUserResponse{
		ID:          user.ID,
		Nickname:    user.Nickname,
		PhoneNumber: user.PhoneNumber,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
	})
}
