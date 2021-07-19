package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	hosts              = "mongodb_producer:27017"
	database           = "db"
	username           = ""
	password           = ""
	collection         = "producer_jobs"
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

	// 2. Init router
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/jobs", jobsPostHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8081", router))

}

func initConnectMongoDB() *mgo.Session {
	fmt.Println("CONGPVINITMONGO")
	session, err := mgo.Dial(hosts)
	if err != nil {
		log.Fatalln(err)
		log.Fatalln("mongo err")
		os.Exit(1)
	}
	return session
}

func jobsPostHandler(w http.ResponseWriter, r *http.Request) {
	//Retrieve body from http request
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}

	//Save data into Job struct
	var _job Job
	err = json.Unmarshal(b, &_job)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Println("CONGPV")

	// Read data MongoDB
	jobs := readDataProducerFromMongo(_job)
	for i, value := range jobs {
		fmt.Println("Start Save Job", i)
		saveJobToKafka(value)
		fmt.Println("End Save Job", i)
		fmt.Println("------------------------\n")
	}
	if len(jobs) > 0 {
		_job = jobs[0]
	}
	//	_job.Company = "SaveProducer"
	//Convert job struct into json
	jsonString, err := json.Marshal(_job)
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
func readDataProducerFromMongo(job Job) []Job {
	// Query All
	col := mongoStore.session.DB(database).C(collection)
	var results []Job
	query := bson.M{
		"title": job.Title,
	}
	err := col.Find(query).All(&results)
	if err != nil {
		panic(err)
	}
	fmt.Println("Results All: ", results)
	return results
}

func saveDataProducerToMongo(jobString string) {
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

func saveJobToKafka(job Job) {
	fmt.Println("save to kafka")

	jsonString, err := json.Marshal(job)

	jobString := string(jsonString)
	fmt.Print(jobString)

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaServer})
	if err != nil {
		panic(err)
	}
	fmt.Println("CONGPVxx", p)

	// Produce messages to topic (asynchronously)
	topic := kafkaTopic
	for _, word := range []string{string(jobString)} {
		err = p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}, nil)
		fmt.Println("CONGPV_OK2", err)
	}
	fmt.Println("CONGPV_OK")
}
