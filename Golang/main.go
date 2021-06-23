package main

import (
	handler "golang/handler"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/order", handler.CreateOrder).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", r))
}
