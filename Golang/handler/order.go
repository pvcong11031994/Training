package handler

import (
	"encoding/json"
	"golang/entity"
	helper "golang/helpers"
	"golang/models"
	"net/http"
)

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var order models.Orders
	_ = json.NewDecoder(r.Body).Decode(&order)
	//Validate
	err := models.ValidateInputForm(order)
	if err != nil {
		helper.BadRequestError(err, w)
		return
	}

	// Standardized Address

	result, err := entity.CreateOrder(order)
	if err != nil {
		helper.InternalError(err, w)
		return
	}
	json.NewEncoder(w).Encode(result)
}
