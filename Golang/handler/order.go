package handler

import (
	"github.com/labstack/echo/v4"
	"golang/entity"
	"golang/models"
	"net/http"
)

func CreateOrder(c echo.Context) error{
	order := &models.Orders{}
	if err := c.Bind(order); err != nil {
		return err
	}
	//Validate
	err := models.ValidateInputForm(*order)
	if err != nil {
		panic(err.Error())
	}
	result, err := entity.CreateOrder(*order)
	if err != nil {
		panic(err.Error())
	}
	return c.JSON(http.StatusCreated, result)
}
