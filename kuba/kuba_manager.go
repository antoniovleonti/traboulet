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
	TimeNs   int64      `json:"time_ns"`
	Deadline *time.Time `json:"deadline,omitempty"`
}

type ClientView struct {
	Board         BoardT                      `json:"board"`
	Status        string                      `json:"status"`
	Ko            *Move                       `json:"ko"`
	LastMove      *LastMoveT                  `json:"last_move"`
	WhoseTurn     string                      `json:"whose_turn"` // color of current player
	WinThreshold  int                         `json:"win_threshold"`
	ClockEnabled  bool                        `json:"clock_enableed"`
	ColorToPlayer map[string]clientViewPlayer `json:"color_to_player"`
}

// Handles mapping cookie -> color (black / white) & ensuring players only move
// when it's their turn to do so (the state validator only checks that it is the
// correct marble color being moved, has no concept of "who" moved it).
type KubaManager struct {
	state        *kubaGame
	cookieToUser map[string]*User
	colorToUser  map[AgentColor]*User
}

func NewKubaManager(
	config Config, white, black *http.Cookie, onAsyncUpdate func(),
	onGameOver func()) *KubaManager {
	if white == nil {
		return nil
	}
	if black == nil {
		return nil
	}
	km := &KubaManager{
		state: newKubaGame(
			config, onAsyncUpdate, onGameOver, 30*time.Second),
		cookieToUser: make(map[string]*User),
		colorToUser:  make(map[AgentColor]*User),
	}
	km.setUser(agentWhite, white)
	km.setUser(agentBlack, black)
	return km
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
	km.cookieToUser[u.cookie.Value] = &u
	km.colorToUser[color] = &u
	return true
}

func (km *KubaManager) TryMove(m Move, c *http.Cookie) error {
	user, ok := km.cookieToUser[c.Value]
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
	user, ok := km.cookieToUser[c.Value]
	if !ok {
		return false
	}
	return km.state.resign(user.color)
}

func (km KubaManager) GetClientView() ClientView {
	colorToPlayer := make(map[string]clientViewPlayer)
	for k, _ := range km.colorToUser {
		agent := km.state.agents[k]
		colorToPlayer[k.String()] = clientViewPlayer{
			TimeNs:   agent.time.Nanoseconds(),
			Deadline: agent.deadline,
		}
	}

	return ClientView{
		Board:         km.state.board,
		Status:        km.state.status.String(),
		Ko:            km.state.ko,
		LastMove:      km.state.lastMove,
		WhoseTurn:     km.state.whoseTurn.String(),
		WinThreshold:  km.state.winThreshold,
		ClockEnabled:  km.state.clockEnabled,
		ColorToPlayer: colorToPlayer,
	}
}

func (km KubaManager) MarshalJSON() ([]byte, error) {
	return json.Marshal(km.GetClientView())
}
