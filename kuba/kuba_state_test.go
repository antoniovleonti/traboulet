package kuba

import (
	"testing"
	"time"
	// "fmt"
)

func TestCreateDefaultKubaState(t *testing.T) {
	kubaWOClock, err := newKubaGame(
		Config{TimeControl: time.Minute}, nil, nil, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if kubaWOClock == nil {
		t.Error("var kuba should not be nil")
	}

	kubaWClock, err := newKubaGame(
		Config{TimeControl: 60 * time.Second}, nil, nil, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if kubaWOClock == nil {
		t.Error("var kuba should not be nil")
	}

	if !kubaWClock.clockEnabled {
		t.Error("clock should be enabled")
	}
}

func TestIsInBounds(t *testing.T) {
	kuba, err := newKubaGame(
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
		if !kuba.isInBounds(testCase[0], testCase[1]) {
			t.Errorf("inBoundsCases[%d] (%d, %d) is out of bounds.",
				idx, testCase[0], testCase[1])
		}
	}

	for idx, testCase := range outOfBoundsCases {
		if kuba.isInBounds(testCase[0], testCase[1]) {
			t.Errorf("outOfBoundsCases[%d] (%d, %d) is in bounds.",
				idx, testCase[0], testCase[1])
		}
	}
}

func TestValidateMove(t *testing.T) {
	// shorter names
	var B, W, R, x Marble = marbleBlack, marbleWhite, marbleRed, marbleNil

	kuba := kubaGame{
		board: [][]Marble{
			{W, W, W, W, W, W, W},
			{W, W, x, W, W, W, W},
			{W, W, W, W, W, W, B},
			{W, W, W, W, W, W, R},
			{W, B, W, W, W, W, W},
			{B, B, W, W, W, W, x},
			{W, x, x, x, x, B, B},
		},
		winThreshold: 7,
		whoseTurn:    agentWhite,
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
		if !kuba.ValidateMove(valid) {
			t.Errorf("validCases[%d] was considered invalid", idx)
		}
	}
	for idx, invalid := range invalidCases {
		if kuba.ValidateMove(invalid) {
			t.Errorf("invalidCases[%d] was considered valid", idx)
		}
	}
}

func TestExecuteMove(t *testing.T) {
	kuba, err := newKubaGame(
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
		prevPlayer := kuba.whoseTurn

		for idx, testCase := range moves {
			time.Sleep(testCase.sleep)
			if ok := kuba.ExecuteMove(testCase.move); ok != testCase.valid {
				t.Errorf("moves[%d]: expected valid == %t, got %t",
					idx, testCase.valid, ok)
			}

			if testCase.valid {
				if kuba.lastMove == nil ||
					kuba.lastMove.X != testCase.move.X ||
					kuba.lastMove.Y != testCase.move.Y ||
					kuba.lastMove.D != testCase.move.D {
					t.Error("last move did not match expectation")
				}
			}

			score_diff := prevScores[prevPlayer] != kuba.agents[prevPlayer].score
			if testCase.score && !score_diff {
				t.Errorf("moves[%d]: expected score, but score didn't change", idx)
			} else if !testCase.score && score_diff {
				t.Errorf("moves[%d]: expected no score, but score changed", idx)
			}

			prevPlayer = kuba.whoseTurn
			for k, v := range kuba.agents {
				prevScores[k] = v.score
			}
		}
	}
	playTestCases(runOutOfTime)
}

func TestGetStatus(t *testing.T) {
	type TestCase struct {
		// No need to create a map for the kubaGame; this will be handled before
		// the test happens
		kuba               kubaGame
		overrideWhiteScore int
		overrideBlackScore int
		status             Status
	}

	var x, _, B, W Marble = marbleNil, marbleRed, marbleBlack, marbleWhite

	testCases := []TestCase{
		{ // No valid moves for white
			kuba: kubaGame{
				board:        [][]Marble{{W, W}, {W, W}},
				winThreshold: 7,
				whoseTurn:    agentWhite,
			},
			status: statusBlackWon,
		},
		{ // No valid moves for black (lost all marbles)
			kuba: kubaGame{
				board:        [][]Marble{{W, x}, {x, x}},
				winThreshold: 7,
				whoseTurn:    agentBlack,
			},
			status: statusWhiteWon,
		},
		{ // Win by score (white)
			kuba: kubaGame{
				board:        [][]Marble{{W, x}, {x, B}},
				winThreshold: 7,
				whoseTurn:    agentBlack,
			},
			overrideWhiteScore: 7,
			status:             statusWhiteWon,
		},
		{ // Win by score (white)
			kuba: kubaGame{
				board:        [][]Marble{{W, x}, {x, B}},
				winThreshold: 7,
				whoseTurn:    agentWhite,
			},
			overrideBlackScore: 7,
			status:             statusBlackWon,
		},
		{ // No win
			kuba: kubaGame{
				board:        [][]Marble{{W, x}, {x, B}},
				winThreshold: 7,
				whoseTurn:    agentWhite,
			},
			status: statusOngoing,
		},
	}

	for idx, test := range testCases {
		test.kuba.agents = make(map[AgentColor]*agent)
		test.kuba.agents[agentWhite] = &agent{
			score: test.overrideWhiteScore,
		}
		test.kuba.agents[agentBlack] = &agent{
			score: test.overrideBlackScore,
		}
		if actual := test.kuba.updateStatus(); actual != test.status {
			t.Errorf("testCases[%d]: status %d != %d", idx, actual, test.status)
		}
	}
}

func TestResign(t *testing.T) {
	for _, c := range []AgentColor{agentWhite, agentBlack} {
		onGameOverCalled := false
		onGameOver := func() {
			onGameOverCalled = true
		}
		kuba, err := newKubaGame(
			Config{TimeControl: 1 * time.Minute}, nil, onGameOver, 30*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if !kuba.resign(c) {
			t.Error("couldn't resign")
		}
		var expectedStatus Status
		if c == agentWhite {
			expectedStatus = statusBlackWon
		} else {
			expectedStatus = statusWhiteWon
		}

		if kuba.status != expectedStatus {
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
	kuba := kubaGame{
		board: [][]Marble{
			{R, x, x},
			{x, B, x},
			{W, x, x},
		},
		winThreshold: 1,
		whoseTurn:    agentWhite,
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
		if !kuba.ExecuteMove(move) {
			t.Errorf("move[%d] was considered invalid", idx)
		}
	}

	if kuba.status != statusDraw {
		t.Error("expected draw")
	}
}

func TestNotifyOutOfTime(t *testing.T) {
	done := make(chan struct{})
	timeoutCb := func() {
		done <- struct{}{}
	}
	kuba, err := newKubaGame(
		Config{TimeControl: 2 * time.Millisecond}, timeoutCb, nil,
		1*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	kuba.ExecuteMove(Move{X: 0, Y: 0, D: DirDown})
	// Do nothing... wait on black to timeout
	<-done
	if kuba.status != statusWhiteWon {
		t.Error("expected black to timeout and white to win")
	}
}

func TestFirstMoveDeadline(t *testing.T) {
	done := make(chan struct{})
	timeoutCb := func() {
		done <- struct{}{}
	}
	kuba, err := newKubaGame(
		Config{TimeControl: 1 * time.Millisecond}, timeoutCb, nil,
		1*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// (do nothing)

	<-done
	if kuba.status != statusAborted {
		t.Errorf("expected aborted status; got %d", kuba.status)
	}
}

func TestInvalidTimeControl(t *testing.T) {
	times := []time.Duration{
		-1 * time.Minute,
		24 * time.Hour,
		0 * time.Minute,
	}
	for _, tc := range times {
		kuba, err := newKubaGame(
			Config{TimeControl: tc}, nil, nil, 30*time.Second)
		if err == nil {
			t.Error(err)
		}
		if kuba != nil {
			t.Errorf("expected error when creating game with time %s", tc.String())
		}
	}
}
