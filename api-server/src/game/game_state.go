package game

import (
	"encoding/json"
	"errors"
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

type snapshot struct {
	board     BoardT             `json:"board"`
	lastMove  *MoveWMarblesMoved `json:"lastMove"`
	whoseTurn AgentColor         `json:"whoseTurn,string"`
}

func (s snapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Board     BoardT             `json:"board"`
		LastMove  *MoveWMarblesMoved `json:"lastMove"`
		WhoseTurn string             `json:"whoseTurn"`
	}{
		Board:     s.board,
		LastMove:  s.lastMove,
		WhoseTurn: s.whoseTurn.String(),
	})
}

type gameState struct {
	history           []snapshot
	agents            map[AgentColor]*agent
	ko                *Move
	winThreshold      int
	timeControl       time.Duration
	status            Status
	posToCount        map[string]int
	validMoves        []MoveWMarblesMoved
	firstMoveDeadline *time.Time
	firstMoveTimer    *time.Timer
	mutex             sync.RWMutex
	onAsyncUpdate     func()
	onGameOver        func()
}

func newGameState(
	config Config, onAsyncUpdate func(), onGameOver func(),
	firstMoveTimeout time.Duration) (*gameState, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	var x, R, B, W Marble = marbleNil, marbleRed, marbleBlack, marbleWhite
	startPosition := []snapshot{
		snapshot{
			board: [][]Marble{
				{W, W, x, x, x, B, B},
				{W, W, x, R, x, B, B},
				{x, x, R, R, R, x, x},
				{x, R, R, R, R, R, x},
				{x, x, R, R, R, x, x},
				{B, B, x, R, x, W, W},
				{B, B, x, x, x, W, W},
			},
			whoseTurn: agentWhite,
			lastMove:  nil,
		},
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

	gs := gameState{
		history:           startPosition,
		agents:            agents,
		winThreshold:      7,
		timeControl:       config.TimeControl,
		posToCount:        make(map[string]int),
		firstMoveDeadline: &firstMoveDeadline,
		onAsyncUpdate:     onAsyncUpdate,
		onGameOver:        onGameOver,
	}
	gs.firstMoveTimer =
		time.AfterFunc(firstMoveTimeout, gs.firstMoveTimeoutCallback)

	gs.validMoves = gs.getValidMoves()

	return &gs, nil
}

func (gs gameState) boardsize() int {
	return len(gs.history[0].board)
}

func (gs *gameState) lastSnapshot() *snapshot {
	return &gs.history[len(gs.history)-1]
}

func (gs gameState) getPositionString() string {
	return fmt.Sprintf(
		"%v;%d", gs.lastSnapshot().board, gs.lastSnapshot().whoseTurn)
}

func (gs *gameState) isInBounds(x, y int) bool {
	return x >= 0 && x < gs.boardsize() && y >= 0 && y < gs.boardsize()
}

func (gs *gameState) ValidateMove(move Move) (*MoveWMarblesMoved, error) {
	// Check the game is not already over
	if gs.status != statusOngoing {
		return nil, errors.New("Game already ended.")
	}

	// Validate direction
	if !move.D.isValid() {
		return nil, errors.New("Direction is invalid.")
	}

	// Check bounds
	if !gs.isInBounds(move.X, move.Y) ||
		!gs.isInBounds(move.X+move.dx(), move.Y+move.dy()) {
		return nil, errors.New("Index out of bounds.")
	}

	// Check that move is in turn
	if gs.lastSnapshot().board[move.Y][move.X] !=
		gs.lastSnapshot().whoseTurn.marble() {
		return nil, errors.New("Is not an in-turn marble.")
	}

	// Check ko rule
	if gs.ko != nil && move == *gs.ko {
		return nil, errors.New("Prevented by ko.")
	}

	// Check that no piece is blocking this move from behind
	behindx, behindy := move.X-move.dx(), move.Y-move.dy()
	if gs.isInBounds(behindx, behindy) &&
		gs.lastSnapshot().board[behindy][behindx] != marbleNil {
		return nil, errors.New("Blocked by adjacent marble.")
	}

	// Check that you are not pushing your own piece off the board
	foundEmpty := false
	marblesMoved := 0
	var x, y int = move.X, move.Y
	for ; gs.isInBounds(x, y); x, y = x+move.dx(), y+move.dy() {
		if gs.lastSnapshot().board[y][x] == marbleNil {
			foundEmpty = true
			break
		}
		marblesMoved++
	}
	if !foundEmpty {
		// Move back one step to the last valid position
		y -= move.dy()
		x -= move.dx()
		if gs.lastSnapshot().board[y][x] == gs.lastSnapshot().whoseTurn.marble() {
			return nil, errors.New("Can't push own marble off.")
		}
	}

	moveWMarblesMoved := MoveWMarblesMoved{
		X:            move.X,
		Y:            move.Y,
		D:            move.D,
		MarblesMoved: marblesMoved,
	}
	return &moveWMarblesMoved, nil
}

func (gs *gameState) updateStatus() {
	newStatus := statusOngoing
	// If the game is over, it's no one's turn.
	defer func() {
		gs.status = newStatus
		if newStatus != statusOngoing {
      gs.teardown()
		}
	}()
	// Check for preexisting "sticky" status
	if gs.status != statusOngoing {
		newStatus = gs.status
    return
	}

	// Win by score
	for t, p := range gs.agents {
		if p.score >= gs.winThreshold {
			newStatus = t.winStatus()
      return
		}
	}

	// Win by entrapment
	gs.validMoves = gs.getValidMoves()
	if len(gs.validMoves) == 0 {
		newStatus = gs.lastSnapshot().whoseTurn.otherAgent().winStatus()
    return
	}

	// Draw by repetition
	if gs.posToCount[gs.getPositionString()] >= 3 {
		newStatus = statusDraw
    return
	}
}

func (gs *gameState) ExecuteMove(move Move) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	if _, err := gs.ValidateMove(move); err != nil {
		return err
	}
  nextSnapshot := snapshot{
		board:     gs.lastSnapshot().board.deepCopy(),
		whoseTurn: gs.lastSnapshot().whoseTurn.otherAgent(),
		lastMove: &MoveWMarblesMoved{
			X:            move.X,
			Y:            move.Y,
			D:            move.D,
			MarblesMoved: 0, // Will be updated later
		},
	}

	tmp := marbleNil
	var x, y int = move.X, move.Y
	for ; gs.isInBounds(x, y); x, y = x+move.dx(), y+move.dy() {
		nextSnapshot.board[y][x], tmp = tmp, nextSnapshot.board[y][x]
		if tmp == marbleNil {
			break
		}
		nextSnapshot.lastMove.MarblesMoved++
	}
	// A red marble was pushed off the board
	if !gs.isInBounds(x, y) && tmp == marbleRed {
		gs.agents[gs.lastSnapshot().whoseTurn].score++
	}
	// Check for ko
	gs.ko = nil
	if !gs.isInBounds(x, y) {
		x -= move.dx()
		y -= move.dy()
	}
  otherAgentMarble := gs.lastSnapshot().whoseTurn.otherAgent().marble()
	if nextSnapshot.board[y][x] == otherAgentMarble {
		gs.ko = &Move{
			X: x,
			Y: y,
			D: move.D.reverse(),
		}
	}

	if gs.firstMoveTimer != nil {
		if !gs.firstMoveTimer.Stop() {
			panic("Ending first move timer failed!")
		}
		gs.firstMoveTimer = nil
		gs.firstMoveDeadline = nil
	}

	if !gs.agents[gs.lastSnapshot().whoseTurn].endTurn() {
		panic("End player turn failed!")
	}

	gs.history = append(gs.history, nextSnapshot)
	gs.posToCount[gs.getPositionString()]++

	gs.updateStatus()
	if gs.status == statusOngoing {
		agent := gs.agents[gs.lastSnapshot().whoseTurn]
		if !agent.startTurn(gs.playerTimeoutCallback) {
			panic("startTurn failed!")
		}
	}
	return nil
}

