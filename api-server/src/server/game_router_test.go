package server

import (
	"game"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
  "net/url"
  "evtpub"
)

func TestNewGameRouter(t *testing.T) {
	var _ http.Handler = (*gameRouter)(nil)

  urlBase, _ := url.Parse("/")
	gr := newGameRouter(urlBase, evtpub.NewMockEventPublisher())
	if gr == nil {
		t.Error("gameRouter was nil")
	}
}

func TestAddGame(t *testing.T) {
  urlBase, _ := url.Parse("/")
	gr := newGameRouter(urlBase, evtpub.NewMockEventPublisher())

	_, err := gr.addGame(
		func() {}, game.Config{TimeControl: 1 * time.Minute},
		fakeWhiteCookie(), fakeBlackCookie())
	if err != nil {
		t.Error(err)
	}

	if len(gr.games) != 1 {
		t.Error("expected number of games to equal 1")
	}
}

func makeRouterWithTestGame() (*gameRouter, error) {
  urlBase, _ := url.Parse("/")
  evpub, chpub := GetTestPublishers()
	gr := newGameRouter(urlBase, evpub)

	game, err := newGameHandler(
		func() {}, *chpub, game.Config{TimeControl: 1 * time.Minute},
    fakeWhiteCookie(), fakeBlackCookie())
	if err != nil {
		return nil, err
	}
	gr.games["testpath"] = game
	return gr, nil
}

func TestHandleGameRequestForwarding(t *testing.T) {
	gr, err := makeRouterWithTestGame()
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
  evpub := evtpub.NewMockEventPublisher()
  urlBase, _ := url.Parse("/")
	gr := newGameRouter(urlBase, evpub)

	tTooOld := time.Now().Add(-1 * time.Hour)
  chpubTooOld, _ := evpub.NewChannelPublisher("too/old")
	gr.games["too old"] = &gameHandler{
		completionTime: &tTooOld,
    channelPub: *chpubTooOld,
	}
	tNotTooOld := time.Now().Add(-1 * time.Minute)
  chpubNotTooOld, _ := evpub.NewChannelPublisher("not/too/old")
	gr.games["not too old"] = &gameHandler{
		completionTime: &tNotTooOld,
    channelPub: *chpubNotTooOld,
	}
  chpubIncomplete, _ := evpub.NewChannelPublisher("incomplete")
	gr.games["incomplete"] = &gameHandler{
		completionTime: nil,
    channelPub: *chpubIncomplete,
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

func TestLongLivedRequestDoesntBlockMutex(t *testing.T) {
	gr, err := makeRouterWithTestGame()
	if err != nil {
		t.Error(err)
	}

	// Idc about actually receiving anything. It's just important that the request
	// is being serviced.
	req, err := http.NewRequest("GET", "/testpath/event-stream", nil)
	if err != nil {
		t.Error(err)
	}

	// handle request
	rr := httptest.NewRecorder()
	go gr.ServeHTTP(rr, req)

	time.Sleep(5 * time.Millisecond)

	if !gr.mutex.TryLock() {
		t.Error("could not aquire mutex")
	}
}

func TestGameNumberLimit(t *testing.T) {
  urlBase, _ := url.Parse("/")
	gr := newGameRouter(urlBase, evtpub.NewMockEventPublisher())

  for i := 0; i < 100; i++ {
    _, err := gr.addGame(
      func() {}, game.Config{TimeControl: 1 * time.Minute},
      fakeWhiteCookie(), fakeBlackCookie())
    if err != nil {
      t.Error(err)
    }
  }

  empty, err := gr.addGame(
    func() {}, game.Config{TimeControl: 1 * time.Minute},
    fakeWhiteCookie(), fakeBlackCookie())
  if err == nil {
    t.Error("expected 101st game add to fail")
  }
  if empty != nil {
    t.Error("expected 101st game to be nil")
  }

	if len(gr.games) != 100 {
		t.Error("expected number of games to equal 100")
	}
}
