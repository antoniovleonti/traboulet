package game

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type User struct {
	cookie *http.Cookie
	color  AgentColor
}

type clientViewPlayer struct {
	TimeNs   int64      `json:"timeNs"`
	Deadline *time.Time `json:"deadline"`
	ID       string     `json:"id"`
	Color    string     `json:"color"`
	Score    int        `json:"score"`
}

type ClientView struct {
	Board             BoardT                      `json:"board"`
	Status            string                      `json:"status"`
	LastMove          *LastMoveT                  `json:"lastMove"`
	WhoseTurn         string                      `json:"whoseTurn"` // color of current player
	WinThreshold      int                         `json:"winThreshold"`
	ColorToPlayer     map[string]clientViewPlayer `json:"colorToPlayer"`
	IDToPlayer        map[string]clientViewPlayer `json:"idToPlayer"`
	ValidMoves        []Move                      `json:"validMoves"`
	FirstMoveDeadline *time.Time                  `json:"firstMoveDeadline"`
	TimeControl       time.Duration               `json:"timeControl"`
}

// Handles mapping cookie -> color (black / white) & ensuring players only move
// when it's their turn to do so (the state validator only checks that it is the
// correct marble color being moved, has no concept of "who" moved it).
type GameManager struct {
	state        *gameState
	cookieToUser map[string]*User // "string" key is serialized cookie
	colorToUser  map[AgentColor]*User
}

func NewGameManager(
	config Config, white, black *http.Cookie, onAsyncUpdate func(),
	onGameOver func()) (*GameManager, error) {
	if white == nil {
		return nil, errors.New("Missing white cookie")
	}
	if black == nil {
		return nil, errors.New("Missing black cookie")
	}
	state, err := newGameState(config, onAsyncUpdate, onGameOver, 10*time.Minute)
	if err != nil {
		return nil, err
	}
	gm := &GameManager{
		state:        state,
		cookieToUser: make(map[string]*User),
		colorToUser:  make(map[AgentColor]*User),
	}
	gm.setUser(agentWhite, white)
	gm.setUser(agentBlack, black)
	return gm, nil
}

// For tests
func (gm *GameManager) GetWhiteCookie() *http.Cookie {
	return gm.colorToUser[agentWhite].cookie
}

func (gm *GameManager) GetBlackCookie() *http.Cookie {
	return gm.colorToUser[agentBlack].cookie
}

func (gm *GameManager) setUser(color AgentColor, cookie *http.Cookie) bool {
	u := User{
		cookie: cookie,
		color:  color,
	}
	gm.cookieToUser[getKeyFromCookie(cookie)] = &u
	gm.colorToUser[color] = &u
	return true
}

func (gm *GameManager) TryMove(m Move, c *http.Cookie) error {
	user, ok := gm.cookieToUser[getKeyFromCookie(c)]
	if !ok {
		return errors.New("Cookie not found.")
	}
	if user.color != gm.state.whoseTurn {
		return errors.New("It is not your turn.")
	}
	if !gm.state.ExecuteMove(m) {
		return errors.New("Move was invalid.")
	}
	return nil
}

func (gm *GameManager) TryResign(c *http.Cookie) bool {
	user, ok := gm.cookieToUser[getKeyFromCookie(c)]
	if !ok {
		return false
	}
	return gm.state.resign(user.color)
}

func (gm GameManager) GetClientView() ClientView {
	colorToPlayer := make(map[string]clientViewPlayer)
	idToPlayer := make(map[string]clientViewPlayer)
	for color, user := range gm.colorToUser {
		agent := gm.state.agents[color]
		player := clientViewPlayer{
			TimeNs:   agent.time.Nanoseconds(),
			Deadline: agent.deadline,
			Color:    color.String(),
			ID:       user.cookie.Name,
			Score:    agent.score,
		}

		colorToPlayer[color.String()] = player
		idToPlayer[user.cookie.Name] = player
	}

	return ClientView{
		Board:             gm.state.board,
		Status:            gm.state.status.String(),
		LastMove:          gm.state.lastMove,
		WhoseTurn:         gm.state.whoseTurn.String(),
		WinThreshold:      gm.state.winThreshold,
		ColorToPlayer:     colorToPlayer,
		IDToPlayer:        idToPlayer,
		ValidMoves:        gm.state.validMoves,
		FirstMoveDeadline: gm.state.firstMoveDeadline,
		TimeControl:       gm.state.timeControl,
	}
}

func (gm GameManager) MarshalJSON() ([]byte, error) {
	return json.Marshal(gm.GetClientView())
}

// This way we don't have to worry about what fields the client is / is not
// sending. It will always be the same regardless of client / golang behavior.
func getKeyFromCookie(c *http.Cookie) string {
	return c.Name + "=" + c.Value
}
