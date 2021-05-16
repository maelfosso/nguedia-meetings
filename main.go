package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewMongoStore()
	// Need to add the store
	server := NewHttpServer(store)
	log.Fatal(http.ListenAndServe(":5000", server))
}
