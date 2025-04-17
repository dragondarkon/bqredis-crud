package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dragondarkon/bqredis-crud/internal/domain/entity"
	"github.com/dragondarkon/bqredis-crud/internal/usecase"
	"github.com/labstack/echo/v4"
)

// Error response structure
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Common error codes
const (
	ErrCodeValidation = "VALIDATION_ERROR"
	ErrCodeNotFound   = "NOT_FOUND"
	ErrCodeInternal   = "INTERNAL_ERROR"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userUseCase *usecase.UserUseCase
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// handleError standardizes error responses
func handleError(c echo.Context, err error) error {
	var response ErrorResponse

	switch {
	case errors.Is(err, usecase.ErrUserNotFound):
		response = ErrorResponse{
			Code:    ErrCodeNotFound,
			Message: "User not found",
		}
		return c.JSON(http.StatusNotFound, response)
	case errors.Is(err, usecase.ErrValidation):
		response = ErrorResponse{
			Code:    ErrCodeValidation,
			Message: err.Error(),
		}
		return c.JSON(http.StatusBadRequest, response)
	default:
		response = ErrorResponse{
			Code:    ErrCodeInternal,
			Message: "Internal server error",
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
}

// GetUsers handles GET /users
func (h *UserHandler) GetUsers(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse pagination parameters
	page := 1
	pageSize := 10

	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.QueryParam("pageSize"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			pageSize = s
		}
	}

	users, err := h.userUseCase.GetAllUsers(ctx, page, pageSize)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": users,
		"pagination": map[string]int{
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	user, err := h.userUseCase.GetUserByID(ctx, id)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, user)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()
	var user entity.User

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    ErrCodeValidation,
			Message: "Invalid request payload",
		})
	}

	createdUser, err := h.userUseCase.CreateUser(ctx, user)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusCreated, createdUser)
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var user entity.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    ErrCodeValidation,
			Message: "Invalid request payload",
		})
	}

	// Ensure ID matches
	user.ID = id

	updatedUser, err := h.userUseCase.UpdateUser(ctx, user)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, updatedUser)
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.userUseCase.DeleteUser(ctx, id); err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
}
