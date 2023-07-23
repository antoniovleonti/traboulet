package kuba

import (
  "testing"
  // "fmt"
)

func TestCreateDefaultKubaState(t *testing.T) {
  ok, kuba := CreateKubaState(GetDefaultKubaConfig())

  if !ok {
    t.Error("CreateKubaState returned not ok")
  }

  if kuba == nil {
    t.Error("var kuba should not be nil")
  }
}

func TestIsInBounds(t *testing.T) {
  ok, kuba := CreateKubaState(GetDefaultKubaConfig())
  if !ok {
    t.FailNow()
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
  var B, W, R, x Marble = kMarbleBlack, kMarbleWhite, kMarbleRed, kMarbleNil

  config := KubaConfig{
    StartingPosition: [][]Marble{
      {W, W, W, W, W, W, W},
      {W, W, x, W, W, W, W},
      {W, W, W, W, W, W, B},
      {W, W, W, W, W, W, R},
      {W, B, W, W, W, W, W},
      {B, B, W, W, W, W, x},
      {W, x, x, x, x, B, B},
    },
    WinThreshold: 7,
    StartingPlayer: kPlayerWhite,
    StartingScoreWhite: 0,
    StartingScoreBlack: 0,
  }
  ok, kuba := CreateKubaState(config)
  if !ok {
    t.FailNow()
  }

  validCases := []Move{
    Move{ Y: 1, X: 0, Dy: 0, Dx: 1, },
    Move{ Y: 2, X: 0, Dy: 0, Dx: 1, },
    Move{ Y: 3, X: 0, Dy: 0, Dx: 1, },
    Move{ Y: 3, X: 0, Dy: 0, Dx: 1, },
    Move{ Y: 6, X: 0, Dy: 0, Dx: 1, },
  }

  invalidCases := []Move{
    Move{ Y: 0, X: 0, Dy: 0, Dx: 1, }, // Kills own marble
    Move{ Y: 5, X: 0, Dy: 0, Dx: 1, }, // Wrong color marble
    Move{ Y: 1, X: 1, Dy: 0, Dx: 1, }, // Blocked from behind
    Move{ Y: 6, X: 0, Dy: 0, Dx: -1, }, // Kills own marble
    Move{ Y: 0, X: 0, Dy: 1, Dx: 0, }, // Kills own marble
    Move{ Y: 0, X: 0, Dy: 1, Dx: 1, }, // Nonsense dx, dy
    Move{ Y: 0, X: 0, Dy: 2, Dx: 0, }, // Nonsense dx, dy
    Move{ Y: 0, X: 0, Dy: -1, Dx: 1, }, // Nonsense dx, dy
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

func TestPlayerToMarbleConversion(t *testing.T) {
  if kPlayerWhite.marble() != kMarbleWhite {
    t.Fail()
  }
  if kPlayerBlack.marble() != kMarbleBlack {
    t.Fail()
  }
  if kPlayerNoOne.marble() != kMarbleNil {
    t.Fail()
  }
}

func TestOtherPlayer(t *testing.T) {
  if kPlayerWhite.otherPlayer() != kPlayerBlack {
    t.Fail()
  }
  if kPlayerBlack.otherPlayer() != kPlayerWhite {
    t.Fail()
  }
}

func TestExecuteMove(t *testing.T) {
  ok, kuba := CreateKubaState(GetDefaultKubaConfig())
  if !ok {
    t.FailNow()
  }

  moves := []struct{
    move Move
    valid bool
    score bool
  }{
    { move: Move{ X: 0, Y: 0, Dx: 1, Dy: 0}, valid: true, score: false},
    { move: Move{ X: 6, Y: 0, Dx: -1, Dy: 0}, valid: true, score: false },
    { move: Move{ X: 1, Y: 0, Dx: 1, Dy: 0}, valid: true, score: false },
    { move: Move{ X: 5, Y: 0, Dx: -1, Dy: 0}, valid: true, score: false },
    // ko rule
    { move: Move{ X: 1, Y: 0, Dx: 1, Dy: 0}, valid: false, score: false },
    // out of turn
    { move: Move{ X: 4, Y: 0, Dx: -1, Dy: 0}, valid: false, score: false },
    { move: Move{ X: 1, Y: 0, Dx: 0, Dy: 1}, valid: true, score: false },
    { move: Move{ X: 3, Y: 0, Dx: 0, Dy: 1}, valid: true, score: false },
    { move: Move{ X: 0, Y: 1, Dx: 1, Dy: 0}, valid: true, score: false },
    { move: Move{ X: 3, Y: 1, Dx: 0, Dy: 1}, valid: true, score: true },
  }

  // For diffing scores between test cases
  prevScores := make(map[Player]int)
  prevPlayer := kuba.Turn

  for idx, testCase := range moves {
    if ok, _ := kuba.ExecuteMove(testCase.move); ok != testCase.valid {
      t.Errorf("moves[%d]: expected valid == %t, got %t",
               idx, testCase.valid, ok)
    }

    score_diff := prevScores[prevPlayer] != kuba.PlayerToScore[prevPlayer]
    if testCase.score && !score_diff {
      t.Errorf("moves[%d]: expected score, but score didn't change", idx)
    } else if !testCase.score && score_diff {
      t.Errorf("moves[%d]: expected no score, but score changed", idx)
    }

    prevPlayer = kuba.Turn
    for k, v := range kuba.PlayerToScore {
      prevScores[k] = v
    }
  }
}

func TestPlayerToWinStatus(t *testing.T) {
  if kPlayerWhite.winStatus() != kWhiteWon {
    t.Fail()
  }
  if kPlayerBlack.winStatus() != kBlackWon {
    t.Fail()
  }
}

func TestGetStatus(t *testing.T) {
  type TestCase struct {
    config KubaConfig
    status StatusT
  }

  var x, _, B, W Marble = kMarbleNil, kMarbleRed, kMarbleBlack, kMarbleWhite

  testCases := []TestCase {
    {  // No valid moves for white
      config: KubaConfig {
        StartingPosition: [][]Marble{ {W, W}, {W, W}, },
        WinThreshold: 7,
        StartingPlayer: kPlayerWhite,
      },
      status: kBlackWon,
    },
    {  // No valid moves for black (lost all marbles)
      config: KubaConfig {
        StartingPosition: [][]Marble{ {W, x}, {x, x}, },
        WinThreshold: 7,
        StartingPlayer: kPlayerBlack,
      },
      status: kWhiteWon,
    },
    {  // Win by score (white)
      config: KubaConfig {
        StartingPosition: [][]Marble{ {W, x}, {x, B}, },
        WinThreshold: 7,
        StartingPlayer: kPlayerBlack,
        StartingScoreWhite: 7,
      },
      status: kWhiteWon,
    },
    {  // Win by score (white)
      config: KubaConfig {
        StartingPosition: [][]Marble{ {W, x}, {x, B}, },
        WinThreshold: 7,
        StartingPlayer: kPlayerWhite,
        StartingScoreBlack: 7,
      },
      status: kBlackWon,
    },
    {  // No win
      config: KubaConfig {
        StartingPosition: [][]Marble{ {W, x}, {x, B}, },
        WinThreshold: 7,
        StartingPlayer: kPlayerWhite,
      },
      status: kOngoing,
    },
  }

  for idx, test := range testCases {
    ok, kuba := CreateKubaState(test.config)
    if !ok {
      t.Errorf("testCases[%d]: CreateKubaState failed", idx)
      continue
    }

    if actual := kuba.getStatus(); actual != test.status {
      t.Errorf("testCases[%d]: status %d != %d", idx, actual, test.status)
    }
  }
}
