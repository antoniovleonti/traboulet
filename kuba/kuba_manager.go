package kuba

import (
  "errors"
	"net/http"
	"crypto/rand"
  "encoding/base64"
  "time"
  "encoding/json"
)

type User struct {
  id string
  cookie *http.Cookie
  color AgentColor
}

type clientViewPlayer struct {
  ID string `json:"id"`
  TimeNs int64 `json:"time_ns"`
  Deadline *time.Time `json:"deadline,omitempty"`
}

type ClientView struct {
  Board BoardT `json:"board"`
  Status string `json:"status"`
  Ko *Move `json:"ko"`
  LastMove *LastMoveT `json:"last_move"`
  WhoseTurn string `json:"whose_turn"` // color of current player
  WinThreshold int `json:"win_threshold"`
  ClockEnabled bool `json:"clock_enableed"`
  ColorToPlayer map[string]clientViewPlayer `json:"color_to_player"`
}

type KubaManager struct {
  state *kubaGame
  asyncCh chan struct{}
  cookieToUser map[string]*User
  colorToUser map[AgentColor]*User
  path string
}

func NewKubaManager(path string, t time.Duration, idWhite, idBlack string) (
    *KubaManager) {
  km := &KubaManager {
    state: newKubaGame(t),
    asyncCh: make(chan struct{}),
    cookieToUser: make(map[string]*User),
    colorToUser: make(map[AgentColor]*User),
    path: path,
  }
  km.setUser(agentWhite, idWhite)
  km.setUser(agentBlack, idBlack)
  return km
}

func (km *KubaManager) newCookie(id string) *http.Cookie {
  c := new(http.Cookie)
  c.Name = id
  c.Value = getRandBase64String(32)
  c.Expires = time.Now().Add(24 * time.Hour)
  c.Path = km.path
  return c
}

// For tests
func (km *KubaManager) GetWhiteCookie() *http.Cookie {
  return km.colorToUser[agentWhite].cookie
}

func (km *KubaManager) GetBlackCookie() *http.Cookie {
  return km.colorToUser[agentBlack].cookie
}

func getRandBase64String(length int) string {
  randomBytes := make([]byte, length)
  _, err := rand.Read(randomBytes)
  if err != nil {
    panic(err)
  }
  return base64.StdEncoding.EncodeToString(randomBytes)[:length]
}

func (km *KubaManager) setUser(color AgentColor, id string) bool {
  u := User {
    id: id,
    cookie: km.newCookie(id),
    color: color,
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

func (km *KubaManager) tryResign(c *http.Cookie) bool {
  user, ok := km.cookieToUser[c.Value]
  if !ok {
    return false
  }
  return km.state.resign(user.color)
}

func (km KubaManager) GetClientView() ClientView {
  colorToPlayer := make(map[string]clientViewPlayer)
  for k, v := range km.colorToUser {
    agent := km.state.agents[k]
    colorToPlayer[k.String()] = clientViewPlayer{
      ID: v.id,
      TimeNs: agent.time.Nanoseconds(),
      Deadline: agent.deadline,
    }
  }

  return ClientView {
    Board: km.state.board,
    Status: km.state.status.String(),
    Ko: km.state.ko,
    LastMove: km.state.lastMove,
    WhoseTurn: km.state.whoseTurn.String(),
    WinThreshold: km.state.winThreshold,
    ClockEnabled: km.state.clockEnabled,
    ColorToPlayer: colorToPlayer,
  }
}

func (km KubaManager) MarshalJSON() ([]byte, error) {
  return json.Marshal(km.GetClientView())
}
