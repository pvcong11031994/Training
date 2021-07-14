package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
)

type Job struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Company     string `json:"company"`
	Salary      string `json:"salary"`
}

// const (
// 	bootstrapServersVar string = "BOOTSTRAP_SERVERS"
// )

// var bootstrapServers string = os.Getenv(bootstrapServersVar)

func main() {

	// producer, err := kafka.NewProducer(
	// 	&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})
	// if err != nil {
	// 	panic(fmt.Sprintf("Failed to create producer %s", err))
	// }
	// defer producer.Close()

	// go func() {
	// 	for e := range producer.Events() {
	// 		switch ev := e.(type) {
	// 		case *kafka.Message:
	// 			if ev.TopicPartition.Error != nil {
	// 				fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
	// 			} else {
	// 				fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
	// 			}
	// 		}
	// 	}
	// }()

	// topic := "messages"
	// keepRunning := true

	// sigchan := make(chan os.Signal, 1)
	// signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	// for keepRunning == true {
	// 	select {
	// 	case sig := <-sigchan:
	// 		log.Printf("Caught signal %v: terminating\n", sig)
	// 		keepRunning = false
	// 	default:
	// 		producer.Produce(&kafka.Message{
	// 			TopicPartition: kafka.TopicPartition{
	// 				Topic: &topic, Partition: kafka.PartitionAny},
	// 			Value: []byte("Hello World from Docker"),
	// 		}, nil)
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/jobs", jobsPostHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))

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
	_job.Company = "NONONO"

	//saveJobToKafka(_job)

	//Convert job struct into json
	jsonString, err := json.Marshal(_job)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//Set content-type http header
	w.Header().Set("content-type", "application/json")

	//Send back data as response
	w.Write(jsonString)

}

func saveJobToKafka(job Job) {

	fmt.Println("save to kafka")

	jsonString, err := json.Marshal(job)

	jobString := string(jsonString)
	fmt.Print(jobString)

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "kafka:9092"})
	if err != nil {
		panic(err)
	}

	// Produce messages to topic (asynchronously)
	topic := "jobs-topic1"
	for _, word := range []string{string(jobString)} {
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}, nil)
	}
}
