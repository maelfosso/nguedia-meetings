package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
)

type StubMeetingStore struct {
	meetings []Meeting
	members  map[string][]Member
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

func (s *StubMeetingStore) isMeetingExists(meeting string) (bool, error) {
	for _, m := range s.meetings {
		if m.Name == meeting {
			return true, nil
		}
	}

	return false, nil
}

func (s *StubMeetingStore) AddMember(meeting string, member *Member) error {
	s.members[meeting] = append(s.members[meeting], *member)

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
			Name:        faker.Name(),
			Description: faker.Paragraph(),
			Date:        time.Now(),
		}
		request := newCreateMeetingRequest(meeting)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := strings.TrimSpace(response.Body.String())
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
	if got != want {
		t.Errorf("\ngot %q \nwant %q", got, want)
	}
}

func TestUploadingMembers(t *testing.T) {
	store := StubMeetingStore{
		meetings: []Meeting{
			{
				Name:        faker.Name(),
				Description: faker.Paragraph(),
				Date:        time.Now(),
			},
		},
		members: map[string][]Member{},
	}
	server := NewHttpServer(&store)

	t.Run("add members to the meeting", func(t *testing.T) {
		members := make([]Member, 0)
		for i := 1; i <= 10; i++ {
			m := Member{
				Name:           faker.Name(),
				PhoneNumber:    faker.Phonenumber(),
				Email:          faker.Email(),
				MembershipDate: time.Now(),
			}
			members = append(members, m)
		}

		request := newUploadMembersRequest(store.meetings[0].Name, members)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusCreated)
	})

	t.Run("returns 400 if meeting does not exits", func(t *testing.T) {
		members := make([]Member, 0)
		for i := 1; i <= 10; i++ {
			m := Member{
				Name:           faker.Name(),
				PhoneNumber:    faker.Phonenumber(),
				Email:          faker.Email(),
				MembershipDate: time.Now(),
			}
			members = append(members, m)
		}

		request := newUploadMembersRequest(faker.Name(), members)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("return result with error if a member already exists", func(t *testing.T) {})
}

func newUploadMembersRequest(meeting string, members []Member) *http.Request {
	m, _ := json.Marshal(members)
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/meetings/%s/members", meeting), bytes.NewBuffer(m))
	return request
}
