package models

import "errors"

type Orders struct {
	Code          string  `json:"code" bson:"code"`
	Address       string  `json:"address" bson:"address"`
	CityId        int     `json:"cityId,omitempty" bson:"cityId,omitempty"`
	Phone         string  `json:"phone" bson:"phone"`
	PromotionCode string  `json:"promotionCode,omitempty" bson:"promotionCode,omitempty"`
	Total         float32 `json:"total" bson:"total"`
	Note          string  `json:"note,omitempty" bson:"note,omitempty"`
	CreatedBy     string  `json:"createdBy" bson:"createdBy"`
}

func ValidateInputForm(order Orders) (err error) {
	if order.Code == "" {
		return errors.New("code is requied not null")
	}
	if order.Address == "" {
		return errors.New("address is requied not null")
	}
	if order.Phone == "" {
		return errors.New("phone is requied not null")
	}
	if order.Total == 0 {
		return errors.New("total is requied not null")
	}
	if order.CreatedBy == "" {
		return errors.New("createdBy is requied not null")
	}
	return
}
