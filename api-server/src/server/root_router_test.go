package server

import (
  "log"
	"bytes"
	"encoding/json"
	"game"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
  "evtpub"
)

func TestNewMatchmaker(t *testing.T) {
	var _ http.Handler = (*rootRouter)(nil)

	rtr := NewRootRouter(evtpub.NewMockEventPublisher())
	if rtr == nil {
		t.Error("rootRouter was nil")
	}
}

func TestServeHTTPSetsCookie(t *testing.T) {
	rtr := NewRootRouter(evtpub.NewMockEventPublisher())

	req, err := http.NewRequest("GET", "/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := httptest.NewRecorder()
	rtr.ServeHTTP(resp, req)

	if resp.Header().Get("Set-Cookie") == "" {
		t.Error("cookie was not set")
	}
}

// test fwdToChallengeRouter
func TestGetChallengesMapFromRoot(t *testing.T) {
	rtr := NewRootRouter(evtpub.NewMockEventPublisher())

	// test both with and without trailing slash
	for _, path := range []string{"/challenges", "/challenges/"} {
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp := httptest.NewRecorder()
		rtr.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("expected code %d, got %d", http.StatusOK, resp.Code)
		}
	}
}

func addAcceptedGame(rtr *rootRouter, t *testing.T) string {
	// add challenge
	b, err := json.Marshal(game.Config{TimeControl: 1 * time.Minute})
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
	rtr.ServeHTTP(postChallengeResp, postChallengeReq)

	// get challenges
	getChallengesReq, err := http.NewRequest("GET", "/challenges", nil)
	if err != nil {
		t.Error(err)
	}
	getChallengesResp := httptest.NewRecorder()
	rtr.ServeHTTP(getChallengesResp, getChallengesReq)
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
	rtr.ServeHTTP(joinChallengeResp, joinChallengeReq)

	// check response
	if joinChallengeResp.Code != http.StatusSeeOther {
		t.Errorf("expected status %d; got %d",
			http.StatusSeeOther, joinChallengeResp.Code)
	}
  log.Print(joinChallengeResp.Header().Get("Location"))
	return joinChallengeResp.Header().Get("Location")

}

// Test full user journey from challenge to playing moves in a game.
func TestFlowFromChallengeToPlay(t *testing.T) {
	rtr := NewRootRouter(evtpub.NewMockEventPublisher())

  gamePath := addAcceptedGame(rtr, t)

	// check existence of game
	getGameReq, err := http.NewRequest("GET", gamePath+"/state", nil)
	if err != nil {
		t.Error(err)
	}
	getGameResp := httptest.NewRecorder()
	rtr.ServeHTTP(getGameResp, getGameReq)

	if getGameResp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, getGameResp.Code)
	}

	// Make a move
	move := game.Move{X: 0, Y: 0, D: game.DirRight}
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
	postMoveReq.AddCookie(rtr.gameRtr.games[gameID].gm.GetWhiteCookie())
	postMoveResp := httptest.NewRecorder()
	rtr.ServeHTTP(postMoveResp, postMoveReq)

	if postMoveResp.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, postMoveResp.Code)
	}

	completionT := time.Now().Add(-1 * time.Hour)
	rtr.gameRtr.games[gameID].completionTime = &completionT
	rtr.gameRtr.deleteGamesOlderThan(1 * time.Minute)

	if len(rtr.challengeRtr.challenges) != 0 {
		t.Error("expected challenge to be deleted")
	}
}

func TestDeleteOldChallengesAndDeleteChallengeCbRace(t *testing.T) {
	rtr := NewRootRouter(evtpub.NewMockEventPublisher())
  path := addAcceptedGame(rtr, t)
	gameID := strings.Split(path, "/")[2]

  game, ok := rtr.gameRtr.games[gameID]
  if !ok {
    t.Error("game doesn't exist")
  }
  go rtr.challengeRtr.deleteOldChallenges(10 * time.Minute)
  game.deleteChallengeCb()
}
