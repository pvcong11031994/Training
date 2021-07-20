package infra

import (
	"log"
	"os"

	"gopkg.in/mgo.v2"
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

func InitConnectMongoDB() *mgo.Session {
	session, err := mgo.Dial(hosts)
	if err != nil {
		log.Fatalln(err)
		log.Fatalln("mongo err")
		os.Exit(1)
	}
	return session
}
