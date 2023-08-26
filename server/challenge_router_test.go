package server

import (
	"bytes"
	"encoding/json"
	"kuba"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewChallengeRouter(t *testing.T) {
	var _ http.Handler = (*challengeRouter)(nil)

	cr := newChallengeRouter("/", nil)
	if cr == nil {
		t.Error("matchmaker was nil")
	}
}

func TestPostChallenge(t *testing.T) {
	cr := newChallengeRouter("/", nil)

	// create body
	config := kuba.Config{10 * time.Minute}
	b, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	// make request
	req, err := http.NewRequest("POST", "/", bytes.NewReader(b))
	req.AddCookie(fakeWhiteCookie())

	// handle request
	rr := httptest.NewRecorder()
	cr.ServeHTTP(rr, req)

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
	path := rr.Header().Get("Location")[1:] // remove leading "/"
	if ch, ok := cr.challenges[path]; !ok || ch == nil {
		t.Error("challenge not set / path not correct")
	}
}

func TestHandleChallengeRequestForwarding(t *testing.T) {
	cr := newChallengeRouter("/", nil)

	cr.challenges["testpath"] =
		newChallengeHandler(fakeWhiteCookie(), kuba.Config{}, nil)

	req, err :=
		http.NewRequest("GET", "/testpath/", nil)
	if err != nil {
		t.Error(err)
	}
	req.AddCookie(fakeBlackCookie())

	// handle request
	rr := httptest.NewRecorder()
	cr.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestGetChallenges(t *testing.T) {
	cr := newChallengeRouter("/", nil)

	cr.challenges["a"] =
		newChallengeHandler(fakeWhiteCookie(), kuba.Config{}, nil)
	cr.challenges["b"] = newChallengeHandler(
		fakeWhiteCookie(), kuba.Config{InitialTime: 1 * time.Minute}, nil)
	cr.challenges["c"] = newChallengeHandler(
		fakeWhiteCookie(), kuba.Config{InitialTime: 1 * time.Hour}, nil)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}
	rr := httptest.NewRecorder()

	cr.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(
			"unexpected response status %d; expected %d",
			rr.Code, http.StatusOK)
	}

	var result map[string]challengeHandlerView
	err = json.NewDecoder(rr.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	for k, v := range cr.challenges {
		actualv, ok := result[k]
		if !ok {
			t.Error("expected key " + k + " to exist in result")
		}
		if ok && actualv.Config.InitialTime != v.config.InitialTime {
			t.Errorf(
				"time %s does not match expected %s",
				actualv.Config.InitialTime.String(), v.config.InitialTime.String())
		}
	}
}

// Really for safety against crashing due to dereferencing nil map entries etc
func TestJoinNonExistentID(t *testing.T) {
	cr := newChallengeRouter("/", nil)

	// This should trigger the above callback
	req, err :=
		http.NewRequest("POST", "/nonexistent/join", nil)
	if err != nil {
		t.Error(err)
	}
	req.AddCookie(fakeBlackCookie())

	// handle request
	rr := httptest.NewRecorder()
	cr.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestDeleteChallengesOlderThan(t *testing.T) {
  cr := newChallengeRouter("/", nil)
  cr.challenges["too old"] = &challengeHandler{
    timestamp: time.Now().Add(-1 * time.Hour),
  }
  cr.challenges["not too old"] = &challengeHandler{
    timestamp: time.Now().Add(-1 * time.Minute),
  }

  cr.deleteChallengesOlderThan(10 * time.Minute)

  if len(cr.challenges) != 1 {
    t.Errorf("expected exactly 1 game (got %d)", len(cr.challenges))
  }

  if _, ok := cr.challenges["not too old"]; !ok {
    t.Error("expected newer challenge to stay")
  }
}