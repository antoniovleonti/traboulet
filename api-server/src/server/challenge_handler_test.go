package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"game"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewChallengeHandler(t *testing.T) {
	var _ http.Handler = (*challengeHandler)(nil)

	cb := func(
		*challengeHandler, game.Config, *http.Cookie, *http.Cookie) (
		string, error) {
		return "", nil
	}

	ch, err :=
		newChallengeHandler(fakeWhiteCookie(), game.Config{time.Minute}, cb)
	if err != nil {
		t.Error(err)
	}
	if ch == nil {
		t.Error("nil challenge handler")
	}
}

func TestPostJoinValid(t *testing.T) {
	white := fakeWhiteCookie()
	black := fakeBlackCookie()

	callbackCalled := false
	cb := func(
		ch *challengeHandler, c game.Config, w *http.Cookie, b *http.Cookie) (
		string, error) {
		if w.Value != white.Value {
			t.Error("white cookie did not match expectation")
		}
		if b.Value != black.Value {
			t.Error("black cookie did not match expectation")
		}
		callbackCalled = true
		return "", nil
	}

	ch, err := newChallengeHandler(white, game.Config{time.Minute}, cb)
	if err != nil {
		t.Error(err)
	}
	if ch == nil {
		t.Error("nil challenge handler")
	}

	// Build request
	postJoinReq, err := http.NewRequest("POST", "/accept", nil)
	if err != nil {
		t.Fatal(err)
	}
	postJoinReq.AddCookie(black)

	// Add a subscriber to game updates
	getUpdateReq, err := http.NewRequest("GET", "/update", nil)
	if err != nil {
		t.Fatal(err)
	}
	getUpdateResp := httptest.NewRecorder()
	go ch.ServeHTTP(getUpdateResp, getUpdateReq)
	time.Sleep(time.Millisecond)

	// Run the request through our handler
	postJoinResp := httptest.NewRecorder()
	ch.ServeHTTP(postJoinResp, postJoinReq)

	if !callbackCalled {
		t.Error("callback was never called")
	}

	if getUpdateResp.Code != http.StatusSeeOther {
		t.Errorf("expected code %d, got %d",
			http.StatusSeeOther, getUpdateResp.Code)
	}
}

