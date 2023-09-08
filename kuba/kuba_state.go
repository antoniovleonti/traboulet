package kuba

import (
	"fmt"
	"sync"
	"time"
)

type Status int

const (
	statusOngoing Status = iota
	statusWhiteWon
	statusBlackWon
	statusDraw
	statusAborted
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
	} else if s == statusAborted {
		return "ABORTED"
	} else {
		panic(fmt.Sprintf("Invalid Status %d!", s))
	}
}

type kubaGame struct {
	board             BoardT
	agents            map[AgentColor]*agent
	ko                *Move
	lastMove          *LastMoveT
	whoseTurn         AgentColor
	winThreshold      int
	status            Status
	posToCount        map[string]int // string rep of pos -> # of times it's occured.
  validMoves        []Move
	firstMoveDeadline *time.Time
	firstMoveTimer    *time.Timer
	mutex             sync.RWMutex
	onAsyncUpdate     func()
	onGameOver        func()
}

func newKubaGame(
	config Config, onAsyncUpdate func(), onGameOver func(),
	firstMoveTimeout time.Duration) (*kubaGame, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
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
	agents[agentWhite] = &agent{
		score: 0,
		time:  config.TimeControl,
	}
	agents[agentBlack] = &agent{
		score: 0,
		time:  config.TimeControl,
	}

	firstMoveDeadline := time.Now().Add(firstMoveTimeout)

	kg := kubaGame{
		agents:            agents,
		board:             startPosition,
		whoseTurn:         agentWhite,
		winThreshold:      7,
		posToCount:        make(map[string]int),
		firstMoveDeadline: &firstMoveDeadline,
		onAsyncUpdate:     onAsyncUpdate,
		onGameOver:        onGameOver,
	}
	kg.firstMoveTimer =
		time.AfterFunc(firstMoveTimeout, kg.firstMoveTimeoutCallback)

  kg.validMoves = kg.getValidMoves()

	return &kg, nil
}

func (kg *kubaGame) boardSize() int {
	return len(kg.board)
}

func (kg kubaGame) getPositionString() string {
	return fmt.Sprintf("%v;%d", kg.board, kg.whoseTurn)
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
	if !move.D.isValid() {
		return false
	}

	// Check bounds
	if !kg.isInBounds(move.X, move.Y) ||
		!kg.isInBounds(move.X+move.dx(), move.Y+move.dy()) {
		return false
	}

	// Check that move is in turn
	if kg.board[move.Y][move.X] != kg.whoseTurn.marble() {
		return false
	}

	// Check ko rule
	if kg.ko != nil && move == *kg.ko {
		return false
	}

	// Check that no piece is blocking this move from behind
	behindx, behindy := move.X-move.dx(), move.Y-move.dy()
	if kg.isInBounds(behindx, behindy) &&
		kg.board[behindy][behindx] != marbleNil {
		return false
	}

	// Check that you are not pushing your own piece off the board
	foundEmpty := false
	var x, y int = move.X, move.Y
	for ; kg.isInBounds(x, y); x, y = x+move.dx(), y+move.dy() {
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
  kg.validMoves = kg.getValidMoves()
	if len(kg.validMoves) == 0 {
		return kg.whoseTurn.otherAgent().winStatus()
	}

	// Draw by repetition
	if kg.posToCount[kg.getPositionString()] >= 3 {
		return statusDraw
	}

	return statusOngoing
}

func (kg *kubaGame) ExecuteMove(move Move) bool {
	kg.mutex.Lock()
	defer kg.mutex.Unlock()

	if !kg.ValidateMove(move) {
		return false
	}

	marblesMoved := 0
	tmp := marbleNil
	var x, y int = move.X, move.Y
	for ; kg.isInBounds(x, y); x, y = x+move.dx(), y+move.dy() {
		kg.board[y][x], tmp = tmp, kg.board[y][x] // swap
		if tmp == marbleNil {
			break
		}
		marblesMoved++
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
			X: x,
			Y: y,
			D: move.D.reverse(),
		}
	}

	if kg.firstMoveTimer != nil {
		if !kg.firstMoveTimer.Stop() {
			panic("Ending first move timer failed!")
		}
		kg.firstMoveTimer = nil
		kg.firstMoveDeadline = nil
	}

	if !kg.agents[kg.whoseTurn].endTurn() {
		panic("End player turn failed!")
	}

	kg.lastMove = &LastMoveT{
		X:            move.X,
		Y:            move.Y,
		D:            move.D,
		MarblesMoved: marblesMoved,
	}

	kg.whoseTurn = kg.whoseTurn.otherAgent()

	kg.posToCount[kg.getPositionString()]++

	kg.status = kg.updateStatus()
	if kg.status == statusOngoing {
		if !kg.agents[kg.whoseTurn].startTurn(kg.playerTimeoutCallback) {
			panic("startTurn failed!")
		}
	} else {
		kg.teardown()
		if kg.onGameOver != nil {
			kg.onGameOver()
		}
	}

	return true
}

func (kg *kubaGame) playerTimeoutCallback() {
	kg.mutex.Lock()
	defer kg.mutex.Unlock()

	// The other team just won
	kg.status = kg.whoseTurn.otherAgent().winStatus()

	kg.teardown()

	// Notify front-end of update
	if kg.onAsyncUpdate != nil {
		kg.onAsyncUpdate()
	}
	if kg.onGameOver != nil {
		kg.onGameOver()
	}
}

func (kg *kubaGame) firstMoveTimeoutCallback() {
	kg.mutex.Lock()
	defer kg.mutex.Unlock()

	// The other team just won
	kg.status = statusAborted
	kg.firstMoveDeadline = nil

	kg.teardown()

	// Notify front-end of update
	if kg.onAsyncUpdate != nil {
		kg.onAsyncUpdate()
	}
	if kg.onGameOver != nil {
		kg.onGameOver()
	}
}

func (kg *kubaGame) resign(agent AgentColor) bool {
	kg.mutex.Lock()
	defer kg.mutex.Unlock()

	if kg.status != statusOngoing {
		return false
	}
	kg.status = agent.otherAgent().winStatus()

	if kg.onGameOver != nil {
		kg.onGameOver()
	}
	return true
}

// Bring the game to a completely inert state-- ensure no timers are gonna fire
// or anything like that.
func (kg *kubaGame) teardown() {
	if kg.firstMoveTimer != nil {
		kg.firstMoveTimer.Stop()
		kg.firstMoveTimer = nil
	}
	for _, a := range kg.agents {
		a.endTurn()
	}
}

func (kg *kubaGame) getValidMoves() []Move {
  var moves []Move
	for x := 0; x < kg.boardSize(); x++ {
		for y := 0; y < kg.boardSize(); y++ {
			if kg.board[y][x] != kg.whoseTurn.marble() {
				continue
			}
			for _, dir := range []Direction{DirUp, DirDown, DirLeft, DirRight} {
        move := Move{X: x, Y: y, D: dir}
				if kg.ValidateMove(move) {
					moves = append(moves, move)
				}
			}
		}
	}
	return moves
}
