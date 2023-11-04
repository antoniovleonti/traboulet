package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"game"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
  "evtpub"
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
  _, chpub := GetTestPublishers()

	gh, err := newGameHandler(
		func() {}, *chpub, game.Config{TimeControl: 1 * time.Minute},
    fakeWhiteCookie(), fakeBlackCookie())
	if err != nil {
		t.Error(err)
	}
	if gh == nil {
		t.Error("game handler is nil")
	}
}

func TestGetState(t *testing.T) {
  _, chpub := GetTestPublishers()
	gh, err := newGameHandler(
		func() {}, *chpub, game.Config{TimeControl: 1 * time.Minute},
    fakeWhiteCookie(), fakeBlackCookie())
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
	log.Print(rr.Body.String())
	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&actual)

	if err != nil {
		t.Fatal(err)
	}
}

func postMove(
  t *testing.T, gh *gameHandler, evpub *evtpub.MockEventPublisher, body []byte,
  cookies []*http.Cookie, expectedStatus int) {

	// Build request
	postMoveReq, err := http.NewRequest("POST", "/move", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cookies {
		postMoveReq.AddCookie(c)
	}

	handleReqCheckEventStream(gh, evpub, postMoveReq, expectedStatus)
}

func TestPostValidMove(t *testing.T) {
  evpub, chpub := GetTestPublishers()
	gh, err := newGameHandler(
		func() {}, *chpub, game.Config{TimeControl: 1 * time.Minute},
    fakeWhiteCookie(), fakeBlackCookie())
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

	postMove(
    t, gh, evpub, b, []*http.Cookie{gh.gm.GetWhiteCookie()}, http.StatusOK)
}

func TestPostInvalidMove(t *testing.T) {
  evpub, chpub := GetTestPublishers()
	gh, err := newGameHandler(
		func() {}, *chpub, game.Config{TimeControl: 1 * time.Minute},
    fakeWhiteCookie(), fakeBlackCookie())
	if err != nil {
		t.Fatal(err)
	}
	if gh == nil {
		t.Error("game handler is nil")
	}

	postMove(t, gh, evpub, []byte("blah"), []*http.Cookie{gh.gm.GetWhiteCookie()},
		http.StatusBadRequest)
}

func TestPostMoveNoCookie(t *testing.T) {
  evpub, chpub := GetTestPublishers()
	gh, err := newGameHandler(
		func() {}, *chpub, game.Config{TimeControl: 1 * time.Minute},
    fakeWhiteCookie(), fakeBlackCookie())
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

	postMove(t, gh, evpub, b, []*http.Cookie{}, http.StatusUnauthorized)
}

func TestPostMoveEmptyBody(t *testing.T) {
  evpub, chpub := GetTestPublishers()
	gh, err := newGameHandler(
		func() {}, *chpub, game.Config{TimeControl: 1 * time.Minute},
    fakeWhiteCookie(), fakeBlackCookie())
	if err != nil {
		t.Fatal(err)
	}
	if gh == nil {
		t.Error("game handler is nil")
	}

	postMove(
    t, gh, evpub, []byte(""), []*http.Cookie{gh.gm.GetWhiteCookie()},
		http.StatusBadRequest)
}

func handleReqCheckEventStream(
	gh *gameHandler, evpub *evtpub.MockEventPublisher, req *http.Request,
  expectedStatus int) error {

  pushesBefore := len(evpub.Channels[testChannelPath].Pushes)

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
    // Check that event publisher recieved an update.
    if len(evpub.Channels[testChannelPath].Pushes) != pushesBefore + 1 {
      return errors.New("expected to publish exactly 1 event")
    }
	}

	return nil
}

func TestPostResignation(t *testing.T) {
  evpub, chpub := GetTestPublishers()
	gh, _ := newGameHandler(
		nil, *chpub, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())

	req, err := http.NewRequest("POST", "/resignation", nil)
	req.AddCookie(fakeWhiteCookie())

	err = handleReqCheckEventStream(gh, evpub, req, http.StatusOK)
	if err != nil {
		t.Error(err)
	}
}

func TestPostResignationNoCookie(t *testing.T) {
  evpub, chpub := GetTestPublishers()
	gh, _ := newGameHandler(
		nil, *chpub, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())

	req, err := http.NewRequest("POST", "/resignation", nil)

	err = handleReqCheckEventStream(gh, evpub, req, http.StatusUnauthorized)
	if err != nil {
		t.Error(err)
	}
}

func TestPostResignationTwice(t *testing.T) {
  evpub, chpub := GetTestPublishers()
	gh, _ := newGameHandler(
		nil, *chpub, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())

	req, err := http.NewRequest("POST", "/resignation", nil)
	req.AddCookie(fakeWhiteCookie())

	err = handleReqCheckEventStream(gh, evpub, req, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

	// Will fail the second time because the game's already ended.
	err = handleReqCheckEventStream(gh, evpub, req, http.StatusBadRequest)
	if err != nil {
		t.Error(err)
	}
}

func TestPostRematchOffer(t *testing.T) {
  evpub, chpub := GetTestPublishers()
	gh, _ := newGameHandler(
		nil, *chpub, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())

  // end the game
	req, err := http.NewRequest("POST", "/resignation", nil)
	req.AddCookie(fakeWhiteCookie())

	err = handleReqCheckEventStream(gh, evpub, req, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

  whiteRematchReq, err := http.NewRequest("POST", "/rematch-offer", nil)
  if err != nil {
    t.Error(nil)
  }
  whiteRematchReq.AddCookie(fakeWhiteCookie())

	err = handleReqCheckEventStream(gh, evpub, whiteRematchReq, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

  blackRematchReq, err := http.NewRequest("POST", "/rematch-offer", nil)
  if err != nil {
    t.Error(nil)
  }
  blackRematchReq.AddCookie(fakeBlackCookie())

	err = handleReqCheckEventStream(gh, evpub, blackRematchReq, http.StatusOK)
	if err != nil {
		t.Error(err)
	}
}

func TestPostRematchOfferNoCookie(t *testing.T) {
  evpub, chpub := GetTestPublishers()
	gh, _ := newGameHandler(
		nil, *chpub, game.Config{TimeControl: 1 * time.Minute}, fakeWhiteCookie(),
		fakeBlackCookie())

	req, err := http.NewRequest("POST", "/resignation", nil)
	req.AddCookie(fakeWhiteCookie())

	err = handleReqCheckEventStream(gh, evpub, req, http.StatusOK)
	if err != nil {
		t.Error(err)
	}

  whiteRematchReq, err := http.NewRequest("POST", "/rematch-offer", nil)
  if err != nil {
    t.Error(nil)
  }

	err =
    handleReqCheckEventStream(gh, evpub, whiteRematchReq, http.StatusBadRequest)
	if err == nil {
		t.Error("expected request to fail")
	}
}
