package main

import (
	 "golang/routers"
	"log"
)

func main() {
	e := routers.New()
	log.Fatal(e.Start(":8080"))
}
