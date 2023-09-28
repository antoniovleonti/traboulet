package game

import (
	"net/http"
	"testing"
	"time"
)

func fakeWhiteCookie() *http.Cookie {
	c := http.Cookie{
		Name:  "white",
		Value: "1234",
		Path:  "/",
	}
	return &c
}

func fakeBlackCookie() *http.Cookie {
	c := http.Cookie{
		Name:  "black",
		Value: "5678",
		Path:  "/",
	}
	return &c
}

func TestNewGameManager(t *testing.T) {
	gm, err := NewGameManager(
		Config{TimeControl: time.Minute}, fakeWhiteCookie(), fakeBlackCookie(),
		nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(gm.cookieToUser) != 2 {
		t.Error("num users did not match expectation")
	}

	used := make(map[AgentColor]bool)
	for _, v := range gm.cookieToUser {
		used[v.color] = true
	}

	if !(used[agentWhite] && used[agentBlack]) {
		t.Error("did not assign players to both colors!")
	}
}

func TestTryMove(t *testing.T) {
	gm, err := NewGameManager(
		Config{TimeControl: time.Minute}, fakeWhiteCookie(), fakeBlackCookie(),
		nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		m     Move
		a     AgentColor
		valid bool
	}
	testCases := []testCase{
		testCase{m: Move{X: 0, Y: 0, D: DirDown}, a: agentWhite, valid: true},
		testCase{m: Move{X: 0, Y: 0, D: DirDown}, a: agentWhite, valid: false},
		testCase{m: Move{X: 0, Y: 1, D: DirDown}, a: agentWhite, valid: false},
		testCase{m: Move{X: 6, Y: 0, D: DirDown}, a: agentBlack, valid: true},
		testCase{m: Move{X: 0, Y: 1, D: DirDown}, a: agentWhite, valid: true},
	}
	for idx, tc := range testCases {
		actual := gm.TryMove(tc.m, gm.colorToUser[tc.a].cookie)
		if (actual == nil) != tc.valid {
			t.Errorf("testCases[%d]: expected %t, got %t\n", idx, tc.valid, actual)
		}
	}
}

func TestTryResign(t *testing.T) {
	gm, err := NewGameManager(
		Config{TimeControl: 600 * time.Second}, fakeWhiteCookie(),
		fakeBlackCookie(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if gm == nil {
		t.Error("manager is nil")
	}

	if !gm.TryResign(gm.GetWhiteCookie()) {
		t.Error("couldn't resign with white player")
	}
	if gm.state.status != statusBlackWon {
		t.Errorf("unexpected status; expected %d, got %d",
			statusBlackWon, gm.state.status)
	}

	if gm.TryResign(gm.GetBlackCookie()) {
		t.Error("Was able to resign with black after game was already over!")
	}
}

func TestRematchHappyPath(t *testing.T) {
  gameOverCbCalled := false
  gameOverCb := func() { gameOverCbCalled = true }
  rematchCbCalled := false
  rematchCb := func() { rematchCbCalled = true }

	gm, err := NewGameManager(
		Config{TimeControl: 600 * time.Second}, fakeWhiteCookie(),
		fakeBlackCookie(), nil, gameOverCb, rematchCb)

  success := gm.TryResign(fakeWhiteCookie())
  if !success {
    t.Error("expected to be able to resign with white")
  }

  if !gameOverCbCalled {
    t.Error("expected game over callback to be called")
  }

  // game is now over, try to rematch.
  rematchStarted, err := gm.OfferRematch(fakeWhiteCookie())
  if err != nil {
    t.Error(err)
  }
  if rematchStarted {
    t.Error("rematch started after only one player offered a rematch")
  }
  if !gm.colorToUser[agentWhite].wantsRematch {
    t.Error("expected white to want rematch")
  }
  // Try it again with the same cookie to make sure there's an error
  rematchStarted, err = gm.OfferRematch(fakeWhiteCookie())
  if err == nil {
    t.Error("no error on repeated rematch request")
  }

  rematchStarted, err = gm.OfferRematch(fakeBlackCookie())
  if err != nil {
    t.Error(err)
  }
  if !rematchStarted {
    t.Error("rematch did not start after both players offered rematch")
  }
  if !rematchCbCalled {
    t.Error("expected rematch callback to be called on successful rematch.")
  }

  // player colors get swapped
  if gm.cookieToUser[getKeyFromCookie(fakeWhiteCookie())].color != agentBlack {
    t.Error("expected previous white player to become black")
  }
  if gm.cookieToUser[getKeyFromCookie(fakeBlackCookie())].color != agentWhite {
    t.Error("expected previous black player to become white")
  }

  if gm.state.status != statusOngoing {
    t.Error(
      "expected (entire game to reset and as a result) status to reset to "+
      "ongoing")
  }

  if len(gm.state.history) != 1 {
    t.Error(
      "expected (entire game to reset and as a result) history length of "+
      "exactly 1")
  }

  for _, user := range gm.colorToUser {
    if user.wantsRematch {
      t.Error("expected user.wantsRematch to be reset to false")
    }
  }
}

func TestRematchOngoingGame(t *testing.T) {
	gm, err := NewGameManager(
		Config{TimeControl: 600 * time.Second}, fakeWhiteCookie(),
		fakeBlackCookie(), nil, nil, nil)

  _, err = gm.OfferRematch(fakeWhiteCookie())
  if err == nil {
    t.Error("expected error; rematch offered while game is still in play.")
  }
}