func (gs *gameState) playerTimeoutCallback() {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	// The other team just won
	gs.status = gs.lastSnapshot().whoseTurn.otherAgent().winStatus()

	gs.updateStatus()

	// Notify front-end of update
	if gs.onAsyncUpdate != nil {
		gs.onAsyncUpdate()
	}
	if gs.onGameOver != nil {
		gs.onGameOver()
	}
}

func (gs *gameState) firstMoveTimeoutCallback() {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	// The other team just won
	gs.status = statusAborted
	gs.firstMoveDeadline = nil

	gs.teardown()

	// Notify front-end of update
	if gs.onAsyncUpdate != nil {
		gs.onAsyncUpdate()
	}
	if gs.onGameOver != nil {
		gs.onGameOver()
	}
}

func (gs *gameState) resign(agent AgentColor) bool {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	if gs.status != statusOngoing {
		return false
	}
	gs.status = agent.otherAgent().winStatus()

	if gs.onGameOver != nil {
		gs.onGameOver()
	}

  gs.teardown()
	return true
}

// Bring the game to a completely inert state-- ensure no timers are gonna fire
// or anything like that.
func (gs *gameState) teardown() {
	if gs.firstMoveTimer != nil {
		gs.firstMoveTimer.Stop()
		gs.firstMoveTimer = nil
	}
	for _, a := range gs.agents {
		a.endTurn()
	}
}

func (gs *gameState) getValidMoves() []MoveWMarblesMoved {
	var moves []MoveWMarblesMoved
	for x := 0; x < gs.boardsize(); x++ {
		for y := 0; y < gs.boardsize(); y++ {
			if gs.lastSnapshot().board[y][x] != gs.lastSnapshot().whoseTurn.marble() {
				continue
			}
			for _, dir := range []Direction{DirUp, DirDown, DirLeft, DirRight} {
				move := Move{X: x, Y: y, D: dir}
				if fullMove, err := gs.ValidateMove(move); err == nil {
					moves = append(moves, *fullMove)
				}
			}
		}
	}
	return moves
}
