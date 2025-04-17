package http

import (
	"github.com/dragondarkon/bqredis-crud/internal/usecase"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupRoutes configures the HTTP routes using Echo framework
func SetupRoutes(e *echo.Echo, userUseCase *usecase.UserUseCase) {
	// Add middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Create handler
	handler := NewUserHandler(userUseCase)

	// User routes
	e.GET("/users", handler.GetUsers)
	e.GET("/users/:id", handler.GetUser)
	e.POST("/users", handler.CreateUser)
	e.PUT("/users/:id", handler.UpdateUser)
	e.DELETE("/users/:id", handler.DeleteUser)
}
