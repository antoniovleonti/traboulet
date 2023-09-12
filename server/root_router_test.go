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
	var _ http.Handler = (*rootRouter)(nil)

	rr := NewRootRouter()
	if rr == nil {
		t.Error("rootRouter was nil")
	}
}

func TestServeHTTPSetsCookie(t *testing.T) {
	rr := NewRootRouter()

	req, err := http.NewRequest("GET", "/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	rr.ServeHTTP(resp, req)

	if resp.Header().Get("Set-Cookie") == "" {
		t.Error("cookie was not set")
	}
}

// test fwdToChallengeRouter
func TestGetChallengesMapFromRoot(t *testing.T) {
	rr := NewRootRouter()

	// test both with and without trailing slash
	for _, path := range []string{"/challenges", "/challenges/"} {
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp := httptest.NewRecorder()
		rr.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("expected code %d, got %d", http.StatusOK, resp.Code)
		}
	}
}

// Test full user journey from challenge to playing moves in a game.
func TestFlowFromChallengeToPlay(t *testing.T) {
	rr := NewRootRouter()

	// add challenge
	b, err := json.Marshal(kuba.Config{TimeControl: 1 * time.Minute})
	if err != nil {
		t.Error(err)
	}
	postChallengeReq, err :=
		http.NewRequest("POST", "/challenges", bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}
	postChallengeReq.AddCookie(fakeWhiteCookie())
	postChallengeResp := httptest.NewRecorder()
	rr.ServeHTTP(postChallengeResp, postChallengeReq)

	// get challenges
	getChallengesReq, err := http.NewRequest("GET", "/challenges", nil)
	if err != nil {
		t.Error(err)
	}
	getChallengesResp := httptest.NewRecorder()
	rr.ServeHTTP(getChallengesResp, getChallengesReq)
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
		http.NewRequest("POST", "/challenges/"+challengeID+"/accept", nil)
	if err != nil {
		t.Error(err)
	}
	joinChallengeReq.AddCookie(fakeBlackCookie())
	joinChallengeResp := httptest.NewRecorder()
	rr.ServeHTTP(joinChallengeResp, joinChallengeReq)

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
	rr.ServeHTTP(getGameResp, getGameReq)

	if getGameResp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, getGameResp.Code)
	}

	// Make a move
	move := kuba.Move{X: 0, Y: 0, D: kuba.DirRight}
	bmove, err := json.Marshal(move)
	if err != nil {
		t.Fatal(err)
	}
	postMoveReq, err :=
		http.NewRequest("POST", gamePath+"/move", bytes.NewReader(bmove))
	if err != nil {
		t.Error(err)
	}
	gameID := strings.Split(gamePath, "/")[2]
	postMoveReq.AddCookie(rr.gameRtr.games[gameID].km.GetWhiteCookie())
	postMoveResp := httptest.NewRecorder()
	rr.ServeHTTP(postMoveResp, postMoveReq)

	if postMoveResp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, postMoveResp.Code)
	}
}
