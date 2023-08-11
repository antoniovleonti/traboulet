package kuba

import (
  "testing"
  "time"
  // "fmt"
)

func TestCreateDefaultKubaState(t *testing.T) {
  kubaWOClock := newKubaGame(0 * time.Second)
  if kubaWOClock == nil {
    t.Error("var kuba should not be nil")
  }

  if kubaWOClock.clockEnabled {
    t.Error("clock should not be enabled")
  }

  kubaWClock := newKubaGame(60 * time.Second)
  if kubaWOClock == nil {
    t.Error("var kuba should not be nil")
  }

  if !kubaWClock.clockEnabled {
    t.Error("clock should be enabled")
  }
}

func TestIsInBounds(t *testing.T) {
  kuba := newKubaGame(0 * time.Second)

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

  kuba := kubaGame {
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
    whoseTurn: agentWhite,
  }

  validCases := []Move{
    Move{ y: 1, x: 0, d: DirRight },
    Move{ y: 2, x: 0, d: DirRight },
    Move{ y: 3, x: 0, d: DirRight },
    Move{ y: 3, x: 0, d: DirRight },
    Move{ y: 6, x: 0, d: DirRight },
  }

  invalidCases := []Move{
    Move{ y: 0, x: 0, d: DirRight }, // Kills own marble
    Move{ y: 5, x: 0, d: DirRight }, // Wrong color marble
    Move{ y: 1, x: 1, d: DirRight }, // Blocked from behind
    Move{ y: 6, x: 0, d: DirLeft }, // Kills own marble
    Move{ y: 0, x: 0, d: DirDown }, // Kills own marble
    Move{ y: 0, x: 0, d: DirNil }, // Nonsense dx, dy
    Move{ y: 0, x: 0, d: Direction(8)}, // Nonsense dx, dy
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
  kuba := newKubaGame(500 * time.Millisecond)

  type moveTest struct {
    move Move
    valid bool
    score bool
    sleep time.Duration
  }

  runOutOfTime := []moveTest{
    { move: Move{ x: 0, y: 0, d: DirRight },
      valid: true, score: false, sleep: 50 * time.Millisecond },
    { move: Move{ x: 6, y: 0, d: DirLeft },
      valid: true, score: false, sleep: 50 * time.Millisecond },
    { move: Move{ x: 1, y: 0, d: DirRight },
      valid: true, score: false, sleep: 50 * time.Millisecond },
    { move: Move{ x: 5, y: 0, d: DirLeft },
      valid: true, score: false, sleep: 50 * time.Millisecond },
    // ko rule
    { move: Move{ x: 1, y: 0, d: DirRight },
      valid: false, score: false, sleep: 50 * time.Millisecond },
    // out of turn
    { move: Move{ x: 4, y: 0, d: DirLeft },
      valid: false, score: false, sleep: 50 * time.Millisecond },
    { move: Move{ x: 1, y: 0, d: DirDown },
      valid: true, score: false, sleep: 50 * time.Millisecond },
    { move: Move{ x: 3, y: 0, d: DirDown },
      valid: true, score: false, sleep: 50 * time.Millisecond },
    { move: Move{ x: 0, y: 1, d: DirRight },
      valid: true, score: false, sleep: 50 * time.Millisecond },
    // timeout
    { move: Move{ x: 3, y: 1, d: DirDown },
      valid: false, score: false, sleep: 450 * time.Millisecond },
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
    kuba kubaGame
    overrideWhiteScore int
    overrideBlackScore int
    status Status
  }

  var x, _, B, W Marble = marbleNil, marbleRed, marbleBlack, marbleWhite

  testCases := []TestCase {
    {  // No valid moves for white
      kuba: kubaGame {
        board: [][]Marble{ {W, W}, {W, W}, },
        winThreshold: 7,
        whoseTurn: agentWhite,
      },
      status: statusBlackWon,
    },
    {  // No valid moves for black (lost all marbles)
      kuba: kubaGame {
        board: [][]Marble{ {W, x}, {x, x}, },
        winThreshold: 7,
        whoseTurn: agentBlack,
      },
      status: statusWhiteWon,
    },
    {  // Win by score (white)
      kuba: kubaGame {
        board: [][]Marble{ {W, x}, {x, B}, },
        winThreshold: 7,
        whoseTurn: agentBlack,
      },
      overrideWhiteScore: 7,
      status: statusWhiteWon,
    },
    {  // Win by score (white)
      kuba: kubaGame {
        board: [][]Marble{ {W, x}, {x, B}, },
        winThreshold: 7,
        whoseTurn: agentWhite,
      },
      overrideBlackScore: 7,
      status: statusBlackWon,
    },
    {  // No win
      kuba: kubaGame {
        board: [][]Marble{ {W, x}, {x, B}, },
        winThreshold: 7,
        whoseTurn: agentWhite,
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
    kuba := newKubaGame(0 * time.Second)
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
  }
}