func TestPostJoinNoCookie(t *testing.T) {
	white := fakeWhiteCookie()

	callbackCalled := false
	cb := func(
		*challengeHandler, game.Config, *http.Cookie, *http.Cookie) (
		string, error) {
		callbackCalled = true
		return "", nil
	}

	ch, err := newChallengeHandler(white, game.Config{time.Minute}, cb)
	if err != nil {
		t.Error(err)
	}
	if ch == nil {
		t.Error("nil challenge handler")
	}

	// Build request
	postJoinReq, err := http.NewRequest("POST", "/accept", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Run the request through our handler
	postJoinResp := httptest.NewRecorder()
	ch.ServeHTTP(postJoinResp, postJoinReq)

	if callbackCalled {
		t.Error("callback was called when it shouldn't have been")
	}

	if postJoinResp.Code != http.StatusUnauthorized {
		t.Errorf("update resp code (%d) did not match expectation (%d)",
			postJoinResp.Code, http.StatusUnauthorized)
	}
}

func TestPostJoinSameCookie(t *testing.T) {
	white := fakeWhiteCookie()

	callbackCalled := false
	cb := func(
		*challengeHandler, game.Config, *http.Cookie, *http.Cookie) (
		string, error) {
		callbackCalled = true
		return "", nil
	}

	ch, err := newChallengeHandler(white, game.Config{time.Minute}, cb)
	if err != nil {
		t.Error(err)
	}
	if ch == nil {
		t.Error("nil challenge handler")
	}

	// Build request
	postJoinReq, err := http.NewRequest("POST", "/accept", nil)
	if err != nil {
		t.Fatal(err)
	}
	postJoinReq.AddCookie(white)

	// Run the request through our handler
	postJoinResp := httptest.NewRecorder()
	ch.ServeHTTP(postJoinResp, postJoinReq)

	if callbackCalled {
		t.Error("callback was called when it shouldn't have been")
	}

	if postJoinResp.Code != http.StatusBadRequest {
		t.Errorf("update resp code (%d) did not match expectation (%d)",
			postJoinResp.Code, http.StatusBadRequest)
	}
}

func getAcceptedChallenge() (*challengeHandler, error) {
	white := fakeWhiteCookie()
	black := fakeBlackCookie()

	callbackCalled := false
	cb := func(
		*challengeHandler, game.Config, *http.Cookie, *http.Cookie) (
		string, error) {
		callbackCalled = true
		return "", nil
	}

	ch, err := newChallengeHandler(white, game.Config{time.Minute}, cb)
	if err != nil {
		return nil, err
	}
	if ch == nil {
		errors.New("nil challenge handler")
	}

	// Build request
	postJoinReq, err := http.NewRequest("POST", "/accept", nil)
	if err != nil {
		return nil, err
	}
	postJoinReq.AddCookie(black)

	// Run the request through our handler
	postJoinResp := httptest.NewRecorder()
	ch.ServeHTTP(postJoinResp, postJoinReq)

	if !callbackCalled {
		return nil, errors.New("callback was never called")
	}

	return ch, nil
}

func TestPostJoinAgain(t *testing.T) {
	ch, err := getAcceptedChallenge()
	if err != nil {
		t.Error(err)
	}

	// just do it again
	postJoinReq, err := http.NewRequest("POST", "/accept", nil)
	if err != nil {
		t.Error(err)
	}
	postJoinReq.AddCookie(fakeBlackCookie())
	postJoinResp := httptest.NewRecorder()
	ch.ServeHTTP(postJoinResp, postJoinReq)

	if postJoinResp.Code != http.StatusSeeOther {
		t.Errorf("update resp code (%d) did not match expectation (%d)",
			postJoinResp.Code, http.StatusInternalServerError)
	}
}

func TestGetStateAfterAccepted(t *testing.T) {
	ch, err := getAcceptedChallenge()
	if err != nil {
		t.Error(err)
	}

	// just do it again
	getStateReq, err := http.NewRequest("GET", "/state", nil)
	if err != nil {
		t.Error(err)
	}
	getStateResp := httptest.NewRecorder()
	ch.ServeHTTP(getStateResp, getStateReq)

	if getStateResp.Code != http.StatusSeeOther {
		t.Errorf("update resp code (%d) did not match expectation (%d)",
			getStateResp.Code, http.StatusInternalServerError)
	}
}

func TestIntegrationWithGameRouter(t *testing.T) {
	// create game router
	gr := newGameRouter("/")
	// create challenge router (with game router fn as callback)
	cr := newChallengeRouter("/", gr.addGame)

	// add challenge
	b, err := json.Marshal(game.Config{TimeControl: 1 * time.Minute})
	if err != nil {
		t.Error(err)
	}
	postChallengeReq, err := http.NewRequest("POST", "/", bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}
	postChallengeReq.AddCookie(fakeWhiteCookie())
	postChallengeResp := httptest.NewRecorder()
	cr.ServeHTTP(postChallengeResp, postChallengeReq)

	// get challenges
	getChallengesReq, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}
	getChallengesResp := httptest.NewRecorder()
	cr.ServeHTTP(getChallengesResp, getChallengesReq)
	// decode
	var challenges map[string]challengeHandlerView
	err = json.NewDecoder(getChallengesResp.Body).Decode(&challenges)
	if err != nil {
		t.Error(err)
	}
	if len(challenges) != 1 {
		t.Error("expected exactly 1 challenge")
	}
	var challengeID string
	for k, _ := range challenges {
		challengeID = k
	}

	// accept challenge
	joinChallengeReq, err :=
		http.NewRequest("POST", "/"+challengeID+"/accept", nil)
	if err != nil {
		t.Error(err)
	}
	joinChallengeReq.AddCookie(fakeBlackCookie())
	joinChallengeResp := httptest.NewRecorder()
	cr.ServeHTTP(joinChallengeResp, joinChallengeReq)

	// check response
	if joinChallengeResp.Code != http.StatusSeeOther {
		t.Errorf("expected status %d; got %d",
			http.StatusSeeOther, joinChallengeResp.Code)
	}
	gamePath := joinChallengeResp.Header().Get("Location")

	// check existence of game
	getGameReq, err := http.NewRequest("GET", gamePath+"/state", nil)
	if err != nil {
		t.Error(err)
	}
	getGameResp := httptest.NewRecorder()
	gr.ServeHTTP(getGameResp, getGameReq)

	if getGameResp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, getGameResp.Code)
	}
}

func TestInvalidConfig(t *testing.T) {
	ch, err :=
		newChallengeHandler(fakeWhiteCookie(), game.Config{TimeControl: 0}, nil)
	if err == nil {
		t.Error("shouldn't be able to make a game with zero time")
	}
	if ch != nil {
		t.Errorf("challenge handler should be nil")
	}
}