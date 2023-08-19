package server

import (
	"bytes"
	"encoding/json"
	"kuba"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewMatchmaker(t *testing.T) {
	var _ http.Handler = (*matchmaker)(nil)

	mm := newMatchmaker()
	if mm == nil {
		t.Error("matchmaker was nil")
	}
}

func TestNewPath(t *testing.T) {
	pg := newPathGenerator()

	p := pg.newPath(8)
	if len(p) != 8 {
		t.Errorf("p had unexpected len %d; expected %d", len(p), 8)
	}
}

func TestServeHTTPSetsCookie(t *testing.T) {
	mm := newMatchmaker()

	req, err := http.NewRequest("GET", "/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	mm.ServeHTTP(rr, req)

	if rr.Header().Get("Set-Cookie") == "" {
		t.Error("cookie was not set")
	}
}

func TestMoveChallengeToGame(t *testing.T) {
	mm := newMatchmaker()

	mm.challenges["testpath"] = newChallengeHandler(
		fakeWhiteCookie(), kuba.Config{}, nil)

	rr := httptest.NewRecorder()
	mm.challenges["testpath"].pub.addSubscriber(rr)

	mm.moveChallengeToGame(
		"testpath", &mm.challenges["testpath"].pub, kuba.Config{},
		fakeWhiteCookie(), fakeBlackCookie())

	if rr.Code != http.StatusResetContent {
		t.Errorf(
			"unexpected subscriber status %d; expected %d",
			rr.Code, http.StatusResetContent)
	}

	if _, ok := mm.challenges["testpath"]; ok {
		t.Error("expected challenge to be deleted")
	}

	if game, ok := mm.games["testpath"]; !ok || game == nil {
		t.Error("expected game to be created")
	}
}

func TestPostGame(t *testing.T) {
	mm := newMatchmaker()

	// create body
	config := kuba.Config{10 * time.Minute}
	b, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	// make request
	req, err := http.NewRequest("POST", "/games", bytes.NewReader(b))
	req.AddCookie(fakeWhiteCookie())

	// handle request
	rr := httptest.NewRecorder()
	mm.ServeHTTP(rr, req)

	// check status
	if rr.Code != http.StatusSeeOther {
		t.Errorf(
			"code %d does not match expectation %d", rr.Code, http.StatusSeeOther)
	}
	// parse location
	if rr.Header().Get("Location") == "" {
		t.Error("expected Location header field to be set")
	}
	// check challenge
	path := strings.Split(rr.Header().Get("Location"), "/")[2]
	if ch, ok := mm.challenges[path]; !ok || ch == nil {
		t.Error("challenge not set / path not correct")
	}
}

func TestHandleChallengeRequestForwarding(t *testing.T) {
	mm := newMatchmaker()

	done := false
	cb := func(*publisher, kuba.Config, *http.Cookie, *http.Cookie) {
		done = true
	}

	mm.challenges["testpath"] =
		newChallengeHandler(fakeWhiteCookie(), kuba.Config{}, cb)

	// This should trigger the above callback
	req, err :=
		http.NewRequest("POST", "/games/testpath/join", nil)
	if err != nil {
		t.Error(err)
	}
	req.AddCookie(fakeBlackCookie())

	// handle request
	rr := httptest.NewRecorder()
	mm.ServeHTTP(rr, req)

	if !done {
		t.Error("callback was not called")
	}
}

func TestHandleGameRequestForwarding(t *testing.T) {
	mm := newMatchmaker()

	done := false
	cb := func(*publisher, kuba.Config, *http.Cookie, *http.Cookie) {
		done = true
	}

	mm.challenges["testpath"] =
		newChallengeHandler(fakeWhiteCookie(), kuba.Config{}, cb)

	// This should trigger the above callback
	req, err :=
		http.NewRequest("POST", "/games/testpath/join", nil)
	if err != nil {
		t.Error(err)
	}
	req.AddCookie(fakeBlackCookie())

	// handle request
	rr := httptest.NewRecorder()
	mm.ServeHTTP(rr, req)

	if !done {
		t.Error("callback was not called")
	}
}
