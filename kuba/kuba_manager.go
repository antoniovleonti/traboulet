package kuba

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
type KubaManager struct {
	state        *kubaGame
	cookieToUser map[string]*User // "string" key is serialized cookie
	colorToUser  map[AgentColor]*User
}

func NewKubaManager(
	config Config, white, black *http.Cookie, onAsyncUpdate func(),
	onGameOver func()) (*KubaManager, error) {
	if white == nil {
		return nil, errors.New("Missing white cookie")
	}
	if black == nil {
		return nil, errors.New("Missing black cookie")
	}
	state, err := newKubaGame(config, onAsyncUpdate, onGameOver, 10*time.Minute)
	if err != nil {
		return nil, err
	}
	km := &KubaManager{
		state:        state,
		cookieToUser: make(map[string]*User),
		colorToUser:  make(map[AgentColor]*User),
	}
	km.setUser(agentWhite, white)
	km.setUser(agentBlack, black)
	return km, nil
}

// For tests
func (km *KubaManager) GetWhiteCookie() *http.Cookie {
	return km.colorToUser[agentWhite].cookie
}

func (km *KubaManager) GetBlackCookie() *http.Cookie {
	return km.colorToUser[agentBlack].cookie
}

func (km *KubaManager) setUser(color AgentColor, cookie *http.Cookie) bool {
	u := User{
		cookie: cookie,
		color:  color,
	}
	km.cookieToUser[getKeyFromCookie(cookie)] = &u
	km.colorToUser[color] = &u
	return true
}

func (km *KubaManager) TryMove(m Move, c *http.Cookie) error {
	user, ok := km.cookieToUser[getKeyFromCookie(c)]
	if !ok {
		return errors.New("Cookie not found.")
	}
	if user.color != km.state.whoseTurn {
		return errors.New("It is not your turn.")
	}
	if !km.state.ExecuteMove(m) {
		return errors.New("Move was invalid.")
	}
	return nil
}

func (km *KubaManager) TryResign(c *http.Cookie) bool {
	user, ok := km.cookieToUser[getKeyFromCookie(c)]
	if !ok {
		return false
	}
	return km.state.resign(user.color)
}

func (km KubaManager) GetClientView() ClientView {
	colorToPlayer := make(map[string]clientViewPlayer)
	idToPlayer := make(map[string]clientViewPlayer)
	for color, user := range km.colorToUser {
		agent := km.state.agents[color]
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
		Board:             km.state.board,
		Status:            km.state.status.String(),
		LastMove:          km.state.lastMove,
		WhoseTurn:         km.state.whoseTurn.String(),
		WinThreshold:      km.state.winThreshold,
		ColorToPlayer:     colorToPlayer,
		IDToPlayer:        idToPlayer,
		ValidMoves:        km.state.validMoves,
		FirstMoveDeadline: km.state.firstMoveDeadline,
    TimeControl:       km.state.timeControl,
	}
}

func (km KubaManager) MarshalJSON() ([]byte, error) {
	return json.Marshal(km.GetClientView())
}

// This way we don't have to worry about what fields the client is / is not
// sending. It will always be the same regardless of client / golang behavior.
func getKeyFromCookie(c *http.Cookie) string {
	return c.Name + "=" + c.Value
}
