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
	database           = "db"
	username           = ""
	password           = ""
	collection         = "jobs"
	defaultKafkaBroker = "localhost:9092"
	defaultKafkaTopic  = "test"
)

type Job struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Company     string `json:"company"`
	Salary      string `json:"salary"`
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
		panic(err)
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
				saveJobToMongo(job)
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

func saveJobToMongo(jobString string) {
	fmt.Println("Save to MongoDB")
	col := mongoStore.session.DB(database).C(collection)

	//Save data into Job struct
	var _job Job
	b := []byte(jobString)
	err := json.Unmarshal(b, &_job)
	if err != nil {
		panic(err)
	}

	//Insert job into MongoDB
	errMongo := col.Insert(_job)
	if errMongo != nil {
		panic(errMongo)
	}

	fmt.Printf("Saved to MongoDB : %s", jobString)

}
