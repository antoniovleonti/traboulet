package kuba

import (
  // "math"
  // "fmt"
)

type Marble int
const (
  kMarbleNil Marble = iota
  kMarbleWhite
  kMarbleBlack
  kMarbleRed
)

type Player int
const (
  kPlayerNoOne Player = iota
  kPlayerWhite
  kPlayerBlack
)

type KubaConfig struct {
  StartingPosition [][]Marble
  WinThreshold int
  StartingPlayer Player
  StartingScoreWhite int
  StartingScoreBlack int
  MaxMoves int
}

type Move struct {
  X int
  Y int
  Dx int
  Dy int
}

type LastMoveT struct {
  X int
  Y int
  Dx int
  Dy int
  Length int
}

type StatusT int
const (
  kOngoing StatusT = iota
  kWhiteWon
  kBlackWon
  kDraw
)

type KubaState struct {
  Board [][]Marble
  PlayerToScore map[Player]int
  Ko *Move
  LastMove *LastMoveT
  Turn Player
  WinThreshold int
  Status StatusT
  MoveCount int
  MaxMoves int
}

func (p Player) marble() Marble {
  return Marble(p)
}

func (p Player) otherPlayer() Player {
  return p % 2 + 1
}

func (p Player) winStatus() StatusT {
  return StatusT(p)
}

func GetDefaultKubaConfig() KubaConfig {
  var x, R, B, W Marble = kMarbleNil, kMarbleRed, kMarbleBlack, kMarbleWhite
  return KubaConfig{
    StartingPosition: [][]Marble{
      {W, W, x, x, x, B, B},
      {W, W, x, R, x, B, B},
      {x, x, R, R, R, x, x},
      {x, R, R, R, R, R, x},
      {x, x, R, R, R, x, x},
      {B, B, x, R, x, W, W},
      {B, B, x, x, x, W, W},
    },
    WinThreshold: 7,
    StartingPlayer: kPlayerWhite,
    StartingScoreWhite: 0,
    StartingScoreBlack: 0,
  }
}

func CreateKubaState(config KubaConfig) (
    ok bool, state *KubaState) {
  boardSize := len(config.StartingPosition)
  for _, row := range config.StartingPosition {
    if boardSize != len(row) {
      return false, nil
    }
  }

  scores := make(map[Player]int)
  scores[kPlayerWhite] = config.StartingScoreWhite
  scores[kPlayerBlack] = config.StartingScoreBlack

  return true, &KubaState{
    PlayerToScore: scores,
    Board: config.StartingPosition,
    Turn: config.StartingPlayer,
    WinThreshold: config.WinThreshold,
    Ko: nil,
    LastMove: nil,
  }
}

func (ks *KubaState) boardSize() int {
  return len(ks.Board)
}

func (ks *KubaState) isInBounds(x, y int) bool {
  return x >= 0 && x < ks.boardSize() && y >= 0 && y < ks.boardSize()
}

func (ks *KubaState) ValidateMove(move Move) bool {
  // Check the game is not already over
  if ks.Status != kOngoing {
    return false
  }

  // Check that dx, dy makes sense
  absDx, absDy := move.Dx, move.Dy
  if move.Dx < 0 {
    absDx *= -1
  }
  if move.Dy < 0 {
    absDy *= -1
  }
  if ((absDx == 1) == (absDy == 1)) || (absDx > 1 || absDy > 1) {
    return false
  }

  // Check bounds
  if !ks.isInBounds(move.X, move.Y)  ||
     !ks.isInBounds(move.X + move.Dx, move.Y + move.Dy) {
    return false
  }

  // Check that move is in turn
  if ks.Board[move.Y][move.X] != ks.Turn.marble() {
    return false
  }

  // Check ko rule
  if ks.Ko != nil && move == *ks.Ko {
    return false
  }

  // Check that no piece is blocking this move from behind
  if behindx, behindy := move.X - move.Dx, move.Y - move.Dy;
      ks.isInBounds(behindx, behindy) &&
      ks.Board[behindy][behindx] != kMarbleNil {
    return false
  }

  // Check that you are not pushing your own piece off the board
  foundEmpty := false
  var x, y int = move.X, move.Y
  for ; ks.isInBounds(x, y); x, y = x + move.Dx, y + move.Dy {
    if ks.Board[y][x] == kMarbleNil {
      foundEmpty = true
      break
    }
  }
  if !foundEmpty {
    // Move back one step to the last valid position
    y -= move.Dy
    x -= move.Dx
    if ks.Board[x][y] == ks.Turn.marble() {
      return false
    }
  }

  return true
}

func (ks *KubaState) validMoveExists() bool {
  for x := 0; x < ks.boardSize(); x++ {
    for y := 0; y < ks.boardSize(); y++ {
      if ks.Board[y][x] != ks.Turn.marble() {
        continue
      }
      for dx := range []int{-1, 0, 1} {
        for dy := range []int{-1, 0, 1} {
          if ks.ValidateMove(Move{ X: x, Y: y, Dx: dx, Dy: dy }) {
            return true
          }
        }
      }
    }
  }
  return false
}

func (ks *KubaState) getStatus() StatusT {
  // Check for preexisting "sticky" status
  if ks.Status != kOngoing {
    return ks.Status
  }

  // Win by score
  for k, v := range ks.PlayerToScore {
    if v >= ks.WinThreshold {
      return k.winStatus()
    }
  }

  // Win by entrapment
  if !ks.validMoveExists() {
    return ks.Turn.otherPlayer().winStatus()
  }

  return kOngoing
}

func (ks *KubaState) ExecuteMove(move Move) (ok bool, status StatusT) {
  if !ks.ValidateMove(move) {
    return false, ks.Status
  }

  tmp := kMarbleNil
  var x, y int = move.X, move.Y
  for ; ks.isInBounds(x, y); x, y = x + move.Dx, y + move.Dy {
    ks.Board[y][x], tmp = tmp, ks.Board[y][x]  // swap
    if tmp == kMarbleNil {
      break;
    }
  }
  // A red marble was pushed off the board
  if !ks.isInBounds(x, y) && tmp == kMarbleRed {
    ks.PlayerToScore[ks.Turn]++
  }
  // Check for ko
  ks.Ko = nil
  if !ks.isInBounds(x, y) {
    x -= move.Dx
    y -= move.Dy
  }
  if ks.Board[y][x] == ks.Turn.otherPlayer().marble() {
    ks.Ko = &Move{
      X: x,
      Y: y,
      Dx: -move.Dx,
      Dy: -move.Dy,
    }
  }

  ks.Turn = ks.Turn.otherPlayer()
  ks.Status = ks.getStatus()
  return true, ks.Status
}
