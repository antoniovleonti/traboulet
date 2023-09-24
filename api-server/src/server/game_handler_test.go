package server

import (
	"bytes"
	"encoding/json"
  "errors"
	"game"
	"github.com/r3labs/sse/v2"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"fmt"
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
		func() {}, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
	if err != nil {
		t.Error(err)
	}
	if gh == nil {
		t.Error("game handler is nil")
	}
}

func TestGetState(t *testing.T) {
	gh, err := newGameHandler(
		func() {}, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
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
  if rr.Code != http.StatusOK {
    t.Errorf("expected code %d, got %d", http.StatusOK, rr.Code)
  }

	var actual game.ClientView
	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&actual)

	if err != nil {
		t.Fatal(err)
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

	handleReqCheckEventStream(gh, postMoveReq, expectedStatus)
}

func TestPostValidMove(t *testing.T) {
	gh, err := newGameHandler(
		func() {}, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
	if err != nil {
		t.Fatal(err)
	}
	if gh == nil {
		t.Error("game handler is nil")
	}

	// Create the body
	move := game.Move{X: 0, Y: 0, D: game.DirRight}
	b, err := json.Marshal(move)
	if err != nil {
		t.Fatal(err)
	}

	postMove(t, gh, b, []*http.Cookie{gh.gm.GetWhiteCookie()}, http.StatusOK)
}

func TestPostInvalidMove(t *testing.T) {
	gh, err := newGameHandler(
		func() {}, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
	if err != nil {
		t.Fatal(err)
	}
	if gh == nil {
		t.Error("game handler is nil")
	}

	postMove(t, gh, []byte("blah"), []*http.Cookie{gh.gm.GetWhiteCookie()},
		http.StatusBadRequest)
}

func TestPostMoveNoCookie(t *testing.T) {
	gh, err := newGameHandler(
		func() {}, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
	if err != nil {
		t.Fatal(err)
	}
	if gh == nil {
		t.Error("game handler is nil")
	}

	// Create the body
	move := game.Move{X: 0, Y: 0, D: game.DirRight}
	b, err := json.Marshal(move)
	if err != nil {
		t.Fatal(err)
	}

	postMove(t, gh, b, []*http.Cookie{}, http.StatusUnauthorized)
}

func TestPostMoveEmptyBody(t *testing.T) {
	gh, err := newGameHandler(
		func() {}, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
	if err != nil {
		t.Fatal(err)
	}
	if gh == nil {
		t.Error("game handler is nil")
	}

	postMove(t, gh, []byte(""), []*http.Cookie{gh.gm.GetWhiteCookie()},
		http.StatusBadRequest)
}

func TestSendKeepAlive(t *testing.T) {
	gh, err := newGameHandler(
		func() {}, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(gh)
	url := server.URL

	msgsReceived := 0
	done := make(chan struct{})
	handleUpdate := func(msg *sse.Event) {
		msgsReceived++
		if string(msg.Event) != "keep-alive" {
			t.Errorf("Expected msg.Event == \"keep-alive\"; got %q", msg.Event)
		}
		log.Print("Sending done signal.")
		done <- struct{}{}
	}
	updateClient := sse.NewClient(url + "/event-stream")
	go func() {
		err = updateClient.SubscribeRaw(handleUpdate)
		if err != nil {
			t.Error(err)
		}
	}()

	log.Print("Waiting for subscriber to be ready.")
	for {
		if gh.pub.subscribers.Len() != 0 {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	gh.sendKeepAlive()
	log.Print("Waiting on done signal.")
	<-done
	if msgsReceived != 1 {
		t.Error("expected 1 message")
	}
}

func TestDeleteChallenge(t *testing.T) {
	cbCalled := false
	cb := func() {
		cbCalled = true
	}
	gh, _ := newGameHandler(
		cb, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())

	gh.DeleteChallenge()
	if !cbCalled {
		t.Error("callback not called")
	}
}

func handleReqCheckEventStream(
  gh *gameHandler, req *http.Request, expectedStatus int) error {
  // set up a server (sse client requires this)
	server := httptest.NewServer(gh)
	url := server.URL
  // Prepare event listener
	msgsReceived := 0
	done := make(chan struct{})
	handleUpdate := func(msg *sse.Event) {
		msgsReceived++
		log.Print(string(msg.Data))
		log.Print("sending done signal")
		done <- struct{}{}
	}
	updateClient := sse.NewClient(url + "/event-stream")
	go func() {
		updateClient.SubscribeRaw(handleUpdate)
	}()
	// Wait for the subscriber to get registered
	time.Sleep(5 * time.Millisecond)
	// Run the request through our handler
	resp := httptest.NewRecorder()
	gh.ServeHTTP(resp, req)
	time.Sleep(time.Millisecond)

	// Check the response body is what we expect.
	if resp.Code != expectedStatus {
		return fmt.Errorf(
      "handler returned unexpected status:\ngot: %d\nexpected: %d\nbody: %s",
			resp.Code, expectedStatus, resp.Body.String())
	}

	// Check we recieved a game update
	if expectedStatus == http.StatusOK {
		log.Print("awaiting done signal")
		<-done
		if msgsReceived != 1 {
			return errors.New("expected subscriber to receive exactly one message")
		}
	}
  return nil
}

func TestPostResignation(t *testing.T) {
	gh, _ := newGameHandler(
		nil, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())

  req, err := http.NewRequest("POST", "/resignation", nil)
  req.AddCookie(fakeWhiteCookie())

  err = handleReqCheckEventStream(gh, req, http.StatusOK)
  if err != nil {
    t.Error(err)
  }
}

func TestPostResignationNoCookie(t *testing.T) {
	gh, _ := newGameHandler(
		nil, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())

  req, err := http.NewRequest("POST", "/resignation", nil)

  err = handleReqCheckEventStream(gh, req, http.StatusUnauthorized)
  if err != nil {
    t.Error(err)
  }
}

func TestPostResignationTwice(t *testing.T) {
	gh, _ := newGameHandler(
		nil, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())

  req, err := http.NewRequest("POST", "/resignation", nil)
  req.AddCookie(fakeWhiteCookie())

  err = handleReqCheckEventStream(gh, req, http.StatusOK)
  if err != nil {
    t.Error(err)
  }

  // Will fail the second time because the game's already ended.
  err = handleReqCheckEventStream(gh, req, http.StatusBadRequest)
  if err != nil {
    t.Error(err)
  }
}
