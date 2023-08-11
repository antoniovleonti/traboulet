package kuba

import (
  "time"
  "sync"
  // "math"
  // "fmt"
)


type Status int
const (
  statusOngoing Status = iota
  statusWhiteWon
  statusBlackWon
  statusDraw
)

func (s Status) String() string {
  if s == statusOngoing {
    return "ONGOING"
  } else if s == statusWhiteWon {
    return "WHITE_WON"
  } else if s == statusBlackWon {
    return "BLACK_WON"
  } else if s == statusDraw {
    return "DRAW"
  } else {
    panic("Invalid Status!")
  }
}

type kubaGame struct {
  board BoardT
  agents map[AgentColor]*agent
  ko *Move
  lastMove *LastMoveT
  whoseTurn AgentColor
  winThreshold int
  status Status
  clockEnabled bool
  mutex sync.RWMutex
}

func newKubaGame(playerTime time.Duration) *kubaGame {
  var x, R, B, W Marble = marbleNil, marbleRed, marbleBlack, marbleWhite
  startPosition := [][]Marble{
    {W, W, x, x, x, B, B},
    {W, W, x, R, x, B, B},
    {x, x, R, R, R, x, x},
    {x, R, R, R, R, R, x},
    {x, x, R, R, R, x, x},
    {B, B, x, R, x, W, W},
    {B, B, x, x, x, W, W},
  }

  agents := make(map[AgentColor]*agent)
  agents[agentWhite] = &agent {
    score: 0,
    time: playerTime,
  }
  agents[agentBlack] = &agent {
    score: 0,
    time: playerTime,
  }

  return &kubaGame{
    agents: agents,
    board: startPosition,
    whoseTurn: agentWhite,
    winThreshold: 7,
    clockEnabled: playerTime > 0 * time.Second,
  }
}

func (kg *kubaGame) boardSize() int {
  return len(kg.board)
}

func (kg *kubaGame) isInBounds(x, y int) bool {
  return x >= 0 && x < kg.boardSize() && y >= 0 && y < kg.boardSize()
}

func (kg *kubaGame) ValidateMove(move Move) bool {
  // Check the game is not already over
  if kg.status != statusOngoing {
    return false
  }

  // Validate direction
  if !move.d.isValid() {
    return false
  }

  // Check bounds
  if !kg.isInBounds(move.x, move.y)  ||
     !kg.isInBounds(move.x + move.dx(), move.y + move.dy()) {
    return false
  }

  // Check that move is in turn
  if kg.board[move.y][move.x] != kg.whoseTurn.marble() {
    return false
  }

  // Check ko rule
  if kg.ko != nil && move == *kg.ko {
    return false
  }

  // Check that no piece is blocking this move from behind
  if behindx, behindy := move.x - move.dx(), move.y - move.dy();
      kg.isInBounds(behindx, behindy) &&
      kg.board[behindy][behindx] != marbleNil {
    return false
  }

  // Check that you are not pushing your own piece off the board
  foundEmpty := false
  var x, y int = move.x, move.y
  for ; kg.isInBounds(x, y); x, y = x + move.dx(), y + move.dy() {
    if kg.board[y][x] == marbleNil {
      foundEmpty = true
      break
    }
  }
  if !foundEmpty {
    // Move back one step to the last valid position
    y -= move.dy()
    x -= move.dx()
    if kg.board[x][y] == kg.whoseTurn.marble() {
      return false
    }
  }

  return true
}

func (kg *kubaGame) validMoveExists() bool {
  for x := 0; x < kg.boardSize(); x++ {
    for y := 0; y < kg.boardSize(); y++ {
      if kg.board[y][x] != kg.whoseTurn.marble() {
        continue
      }
      for dx := range []int{-1, 0, 1} {
        for dy := range []int{-1, 0, 1} {
          if ok, dir := directionFromDxDy(dx, dy);
              ok && kg.ValidateMove(Move{ x: x, y: y, d: dir }) {
            return true
          }
        }
      }
    }
  }
  return false
}

func (kg *kubaGame) updateStatus() Status {
  // Check for preexisting "sticky" status
  if kg.status != statusOngoing {
    return kg.status
  }

  // Win by score
  for t, p := range kg.agents {
    if p.score >= kg.winThreshold {
      return t.winStatus()
    }
  }

  // Win by entrapment
  if !kg.validMoveExists() {
    return kg.whoseTurn.otherAgent().winStatus()
  }

  return statusOngoing
}

func (kg *kubaGame) ExecuteMove(move Move) bool {
  kg.mutex.Lock()
  defer kg.mutex.Unlock()

  if !kg.ValidateMove(move) {
    return false
  }

  tmp := marbleNil
  var x, y int = move.x, move.y
  for ; kg.isInBounds(x, y); x, y = x + move.dx(), y + move.dy() {
    kg.board[y][x], tmp = tmp, kg.board[y][x]  // swap
    if tmp == marbleNil {
      break;
    }
  }
  // A red marble was pushed off the board
  if !kg.isInBounds(x, y) && tmp == marbleRed {
    kg.agents[kg.whoseTurn].score++
  }
  // Check for ko
  kg.ko = nil
  if !kg.isInBounds(x, y) {
    x -= move.dx()
    y -= move.dy()
  }
  if kg.board[y][x] == kg.whoseTurn.otherAgent().marble() {
    kg.ko = &Move{
      x: x,
      y: y,
      d: move.d.reverse(),
    }
  }

  if kg.clockEnabled && !kg.agents[kg.whoseTurn].endTurn() {
    panic("end turn failed!")
  }

  kg.whoseTurn = kg.whoseTurn.otherAgent()

  kg.status = kg.updateStatus()
  if kg.status == statusOngoing && kg.clockEnabled {
    if !kg.agents[kg.whoseTurn].startTurn(kg.playerTimeoutCallback) {
      panic("startTurn failed!")
    }
  }

  return true
}

func (kg *kubaGame) playerTimeoutCallback() {
  kg.mutex.Lock()
  defer kg.mutex.Unlock()

  // The other team just won
  kg.status = kg.whoseTurn.otherAgent().winStatus()
}

func (kg *kubaGame) resign(agent AgentColor) bool {
  kg.mutex.Lock()
  defer kg.mutex.Unlock()

  if kg.status != statusOngoing {
    return false
  }
  kg.status = agent.otherAgent().winStatus()
  return true
}
