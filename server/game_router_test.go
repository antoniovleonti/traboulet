package server

import (
	"kuba"
	"net/http"
	"net/http/httptest"
	"testing"
  "time"
)

func TestNewGameRouter(t *testing.T) {
	var _ http.Handler = (*gameRouter)(nil)

	gr := newGameRouter("/")
	if gr == nil {
		t.Error("gameRouter was nil")
	}
}

func TestAddGame(t *testing.T) {
	gr := newGameRouter("/")

	_, err := gr.addGame(
    kuba.Config{TimeControl: 1*time.Minute}, fakeWhiteCookie(),
    fakeBlackCookie())
  if err != nil {
    t.Error(err)
  }

	if len(gr.games) != 1 {
		t.Error("expected number of games to equal 1")
	}
}

func TestHandleGameRequestForwarding(t *testing.T) {
	gr := newGameRouter("/")

	game, err := newGameHandler(
    kuba.Config{TimeControl: 1*time.Minute}, fakeWhiteCookie(),
    fakeBlackCookie())
  if err != nil {
    t.Fatal(err)
  }
  gr.games["testpath"] = game

	// This should trigger the above callback
	req, err := http.NewRequest("GET", "/testpath/state", nil)
	if err != nil {
		t.Error(err)
	}

	// handle request
	rr := httptest.NewRecorder()
	gr.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestDeleteGamesOlderThan(t *testing.T) {
	gr := newGameRouter("/")
  tTooOld := time.Now().Add(-1 * time.Hour)
	gr.games["too old"] = &gameHandler{
		completionTime: &tTooOld,
	}
  tNotTooOld := time.Now().Add(-1 * time.Minute)
	gr.games["not too old"] = &gameHandler{
		completionTime: &tNotTooOld,
	}
	gr.games["incomplete"] = &gameHandler{
		completionTime: nil,
	}

	gr.deleteGamesOlderThan(10 * time.Minute)

	if len(gr.games) != 2 {
		t.Errorf("expected exactly 2 game (got %d)", len(gr.games))
	}

	if _, ok := gr.games["not too old"]; !ok {
		t.Error("expected newer game to stay")
	}
	if _, ok := gr.games["incomplete"]; !ok {
		t.Error("expected incomplete game to stay")
	}
}
