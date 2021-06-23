package routers

import (
	"github.com/labstack/echo/v4"
	"golang/handler"
	"golang/middlewares"
)

func New() *echo.Echo {
	e := echo.New()
	e.GET("/user/signIn", handler.SignInForm()).Name = "userSignInForm"
	e.POST("/user/signIn", handler.SignIn())

	adminGroup := e.Group("/admin")
	adminGroup.Use(middlewares.TokenConfigMiddleware())
	// Attach jwt token refresher.
	adminGroup.Use(middlewares.TokenRefresherMiddleware)

	adminGroup.GET("", handler.Admin())
	adminGroup.POST("/order",  handler.CreateOrder)
	return e
}