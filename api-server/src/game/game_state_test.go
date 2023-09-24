package game

import (
	"testing"
	"time"
	// "fmt"
)

func makeSingleSnapshotHistory(board BoardT, agent AgentColor) []snapshot {
  return []snapshot{
    snapshot{
      board: board,
      whoseTurn: agent,
    },
  }
}

func TestCreateDefaultGameState(t *testing.T) {
	gsWClock, err := newGameState(
		Config{TimeControl: 60 * time.Second}, nil, nil, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if gsWClock == nil {
		t.Error("var gs should not be nil")
	}
}

func TestIsInBounds(t *testing.T) {
	gs, err := newGameState(
		Config{TimeControl: time.Minute}, nil, nil, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	inBoundsCases := [][2]int{
		{0, 0},
		{6, 6},
		{1, 1},
	}

	outOfBoundsCases := [][2]int{
		{-1, 0},
		{7, 7},
		{0, 7},
		{-1, 7},
	}

	for idx, testCase := range inBoundsCases {
		if !gs.isInBounds(testCase[0], testCase[1]) {
			t.Errorf("inBoundsCases[%d] (%d, %d) is out of bounds.",
				idx, testCase[0], testCase[1])
		}
	}

	for idx, testCase := range outOfBoundsCases {
		if gs.isInBounds(testCase[0], testCase[1]) {
			t.Errorf("outOfBoundsCases[%d] (%d, %d) is in bounds.",
				idx, testCase[0], testCase[1])
		}
	}
}

func TestValidateMove(t *testing.T) {
	// shorter names
	var B, W, R, x Marble = marbleBlack, marbleWhite, marbleRed, marbleNil

	gs := gameState{
		history: makeSingleSnapshotHistory(
      [][]Marble{
        {W, W, W, W, W, W, W},
        {W, W, x, W, W, W, W},
        {W, W, W, W, W, W, B},
        {W, W, W, W, W, W, R},
        {W, B, W, W, W, W, W},
        {B, B, W, W, W, W, x},
        {W, x, x, x, x, B, B},
      },
      agentWhite),
		winThreshold: 7,
		posToCount:   make(map[string]int),
	}

	validCases := []Move{
		Move{Y: 1, X: 0, D: DirRight},
		Move{Y: 2, X: 0, D: DirRight},
		Move{Y: 3, X: 0, D: DirRight},
		Move{Y: 3, X: 0, D: DirRight},
		Move{Y: 6, X: 0, D: DirRight},
	}

	invalidCases := []Move{
		Move{Y: 0, X: 0, D: DirRight},     // Kills own marble
		Move{Y: 5, X: 0, D: DirRight},     // Wrong color marble
		Move{Y: 1, X: 1, D: DirRight},     // Blocked from behind
		Move{Y: 6, X: 0, D: DirLeft},      // Kills own marble
		Move{Y: 0, X: 0, D: DirDown},      // Kills own marble
		Move{Y: 0, X: 0, D: DirNil},       // Nonsense dx, dy
		Move{Y: 0, X: 0, D: Direction(8)}, // Nonsense dx, dy
	}

	for idx, valid := range validCases {
    mwmm, err := gs.ValidateMove(valid)
		if err != nil {
			t.Errorf("validCases[%d] was considered invalid: %s", idx, err.Error())
		}
    if mwmm == nil {
      t.Errorf("mwmm was nil")
    }
	}
	for idx, invalid := range invalidCases {
    mwmm, err := gs.ValidateMove(invalid)
		if err == nil {
			t.Errorf("invalidCases[%d] was considered valid", idx)
		}
    if mwmm != nil {
      t.Errorf("mwmm was not nil")
    }
	}
}

func TestCantPushOffEdge(t *testing.T) {
	// shorter names
	var B, W, R, x Marble = marbleBlack, marbleWhite, marbleRed, marbleNil

	gs := gameState{
    history: []snapshot{
      snapshot{
        board: [][]Marble{
          {x, W, x, B, x, x, x},
          {x, x, R, x, W, x, B},
          {x, W, x, x, B, R, R},
          {R, x, x, x, x, x, x},
          {x, x, R, x, R, x, x},
          {W, x, x, x, R, x, x},
          {B, B, x, x, x, W, W},
        },
        whoseTurn: agentBlack,
      },
    },
		winThreshold: 7,
		posToCount:   make(map[string]int),
	}

	if _, err := gs.ValidateMove(Move{X: 1, Y: 6, D: DirLeft}); err == nil {
		t.Error("Could push own marble off")
	}
}

func TestExecuteMove(t *testing.T) {
	gs, err := newGameState(
		Config{TimeControl: 500 * time.Millisecond}, nil, nil, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	type moveTest struct {
		move  Move
		valid bool
		score bool
		sleep time.Duration
	}

	runOutOfTime := []moveTest{
		{move: Move{X: 0, Y: 0, D: DirRight},
			valid: true, score: false, sleep: 50 * time.Millisecond},
		{move: Move{X: 6, Y: 0, D: DirLeft},
			valid: true, score: false, sleep: 50 * time.Millisecond},
		{move: Move{X: 1, Y: 0, D: DirRight},
			valid: true, score: false, sleep: 50 * time.Millisecond},
		{move: Move{X: 5, Y: 0, D: DirLeft},
			valid: true, score: false, sleep: 50 * time.Millisecond},
		// ko rule
		{move: Move{X: 1, Y: 0, D: DirRight},
			valid: false, score: false, sleep: 50 * time.Millisecond},
		// out of turn
		{move: Move{X: 4, Y: 0, D: DirLeft},
			valid: false, score: false, sleep: 50 * time.Millisecond},
		{move: Move{X: 1, Y: 0, D: DirDown},
			valid: true, score: false, sleep: 50 * time.Millisecond},
		{move: Move{X: 3, Y: 0, D: DirDown},
			valid: true, score: false, sleep: 50 * time.Millisecond},
		{move: Move{X: 0, Y: 1, D: DirRight},
			valid: true, score: false, sleep: 50 * time.Millisecond},
		// timeout
		{move: Move{X: 3, Y: 1, D: DirDown},
			valid: false, score: false, sleep: 450 * time.Millisecond},
	}

	playTestCases := func(moves []moveTest) {
		// For diffing scores between test cases
		prevScores := make(map[AgentColor]int)
		prevPlayer := gs.lastSnapshot().whoseTurn

		for idx, testCase := range moves {
			time.Sleep(testCase.sleep)
			if ok := gs.ExecuteMove(testCase.move); ok != testCase.valid {
				t.Errorf("moves[%d]: expected valid == %t, got %t",
					idx, testCase.valid, ok)
			}

			if testCase.valid {
				if gs.lastSnapshot().lastMove == nil ||
					gs.lastSnapshot().lastMove.X != testCase.move.X ||
					gs.lastSnapshot().lastMove.Y != testCase.move.Y ||
					gs.lastSnapshot().lastMove.D != testCase.move.D {
					t.Error("last move did not match expectation")
				}
			}

			score_diff := prevScores[prevPlayer] != gs.agents[prevPlayer].score
			if testCase.score && !score_diff {
				t.Errorf("moves[%d]: expected score, but score didn't change", idx)
			} else if !testCase.score && score_diff {
				t.Errorf("moves[%d]: expected no score, but score changed", idx)
			}

			prevPlayer = gs.lastSnapshot().whoseTurn
			for k, v := range gs.agents {
				prevScores[k] = v.score
			}
		}
	}
	playTestCases(runOutOfTime)
}

func TestUpdateStatus(t *testing.T) {
	type TestCase struct {
		// No need to create a map for the gameState; this will be handled before
		// the test happens
		gs                 gameState
		overrideWhiteScore int
		overrideBlackScore int
		status             Status
	}

	var x, _, B, W Marble = marbleNil, marbleRed, marbleBlack, marbleWhite

	testCases := []TestCase{
		{ // No valid moves for white
			gs: gameState{
        history: []snapshot{
          snapshot{
            board:        [][]Marble{{W, W}, {W, W}},
            whoseTurn:    agentWhite,
          },
        },
				winThreshold: 7,
			},
			status: statusBlackWon,
		},
		{ // No valid moves for black (lost all marbles)
			gs: gameState{
				history: makeSingleSnapshotHistory(
          [][]Marble{{W, x}, {x, x}}, agentBlack),
				winThreshold: 7,
			},
			status: statusWhiteWon,
		},
		{ // Win by score (white)
			gs: gameState{
				history: makeSingleSnapshotHistory(
          [][]Marble{{W, x}, {x, B}}, agentBlack),
				winThreshold: 7,
			},
			overrideWhiteScore: 7,
			status:             statusWhiteWon,
		},
		{ // Win by score (white)
			gs: gameState{
				history: makeSingleSnapshotHistory(
          [][]Marble{{W, x}, {x, B}}, agentWhite),
				winThreshold: 7,
			},
			overrideBlackScore: 7,
			status:             statusBlackWon,
		},
		{ // No win
			gs: gameState{
				history: makeSingleSnapshotHistory(
          [][]Marble{{W, x}, {x, B}}, agentWhite),
				winThreshold: 7,
			},
			status: statusOngoing,
		},
	}

	for idx, test := range testCases {
		test.gs.agents = make(map[AgentColor]*agent)
		test.gs.agents[agentWhite] = &agent{
			score: test.overrideWhiteScore,
		}
		test.gs.agents[agentBlack] = &agent{
			score: test.overrideBlackScore,
		}
    test.gs.updateStatus()
		if actual := test.gs.status; actual != test.status {
			t.Errorf("testCases[%d]: status %d != %d", idx, actual, test.status)
		}
		if test.status != statusOngoing && test.gs.lastSnapshot().whoseTurn != agentNil {
			t.Errorf(
				"testCases[%d]: game not ongoing; expected whoseTurn to be agentNil, "+
					"got %s", idx, test.gs.lastSnapshot().whoseTurn.String())
		}
	}
}

func TestResign(t *testing.T) {
	for _, c := range []AgentColor{agentWhite, agentBlack} {
		onGameOverCalled := false
		onGameOver := func() {
			onGameOverCalled = true
		}
		gs, err := newGameState(
			Config{TimeControl: 1 * time.Minute}, nil, onGameOver, 30*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if !gs.resign(c) {
			t.Error("couldn't resign")
		}
		var expectedStatus Status
		if c == agentWhite {
			expectedStatus = statusBlackWon
		} else {
			expectedStatus = statusWhiteWon
		}

		if gs.status != expectedStatus {
			t.Error("resigning did not yield expected status")
		}

		if !onGameOverCalled {
			t.Error("callback was not called")
		}
	}
}

func TestDrawByRepetition(t *testing.T) {
	// shorter names
	var B, W, R, x Marble = marbleBlack, marbleWhite, marbleRed, marbleNil

	// This is a draw -- neither player has a way to force a win. In fact, if
	// either player breaks the repetition (a3r b2l b3l a2r...), they will lose.
	gs := gameState{
		history: makeSingleSnapshotHistory(
      [][]Marble{ {R, x, x}, {x, B, x}, {W, x, x}, }, agentWhite),
		winThreshold: 1,
		posToCount:   make(map[string]int),
	}

	repeat := []Move{
		Move{Y: 2, X: 0, D: DirRight},
		Move{Y: 1, X: 1, D: DirLeft},
		Move{Y: 2, X: 1, D: DirLeft},
		Move{Y: 1, X: 0, D: DirRight}, // We arrive back at the original position.
		Move{Y: 2, X: 0, D: DirRight},
		Move{Y: 1, X: 1, D: DirLeft},
		Move{Y: 2, X: 1, D: DirLeft},
		Move{Y: 1, X: 0, D: DirRight}, // We arrive back at the original position.
		Move{Y: 2, X: 0, D: DirRight},
	}

	for idx, move := range repeat {
		if !gs.ExecuteMove(move) {
			t.Errorf("move[%d] was considered invalid", idx)
		}
	}

	if gs.status != statusDraw {
		t.Error("expected draw")
	}
}

func TestNotifyOutOfTime(t *testing.T) {
	done := make(chan struct{})
	timeoutCb := func() {
		done <- struct{}{}
	}
	gs, err := newGameState(
		Config{TimeControl: 2 * time.Millisecond}, timeoutCb, nil,
		1*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	gs.ExecuteMove(Move{X: 0, Y: 0, D: DirDown})
	// Do nothing... wait on black to timeout
	<-done
	if gs.status != statusWhiteWon {
		t.Error("expected black to timeout and white to win")
	}
}

func TestFirstMoveDeadline(t *testing.T) {
	done := make(chan struct{})
	timeoutCb := func() {
		done <- struct{}{}
	}
	gs, err := newGameState(
		Config{TimeControl: 1 * time.Millisecond}, timeoutCb, nil,
		1*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// (do nothing)

	<-done
	if gs.status != statusAborted {
		t.Errorf("expected aborted status; got %d", gs.status)
	}
}

func TestInvalidTimeControl(t *testing.T) {
	times := []time.Duration{
		-1 * time.Minute,
		24 * time.Hour,
		0 * time.Minute,
	}
	for _, tc := range times {
		gs, err := newGameState(
			Config{TimeControl: tc}, nil, nil, 30*time.Second)
		if err == nil {
			t.Error(err)
		}
		if gs != nil {
			t.Errorf("expected error when creating game with time %s", tc.String())
		}
	}
}
