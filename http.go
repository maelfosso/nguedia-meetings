package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type MeetingStore interface {
	CheckAvailability(name string) (bool, error)
	CreateMeeting(meeting *Meeting) error
}

type HttpServer struct {
	store MeetingStore
	http.Handler
}

func (h HttpServer) JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

type Meeting struct {
	Name        string
	Description string
	Date        time.Time
}

func (h *HttpServer) CreateMeetingHandler(w http.ResponseWriter, r *http.Request) {
	var meeting Meeting

	if err := json.NewDecoder(r.Body).Decode(&meeting); err != nil {
		h.JSON(w, http.StatusInternalServerError, err)
		return
	}

	ok, err := h.store.CheckAvailability(meeting.Name)
	if err != nil {
		h.JSON(w, http.StatusBadRequest, fmt.Errorf("impossible to check the meeting name availability"))
		return
	}

	if !ok {
		h.JSON(w, http.StatusConflict, fmt.Errorf("meeting name is already used"))
		return
	}

	if err := h.store.CreateMeeting(&meeting); err != nil {
		h.JSON(w, http.StatusBadRequest, fmt.Errorf("impossible to create the meeting"))
		return
	}

	h.JSON(w, http.StatusCreated, meeting)
}
