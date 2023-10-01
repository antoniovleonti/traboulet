package game

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
  "sync"
)

type User struct {
	cookie *http.Cookie
	color  AgentColor
  wantsRematch bool
}

type clientViewPlayer struct {
	TimeNs   int64      `json:"timeNs"`
	Deadline *time.Time `json:"deadline"`
	ID       string     `json:"id"`
	Color    string     `json:"color"`
	Score    int        `json:"score"`
  WantsRematch bool `json:"wantsRematch"`
}

type ClientView struct {
	History           []snapshot                  `json:"history"`
	Status            string                      `json:"status"`
	WinThreshold      int                         `json:"winThreshold"`
	ColorToPlayer     map[string]clientViewPlayer `json:"colorToPlayer"`
	IDToPlayer        map[string]clientViewPlayer `json:"idToPlayer"`
	ValidMoves        []MoveWMarblesMoved         `json:"validMoves"`
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
  // only used for making a rematch
  config Config
  onAsyncUpdate func()
  onGameOver func()
  onRematch func()

  mutex sync.RWMutex
}

func NewGameManager(
	config Config, white, black *http.Cookie, onAsyncUpdate func(),
	onGameOver func(), onRematch func()) (*GameManager, error) {
	if white == nil {
		return nil, errors.New("Missing white cookie")
	}
	if black == nil {
		return nil, errors.New("Missing black cookie")
	}
	state, err := newGameState(config, onAsyncUpdate, onGameOver, 1*time.Minute)
	if err != nil {
		return nil, err
	}
	gm := &GameManager{
		state:        state,
		cookieToUser: make(map[string]*User),
		colorToUser:  make(map[AgentColor]*User),
    config: config,
    onAsyncUpdate: onAsyncUpdate,
    onGameOver: onGameOver,
    onRematch: onRematch,
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
  gm.mutex.RLock()
  defer gm.mutex.RUnlock()

	user, ok := gm.cookieToUser[getKeyFromCookie(c)]
	if !ok {
		return errors.New("Cookie not found.")
	}
	if user.color != gm.state.lastSnapshot().whoseTurn {
		return errors.New("It is not your turn.")
	}
	if err := gm.state.ExecuteMove(m); err != nil {
		return err
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

func (gm *GameManager) OfferRematch(c *http.Cookie) (bool, error) {
  err := gm.RematchOfferErrCheck(c)
  if err != nil {
    return false, err
  }

  gm.mutex.Lock()
  defer gm.mutex.Unlock()

	user := gm.cookieToUser[getKeyFromCookie(c)]
  user.wantsRematch = true

  if !gm.colorToUser[user.color.otherAgent()].wantsRematch {
    // Do nothing, don't start a new game.
    return false, nil
  }

  if gm.onRematch != nil {
    gm.onRematch()
  }

  // Swap colors
  oldWhiteCookie := gm.colorToUser[agentWhite].cookie
  oldBlackCookie := gm.colorToUser[agentBlack].cookie
  gm.setUser(agentWhite, oldBlackCookie)
  gm.setUser(agentBlack, oldWhiteCookie)

  state, err :=
    newGameState(gm.config, gm.onAsyncUpdate, gm.onGameOver, 1*time.Minute)
  if err != nil {
    panic("couldn't make a rematch game! "+err.Error())
  }

  gm.state = state

  return true, nil
}

// In its own function to make the mutex easier to manage
func (gm *GameManager) RematchOfferErrCheck(c *http.Cookie) error {
  gm.mutex.RLock()
  defer gm.mutex.RUnlock()

  if gm.state.status == statusOngoing {
    return errors.New("Current game is not over.")
  }
	user, ok := gm.cookieToUser[getKeyFromCookie(c)]
  if !ok {
    return errors.New("Cookie not found.")
  }
  if user.wantsRematch {
    return errors.New("Rematch already requested.")
  }
  return nil
}

func (gm GameManager) GetClientView() ClientView {
  gm.mutex.RLock()
  defer gm.mutex.RUnlock()

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
      WantsRematch: user.wantsRematch,
		}

		colorToPlayer[color.String()] = player
		idToPlayer[user.cookie.Name] = player
	}

	return ClientView{
		History:           gm.state.history,
		Status:            gm.state.status.String(),
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
