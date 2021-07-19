package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	hosts              = "mongodb_producer:27017"
	database           = "test"
	username           = ""
	password           = ""
	collection         = "restaurants"
	defaultKafkaBroker = "localhost:9092"
	defaultKafkaTopic  = "test"
)

type Restaurant struct {
	RestaurantId string `json:"restaurant_id" bson:"restaurant_id"`
	Address Address `json:"address" bson:"address"`
	Borough string `json:"borough" bson:"borough"`
	Cuisine string `json:"cuisine" bson:"cuisine"`
	Name string `json:"name" bson:"name"`
}

type Address struct {
	Building string `json:"building" bson:"building"`
	Street string `json:"street" bson:"street"`
	Zipcode string `json:"zipcode" bson:"zipcode"`
}


type MongoStore struct {
	session *mgo.Session
}

var mongoStore = MongoStore{}
var kafkaServer, kafkaTopic string

func init() {
	kafkaServer = readFromENV("KAFKA_BROKER", defaultKafkaBroker)
	kafkaTopic = readFromENV("KAFKA_TOPIC", defaultKafkaTopic)

	fmt.Println("Kafka Broker - ", kafkaServer)
	fmt.Println("Kafka topic - ", kafkaTopic)
}

func readFromENV(key, defaultVal string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return value
}

func main() {
	// 1. Init connect mongoDB consumer
	session := initConnectMongoDB()
	defer session.Close()
	mongoStore.session = session

	// 2. Init router
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/restaurants", restaurantsPostHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8081", router))

}

func initConnectMongoDB() *mgo.Session {
	session, err := mgo.Dial(hosts)
	if err != nil {
		log.Fatalln(err)
		log.Fatalln("mongo err")
		os.Exit(1)
	}
	return session
}

func restaurantsPostHandler(w http.ResponseWriter, r *http.Request) {
	//Retrieve body from http request
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//Save data into Job struct
	var _restaurant Restaurant
	err = json.Unmarshal(b, &_restaurant)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Println("Start Save Kafka")

	// Read data MongoDB
	restaurants, err := readDataProducerFromMongo(_restaurant)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if len(restaurants) == 0 {
		err := errors.New("No data")
		http.Error(w, err.Error(), 400)
		return
	}
	for i, value := range restaurants {
		fmt.Println("Start Save Kafka: ", i)
		err = saveRestaurantToKafka(value)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Println("\nEnd Save Kafka: ", i)
		fmt.Println("------------------------")
	}
	if len(restaurants) > 0 {
		_restaurant = restaurants[0]
	}

	fmt.Println("End Save Kafka")
	//	_restaurant.Company = "SaveProducer"
	//Convert restaurantob struct into json
	jsonString, err := json.Marshal(_restaurant)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//	saveDataProducerToMongo(string(jsonString))

	//Set content-type http header
	w.Header().Set("content-type", "application/json")

	//Send back data as response
	w.Write(jsonString)

}
func readDataProducerFromMongo(restaurant Restaurant) ([]Restaurant, error) {
	// Query All
	col := mongoStore.session.DB(database).C(collection)
	var results []Restaurant
	query := bson.M{
		"restaurant_id": restaurant.RestaurantId,
	}
	err := col.Find(query).All(&results)
	if err != nil {
		return results, err
	}
	fmt.Println("Results All: ", results)
	return results, nil
}

// func saveDataProducerToMongo(restaurantString string) {
// 	fmt.Println("Save to MongoDB")
// 	col := mongoStore.session.DB(database).C(collection)

// 	//Save data into Job struct
// 	var _job Restaurant
// 	b := []byte(restaurantString)
// 	err := json.Unmarshal(b, &_job)
// 	if err != nil {
// 		panic(err)
// 	}

// 	//Insert job into MongoDB
// 	errMongo := col.Insert(_job)
// 	if errMongo != nil {
// 		panic(errMongo)
// 	}

// 	fmt.Printf("Saved to MongoDB : %s", restaurantString)

// }

func saveRestaurantToKafka(restaurant Restaurant) (err error) {
	jsonString, err := json.Marshal(restaurant)
	if err != nil {
		return
	}
	restaurantString := string(jsonString)
	fmt.Print(restaurantString)

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaServer})
	if err != nil {
		return
	}

	// Produce messages to topic (asynchronously)
	topic := kafkaTopic
	for _, word := range []string{string(restaurantString)} {
		err = p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}, nil)
	}
	return
}
