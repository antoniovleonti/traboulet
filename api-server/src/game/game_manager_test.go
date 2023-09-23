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
		nil, nil)
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
		nil, nil)
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
		fakeBlackCookie(), nil, nil)
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
