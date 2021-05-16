package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
)

type StubMeetingStore struct {
	meetings []Meeting
}

func (s *StubMeetingStore) CheckAvailability(name string) (bool, error) {
	for _, meeting := range s.meetings {
		if meeting.Name == name {
			return false, nil
		}
	}

	return true, nil
}

func (s *StubMeetingStore) CreateMeeting(meeting *Meeting) error {
	s.meetings = append(s.meetings, *meeting)

	return nil
}

func marshalling(data interface{}) string {
	m, _ := json.Marshal(data)
	return string(m)
}

func TestCreateMeeting(t *testing.T) {
	store := StubMeetingStore{
		meetings: []Meeting{
			{
				Name:        faker.Name(),
				Description: faker.Paragraph(),
				Date:        time.Now(),
			},
		},
	}
	server := NewHttpServer(&store)

	t.Run("creates a meeting", func(t *testing.T) {
		var meeting = Meeting{
			Name:        "CLUB 15",
			Description: "Greet",
			Date:        time.Now(),
		}
		request := newCreateMeetingRequest(meeting)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Body.String()
		want := marshalling(meeting)

		assertStatus(t, response.Code, http.StatusCreated)
		assertResponseBody(t, got, want)
	})

	t.Run("return error if meeting name not available", func(t *testing.T) {
		var meeting = Meeting{
			Name:        store.meetings[0].Name,
			Description: "Greet",
			Date:        time.Now(),
		}
		request := newCreateMeetingRequest(meeting)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusConflict)
	})
}

func newCreateMeetingRequest(meeting Meeting) *http.Request {
	m, _ := json.Marshal(meeting)
	request, _ := http.NewRequest(http.MethodPost, "/api/meetings", bytes.NewBuffer(m))
	return request
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("got status %d want %d", got, want)
	}
}

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()

}
