package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ErrorResponse struct {
	ErrorMessage string `json:"message"`
	StatusCode   int    `json:"status"`
}

func ConnectDB() *mongo.Collection {
	connectionString := "mongodb://mongo2:30001"
	clientOptions := options.Client().ApplyURI(connectionString)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	collection := client.Database("sample").Collection("orders")
	return collection
}
func InternalError(err error, w http.ResponseWriter) {
	var response = ErrorResponse{
		ErrorMessage: err.Error(),
		StatusCode:   http.StatusInternalServerError,
	}

	message, _ := json.Marshal(response)

	w.WriteHeader(response.StatusCode)
	w.Write(message)
}

func BadRequestError(err error, w http.ResponseWriter) {
	response := ErrorResponse{
		ErrorMessage: err.Error(),
		StatusCode:   http.StatusBadRequest,
	}
	message, _ := json.Marshal(response)

	w.WriteHeader(response.StatusCode)
	w.Write(message)
}
