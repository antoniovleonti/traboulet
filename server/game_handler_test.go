package server

import (
	"bytes"
	"encoding/json"
	"github.com/r3labs/sse/v2"
	"kuba"
	"log"
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
		kuba.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
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
		kuba.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
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
	server := httptest.NewServer(gh)
	url := server.URL

	// Build request
	postMoveReq, err := http.NewRequest("POST", "/move", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cookies {
		postMoveReq.AddCookie(c)
	}

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
		err = updateClient.SubscribeRaw(handleUpdate)
		if err != nil {
			t.Error(err)
		}
	}()
	// Wait for the subscriber to get registered
	time.Sleep(2 * time.Millisecond)
	// Run the request through our handler
	postMoveResp := httptest.NewRecorder()
	gh.ServeHTTP(postMoveResp, postMoveReq)
	time.Sleep(time.Millisecond)

	// Check the response body is what we expect.
	if postMoveResp.Code != expectedStatus {
		t.Errorf("handler returned unexpected status:\ngot: %d\nexpected: %d\n",
			postMoveResp.Code, expectedStatus)
		t.Errorf("returned body: %s\n", postMoveResp.Body.String())
	}

	// Check we recieved a game update
	if expectedStatus == http.StatusOK {
		log.Print("awaiting done signal")
		<-done
		if msgsReceived != 1 {
			t.Error("expected subscriber to receive exactly one message")
		}
	}
}

func TestPostValidMove(t *testing.T) {
	gh, err := newGameHandler(
		kuba.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
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
		kuba.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
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
		kuba.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
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
		kuba.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())
	if err != nil {
		t.Fatal(err)
	}
	if gh == nil {
		t.Error("game handler is nil")
	}

	postMove(t, gh, []byte(""), []*http.Cookie{gh.km.GetWhiteCookie()},
		http.StatusBadRequest)
}

func TestSendKeepAlive(t *testing.T) {
	gh, err := newGameHandler(
		kuba.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
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
