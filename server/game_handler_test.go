package server

import (
	"bytes"
	"encoding/json"
	"kuba"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
  "time"
	// "fmt"
)

func fakeWhiteCookie() *http.Cookie {
	c := http.Cookie{
		Name:  "white",
		Value: "1234",
		Path:  "/",
	}
	return &c
}

func fakeBlackCookie() *http.Cookie {
	c := http.Cookie{
		Name:  "black",
		Value: "5678",
		Path:  "/",
	}
	return &c
}

func TestNewGameHandler(t *testing.T) {
	// Make sure gameHandler implements the http.Handler interface
	var _ http.Handler = (*gameHandler)(nil)

	gh, err := newGameHandler(
    kuba.Config{TimeControl: 1*time.Minute}, fakeWhiteCookie(),
    fakeBlackCookie(), nil)
  if err != nil {
    t.Error(err)
  }
	if gh == nil {
		t.Error("game handler is nil")
	}
}

func TestGetState(t *testing.T) {
	gh, err := newGameHandler(
    kuba.Config{TimeControl: 1*time.Minute}, fakeWhiteCookie(),
    fakeBlackCookie(), nil)
  if err != nil {
    t.Fatal(err)
  }
	if gh == nil {
		t.Error("game handler is nil")
	}

	req, err := http.NewRequest("GET", "/state", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	gh.ServeHTTP(rr, req)
	// Check the response body is what we expect.
	var actual kuba.ClientView
	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&actual)
	if err != nil {
		t.Fatal(err)
	}
	expected := gh.km.GetClientView()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("handler returned unexpected body:\ngot: %v\nexpected: %v\n",
			actual, expected)
	}
}

func postMove(t *testing.T, gh *gameHandler, body []byte,
	cookies []*http.Cookie, expectedStatus int) {
	// Build request
	postMoveReq, err := http.NewRequest("POST", "/move", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cookies {
		postMoveReq.AddCookie(c)
	}

	// Add a subscriber to game updates
	getGameUpdateReq, err := http.NewRequest("GET", "/update", nil)
	if err != nil {
		t.Fatal(err)
	}
	getGameUpdateResp := httptest.NewRecorder()
	go gh.ServeHTTP(getGameUpdateResp, getGameUpdateReq)
  // just wait a little for the subscriber to be ready for pushes
  time.Sleep(time.Millisecond)

	// Run the request through our handler
	postMoveResp := httptest.NewRecorder()
	gh.ServeHTTP(postMoveResp, postMoveReq)

	// Check the response body is what we expect.
	if postMoveResp.Code != expectedStatus {
		t.Errorf("handler returned unexpected status:\ngot: %d\nexpected: %d\n",
			postMoveResp.Code, expectedStatus)
		t.Errorf("returned body: %s\n", postMoveResp.Body.String())
	}

	// Check we recieved a game update
	if expectedStatus == http.StatusOK {
		if getGameUpdateResp.Code != http.StatusOK {
			t.Errorf("expected 200 status from game update request; got %d",
				getGameUpdateResp.Code)
		}
		if len(gh.pub.subscribers) != 0 {
			t.Error("expected subscriber list to be empty")
		}
	} else {
		if len(gh.pub.subscribers) != 1 {
			t.Error("expected subscriber list to be exactly 1")
		}
	}
}

func TestPostValidMove(t *testing.T) {
	gh, err := newGameHandler(
    kuba.Config{TimeControl: 1*time.Minute}, fakeWhiteCookie(),
    fakeBlackCookie(), nil)
  if err != nil {
    t.Fatal(err)
  }
	if gh == nil {
		t.Error("game handler is nil")
	}

	// Create the body
	move := kuba.Move{X: 0, Y: 0, D: kuba.DirRight}
	b, err := json.Marshal(move)
	if err != nil {
		t.Fatal(err)
	}

	postMove(t, gh, b, []*http.Cookie{gh.km.GetWhiteCookie()}, http.StatusOK)
}

func TestPostInvalidMove(t *testing.T) {
	gh, err := newGameHandler(
    kuba.Config{TimeControl: 1*time.Minute}, fakeWhiteCookie(),
    fakeBlackCookie(), nil)
  if err != nil {
    t.Fatal(err)
  }
	if gh == nil {
		t.Error("game handler is nil")
	}

	postMove(t, gh, []byte("blah"), []*http.Cookie{gh.km.GetWhiteCookie()},
		http.StatusBadRequest)
}

func TestPostMoveNoCookie(t *testing.T) {
	gh, err := newGameHandler(
    kuba.Config{TimeControl: 1*time.Minute}, fakeWhiteCookie(),
    fakeBlackCookie(), nil)
  if err != nil {
    t.Fatal(err)
  }
	if gh == nil {
		t.Error("game handler is nil")
	}

	// Create the body
	move := kuba.Move{X: 0, Y: 0, D: kuba.DirRight}
	b, err := json.Marshal(move)
	if err != nil {
		t.Fatal(err)
	}

	postMove(t, gh, b, []*http.Cookie{}, http.StatusUnauthorized)
}

func TestPostMoveEmptyBody(t *testing.T) {
	gh, err := newGameHandler(
    kuba.Config{TimeControl: 1*time.Minute}, fakeWhiteCookie(),
    fakeBlackCookie(), nil)
  if err != nil {
    t.Fatal(err)
  }
	if gh == nil {
		t.Error("game handler is nil")
	}

	postMove(t, gh, []byte(""), []*http.Cookie{gh.km.GetWhiteCookie()},
		http.StatusBadRequest)
}
