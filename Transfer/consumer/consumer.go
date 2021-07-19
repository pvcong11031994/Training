package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gopkg.in/mgo.v2"
)

const (
	hosts              = "mongodb_consumer:27017"
	database           = "test"
	username           = ""
	password           = ""
	collection         = "restaurants"
	defaultKafkaBroker = "localhost:9092"
	defaultKafkaTopic  = "test"
)

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

var (
	mongoStore  = MongoStore{}
	kafkaServer string
	kafkaTopic  string
)

type MongoStore struct {
	session *mgo.Session
}

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

	// 2. Save msg MongoDB
	receiveFromKafka()
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

func receiveFromKafka() {
	fmt.Println("Start receiving from Kafka")
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        kafkaServer,
		"group.id":                 "group-id-1",
		"auto.offset.reset":        "earliest",
		"go.events.channel.enable": true,
	})
	if err != nil {
		fmt.Println("Cannot connect consumer: ", err)
		return
	}
	fmt.Println("Start SubscribeTopics")
	subscriptionErr := consumer.SubscribeTopics([]string{kafkaTopic}, nil)
	if subscriptionErr != nil {
		fmt.Println("Unable to subscribe to topic " + kafkaTopic + " due to error - " + subscriptionErr.Error())
		os.Exit(1)
	} else {
		fmt.Println("subscribed to topic ", kafkaTopic)
	}
	fmt.Println("SubscribeTopicsErr1", err)
	for {
		fmt.Println("waiting for event...")
		kafkaEvent := <-consumer.Events()
		if kafkaEvent != nil {
			switch event := kafkaEvent.(type) {
			case *kafka.Message:
				fmt.Println("Message " + string(event.Value))
				fmt.Printf("Received from Kafka %s: %s\n", event.TopicPartition, string(event.Value))
				job := string(event.Value)
				err = saveRestaurantToMongo(job)
				if err != nil {
					fmt.Println("Cannot save Mongo: ", err)
					return
				}
			case kafka.Error:
				fmt.Println("Consumer error ", event.String())
			case kafka.PartitionEOF:
				fmt.Println(kafkaEvent)
			default:
				fmt.Println(kafkaEvent)
			}
		} else {
			fmt.Println("Event was null")
		}
	}
}

func saveRestaurantToMongo(restaurantString string) (err error) {
	fmt.Println("Save to MongoDB")
	col := mongoStore.session.DB(database).C(collection)

	//Save data into Restaurant struct
	var _restaurant Restaurant
	b := []byte(restaurantString)
	err = json.Unmarshal(b, &_restaurant)
	if err != nil {
		return
	}
	//Insert job into MongoDB
	err = col.Insert(_restaurant)
	if err != nil {
		return
	}

	fmt.Printf("Saved to MongoDB : %s", restaurantString)
	return
}
