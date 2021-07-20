package model

type Restaurant struct {
	RestaurantId string  `json:"restaurant_id" bson:"restaurant_id"`
	Address      Address `json:"address" bson:"address"`
	Borough      string  `json:"borough" bson:"borough"`
	Cuisine      string  `json:"cuisine" bson:"cuisine"`
	Name         string  `json:"name" bson:"name"`
}

type Address struct {
	Building string `json:"building" bson:"building"`
	Street   string `json:"street" bson:"street"`
	Zipcode  string `json:"zipcode" bson:"zipcode"`
}
