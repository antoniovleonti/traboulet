package server

import (
	"kuba"
	"net/http"
	"net/http/httptest"
	"testing"
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

	gr.addGame(kuba.Config{}, fakeWhiteCookie(), fakeBlackCookie())

	if len(gr.games) != 1 {
		t.Error("expected number of games to equal 1")
	}
}

func TestHandleGameRequestForwarding(t *testing.T) {
	gr := newGameRouter("/")

	gr.games["testpath"] =
		newGameHandler(kuba.Config{}, fakeWhiteCookie(), fakeBlackCookie(), nil)

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
