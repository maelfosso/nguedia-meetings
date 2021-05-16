package main

import (
	"log"
	"net/http"
)

func main() {
	server := NewHttpServer(nil)
	log.Fatal(http.ListenAndServe(":5000", server))
}
