package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type StubPlayerStore struct {
	scores   map[string]int
	winCalls []string
	league   []Player
}

func (s *StubPlayerStore) GetPlayerScore(name string) int {
	score := s.scores[name]
	return score
}

func (s *StubPlayerStore) RecordWin(name string) {
	s.winCalls = append(s.winCalls, name)
}

func (s *StubPlayerStore) GetLeague() League {
	return s.league
}

type StubCatStore struct {
	cats []Cat
}

func (s *StubCatStore) GetAll() []Cat {
	return s.cats
}

func (s *StubCatStore) GetByID(id int) *Cat {
	for _, cat := range s.GetAll() {
		if cat.ID == id {
			return &cat
		}
	}
	return nil
}

func TestCatsEndpoint(t *testing.T) {
	store := StubCatStore{
		[]Cat{
			{
				1,
				"Melinoe",
			},
			{
				2,
				"Salem",
			},
		},
	}

	server := NewServer(&store)

	t.Run("it returns statuscode 200 on /cats", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/cats", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, response.Code, http.StatusOK)
	})

	t.Run("it returns all cats as JSON", func(t *testing.T) {
		wantedCats := []Cat{
			{
				1,
				"Melinoe",
			},
			{
				2,
				"Salem",
			},
		}

		request, _ := http.NewRequest(http.MethodGet, "/cats", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		var got []Cat

		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Errorf("error parsing response into slice of cats, %v", err)
		}

		assertStatusCode(t, response.Code, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertCats(t, got, wantedCats)
	})

	t.Run("return melinoe", func(t *testing.T) {
		want := Cat{
			1,
			"Melinoe",
		}

		request, _ := http.NewRequest(http.MethodGet, "/cats/1", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		var got Cat

		json.NewDecoder(response.Body).Decode(&got)

		assertCat(t, got, want)
	})
}

func TestGETPlayers(t *testing.T) {
	store := StubPlayerStore{
		map[string]int{
			"Pepper": 20,
			"Floyd":  10,
		},
		nil,
		nil,
	}

	server := NewPlayerServer(&store)

	t.Run("returns Pepper's score", func(t *testing.T) {
		request := newGetScoreRequest("Pepper")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, response.Body.String(), "20")
	})

	t.Run("returns Floyd's score", func(t *testing.T) {
		request := newGetScoreRequest("Floyd")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, response.Code, http.StatusOK)
		assertResponseBody(t, response.Body.String(), "10")
	})

	t.Run("returns 404 on missing player", func(t *testing.T) {
		request := newGetScoreRequest("Apollo")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, response.Code, http.StatusNotFound)
	})
}

func TestStoreWins(t *testing.T) {
	store := StubPlayerStore{
		map[string]int{},
		nil,
		nil,
	}

	server := NewPlayerServer(&store)

	t.Run("it returns accept on POST", func(t *testing.T) {
		player := "Pepper"

		request := newPostWinRequest(player)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatusCode(t, response.Code, http.StatusAccepted)

		if len(store.winCalls) != 1 {
			t.Errorf("got %d calls to RecordWin, wanted %d", len(store.winCalls), 1)
		}

		if store.winCalls[0] != player {
			t.Errorf("did not store correct player. got %q, want %q", store.winCalls[0], player)
		}
	})
}

func TestLeague(t *testing.T) {
	store := StubPlayerStore{}
	server := NewPlayerServer(&store)

	t.Run("it returns 200 on /league", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/league", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		var got []Player

		err := json.NewDecoder(response.Body).Decode(&got)

		if err != nil {
			t.Fatalf("unable to parse reponse from server %q into slice of Player, '%v'", response.Body, err)
		}

		assertStatusCode(t, response.Code, http.StatusOK)
	})

	t.Run("it returns the league table as JSON", func(t *testing.T) {
		wantedLeague := []Player{
			{"Cleo", 32},
			{"Chris", 20},
			{"Tiets", 14},
		}
		store := StubPlayerStore{nil, nil, wantedLeague}
		server := NewPlayerServer(&store)

		request := newLeagueRequest()
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := getLeagueFromResponse(t, response.Body)

		assertContentType(t, response, jsonContentType)
		assertStatusCode(t, response.Code, http.StatusOK)
		assertLeague(t, got, wantedLeague)
	})
}

func newPostWinRequest(name string) *http.Request {
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/players/%s", name), nil)
	return request
}

func newGetScoreRequest(name string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/players/%s", name), nil)
	return request
}

func newLeagueRequest() *http.Request {
	request, _ := http.NewRequest(http.MethodGet, "/league", nil)
	return request
}

func getLeagueFromResponse(t testing.TB, body io.Reader) (league []Player) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&league)

	if err != nil {
		t.Fatalf("unable to parse reponse from server %q into slice of Player, '%v'", body, err)
	}
	return
}

func assertContentType(t testing.TB, response *httptest.ResponseRecorder, want string) {
	if response.Result().Header.Get("content-type") != "application/json" {
		t.Errorf("response did not have content-type of application/json, got %v", response.Result().Header)
	}
}

func assertLeague(t testing.TB, got, want []Player) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertCats(t testing.TB, got, want []Cat) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertCat(t testing.TB, got, want Cat) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertStatusCode(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("dit not get correct statuscode, got %d, want %d", got, want)
	}
}
