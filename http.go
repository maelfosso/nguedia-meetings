package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type MeetingStore interface {
	CheckAvailability(name string) (bool, error)
	CreateMeeting()
}

type HttpServer struct {
	store MeetingStore
	http.Handler
}

func (h HttpServer) CreateMeetingHandler(w http.ResponseWriter, r *http.Request) {

}

func NewHttpServer(store MeetingStore) http.Handler {
	h := new(HttpServer)
	h.store = store

	r := mux.NewRouter()

	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	r.HandleFunc("/api/meetings", h.CreateMeetingHandler).Methods("POST")
	// r.HandleFunc("/api/meetings/{id}/members", UploadMembersHandler)
	// r.HandleFunc("/api/meetings/{id}/members/invite", InviteMemberHandler)
	// r.HandleFunc("/api/meetings/{id}/members/join", JoinMeetingHandler)

	r.NotFoundHandler = r.NewRoute().BuildOnly().HandlerFunc(http.NotFound).GetHandler()

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	handler := cors.AllowAll().Handler(loggedRouter)

	h.Handler = handler

	return h
}
