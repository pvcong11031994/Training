package entity

import (
	"context"
	helper "golang/helpers"
	"golang/models"
)

var collection = helper.ConnectDB()

type OrderEntity interface {
	CreateOrder(order models.Orders) (interface{}, error)
}

func CreateOrder(order models.Orders) (interface{}, error) {
	result, err := collection.InsertOne(context.TODO(), order)
	if err != nil {
		return nil, err
	}
	return result, nil
}
