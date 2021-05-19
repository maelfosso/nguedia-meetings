package main

import (
	"log"
	"net/http"
)

func main() {
	OpenDB()
	defer CloseDB()

	store := NewMongoStore(Database())
	// Need to add the store
	server := NewHttpServer(store)
	log.Fatal(http.ListenAndServe(":5000", server))
}
