package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const NB_WORKERS = 4

type MeetingStore interface {
	CheckAvailability(name string) (bool, error)
	CreateMeeting(meeting *Meeting) error
	isMeetingExists(meetingID string) (bool, error)
	AddMember(meetingID string, member *Member) error
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
	r.HandleFunc("/api/meetings/{id}/members", h.UploadMembersHandler).Methods("POST")
	// r.HandleFunc("/api/meetings/{id}/members/invite", InviteMemberHandler)
	// r.HandleFunc("/api/meetings/{id}/members/join", JoinMeetingHandler)

	r.NotFoundHandler = r.NewRoute().BuildOnly().HandlerFunc(http.NotFound).GetHandler()

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	handler := cors.AllowAll().Handler(loggedRouter)

	h.Handler = handler

	return h
}

type Meeting struct {
	ID          primitive.ObjectID
	Name        string
	Description string
	Date        time.Time
}

type Member struct {
	ID             primitive.ObjectID
	Name           string
	PhoneNumber    string
	Email          string
	MembershipDate time.Time
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

func (h *HttpServer) UploadMembersHandler(w http.ResponseWriter, r *http.Request) {
	var members []Member

	vars := mux.Vars(r)
	meeting := vars["id"]

	if err := json.NewDecoder(r.Body).Decode(&members); err != nil {
		h.JSON(w, http.StatusInternalServerError, err)
		return
	}

	if exists, err := h.store.isMeetingExists(meeting); err != nil {
		h.JSON(w, http.StatusInternalServerError, fmt.Errorf("impossible to check if the meeting exists"))
		return
	} else if !exists {
		h.JSON(w, http.StatusBadRequest, fmt.Errorf("meeting does not exists"))
		return
	}

	lenJobs := len(members)
	jobs := make(chan Member, lenJobs)

	wg := sync.WaitGroup{}
	wg.Add(lenJobs)

	for w := 1; w <= NB_WORKERS; w++ {
		go h.addMeetingMembersWorker(r.Context(), meeting, jobs, &wg)
	}

	for _, member := range members {
		jobs <- member
	}

	close(jobs)

	wg.Wait()

	h.JSON(w, http.StatusCreated, members)
}

func (h *HttpServer) addMeetingMembersWorker(ctx context.Context, meeting string, members chan Member, wg *sync.WaitGroup) {
	for member := range members {
		select {
		case <-ctx.Done():
			return
		default:
			// h.store.AddMember(meeting, &member)
			if err := h.store.AddMember(meeting, &member); err != nil {

				// error if a member in that meeting
				// has an email or phone number or Name already used

				// the error is written in the channel so that we can know
				// if the member has been saved or not
				// it's important for the client
			}
			wg.Done()
		}
	}
}
