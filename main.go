package main

import (
	"log"
	"net/http"
)

func main() {
	// Need to add the store
	server := NewHttpServer(nil)
	log.Fatal(http.ListenAndServe(":5000", server))
}
