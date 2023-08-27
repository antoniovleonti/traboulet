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
    fakeBlackCookie(), nil)
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
