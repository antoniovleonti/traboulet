package server

import (
	"bytes"
	"encoding/json"
	"game"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
  "evtpub"
  "net/url"
)


func TestNewChallengeRouter(t *testing.T) {
	var _ http.Handler = (*challengeRouter)(nil)

  evpub := evtpub.NewMockEventPublisher()
  urlBase, _ := url.Parse("/")

	cr := newChallengeRouter(urlBase, nil, evpub)
	if cr == nil {
		t.Error("matchmaker was nil")
	}
}

func TestPostChallenge(t *testing.T) {
  evpub := evtpub.NewMockEventPublisher()
  urlBase, _ := url.Parse("/")
	cr := newChallengeRouter(urlBase, nil, evpub)

	rr, err := post10MinChallenge(cr)
  if err != nil {
    t.Error(err)
  }

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
  evpub, chpub := GetTestPublishers()
  urlBase, _ := url.Parse("/")
  cr := newChallengeRouter(urlBase, nil, evpub)


	challenge, err := newChallengeHandler(
    fakeWhiteCookie(), game.Config{time.Minute}, nil, chpub)
	if err != nil {
		t.Error(err)
	}
	cr.challenges["testpath"] = challenge

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
  evpub, chpub := GetTestPublishers()
  urlBase, _ := url.Parse("/")
  cr := newChallengeRouter(urlBase, nil, evpub)

	type challengeParams struct {
		s      string
		cookie *http.Cookie
		config game.Config
	}

	newChallengeParams := func(
		s string, cookie *http.Cookie, config game.Config) *challengeParams {
		cp := challengeParams{
			s:      s,
			cookie: cookie,
			config: config,
		}
		return &cp
	}

	paramsList := []*challengeParams{
		newChallengeParams("a", fakeWhiteCookie(), game.Config{time.Minute}),
		newChallengeParams(
			"b", fakeWhiteCookie(), game.Config{time.Minute}),
		newChallengeParams(
			"c", fakeWhiteCookie(), game.Config{time.Hour}),
	}

	for _, params := range paramsList {
		challenge, err := newChallengeHandler(
      params.cookie, params.config, nil, chpub)
		if err != nil {
			t.Fatal(err)
		}
		cr.challenges[params.s] = challenge
	}
	// Add an accepted challenge
	cr.challenges["accepted"] = &challengeHandler{accepted: true}

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
		if k == "accepted" {
			if _, ok := result["accepted"]; ok {
				t.Error("expected accepted challenge to not be sent.")
			}
		} else {
			actualv, ok := result[k]
			if !ok {
				t.Error("expected key " + k + " to exist in result")
			}
			if ok && actualv.Config.TimeControl != v.config.TimeControl {
				t.Errorf(
					"time %s does not match expected %s",
					actualv.Config.TimeControl.String(), v.config.TimeControl.String())
			}
		}
	}
}

// Really for safety against crashing due to dereferencing nil map entries etc
func TestJoinNonExistentID(t *testing.T) {
	evpub := evtpub.NewMockEventPublisher()
  urlBase, _ := url.Parse("/")
  cr := newChallengeRouter(urlBase, nil, evpub)

	// This should trigger the above callback
	req, err := http.NewRequest("POST", "/nonexistent/join", nil)
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

func TestPostChallengeInvalidConfig(t *testing.T) {
	evpub := evtpub.NewMockEventPublisher()
  urlBase, _ := url.Parse("/")
  cr := newChallengeRouter(urlBase, nil, evpub)

	// create body
	config := game.Config{}
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
	if rr.Code != http.StatusBadRequest {
		t.Errorf(
			"code %d does not match expectation %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeleteOldChallenges(t *testing.T) {
	evpub := evtpub.NewMockEventPublisher()
  urlBase, _ := url.Parse("/")
  cr := newChallengeRouter(urlBase, nil, evpub)

  chpub1, _ := evpub.NewChannelPublisher("testpath1")
	cr.challenges["too old"] = &challengeHandler{
		timestamp: time.Now().Add(-1 * time.Hour),
		accepted:  false,
    channelPub: chpub1,
	}
  chpub2, _ := evpub.NewChannelPublisher("testpath2")
	cr.challenges["accepted"] = &challengeHandler{
		timestamp: time.Now().Add(-1 * time.Hour),
		accepted:  true,
    channelPub: chpub2,
	}
  chpub3, _ := evpub.NewChannelPublisher("testpath3")
	cr.challenges["not too old"] = &challengeHandler{
		timestamp: time.Now().Add(-1 * time.Minute),
		accepted:  false,
    channelPub: chpub3,
	}

	cr.deleteOldChallenges(10 * time.Minute)

	if len(cr.challenges) != 2 {
		t.Errorf("expected exactly 2 games (got %d)", len(cr.challenges))
	}

	if _, ok := cr.challenges["not too old"]; !ok {
		t.Error("expected newer challenge to stay")
	}
	if _, ok := cr.challenges["accepted"]; !ok {
		t.Error("expected accepted challenge to stay")
	}
	if _, ok := cr.challenges["too old"]; ok {
		t.Error("expected too old challenge to be deleted")
	}
}

func post10MinChallenge(
  cr *challengeRouter) (*httptest.ResponseRecorder, error) {
	config := game.Config{10 * time.Minute}
	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
  req, err := http.NewRequest("POST", "/", bytes.NewReader(b))
  if err != nil {
    return nil, err
  }
  req.AddCookie(fakeWhiteCookie())
  // handle request
  rr := httptest.NewRecorder()
  cr.ServeHTTP(rr, req)

  return rr, nil
}

func TestChallengeCountLimit(t *testing.T) {
	evpub := evtpub.NewMockEventPublisher()
  urlBase, _ := url.Parse("/")
  cr := newChallengeRouter(urlBase, nil, evpub)

	for i := 0; i < 100; i++ {
    rr, err := post10MinChallenge(cr)
    if err != nil {
      t.Error(err)
    }
    // check status
    if rr.Code != http.StatusSeeOther {
      t.Errorf(
        "code %d does not match expectation %d; body: %s", rr.Code,
        http.StatusSeeOther, rr.Body.String())
    }
  }

	if len(cr.challenges) != 100 {
		t.Errorf("expected exactly 100 challenges (got %d)", len(cr.challenges))
	}

  rr, err := post10MinChallenge(cr)
  if err != nil {
    t.Error(err)
  }
  // check status
  if rr.Code != http.StatusInternalServerError {
    t.Errorf(
      "code %d does not match expectation %d", rr.Code,
      http.StatusInternalServerError)
  }
}
