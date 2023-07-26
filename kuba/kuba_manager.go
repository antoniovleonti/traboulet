package kuba

import (
	"net/http"
	crand "crypto/rand"
  "encoding/base64"
  mrand "math/rand"
  "time"
)

type User struct {
  id string
  cookie *http.Cookie
  color AgentColor
}

type KubaManager struct {
  state *kubaGame
  asyncCh chan struct{}
  cookieToUser map[*http.Cookie]*User
  path string
}

func newKubaManager(t time.Duration, path string) *KubaManager {
  return &KubaManager {
    state: newKubaGame(t),
    asyncCh: make(chan struct{}),
    cookieToUser: make(map[*http.Cookie]*User),
    path: path,
  }
}

func (km *KubaManager) newCookie(id string) *http.Cookie {
  c := new(http.Cookie)
  c.Name = id
  c.Value = getRandBase64String(32)
  c.Expires = time.Now().Add(24 * time.Hour)
  c.Path = km.path
  return c
}

func getRandBase64String(length int) string {
  randomBytes := make([]byte, length)
  _, err := crand.Read(randomBytes)
  if err != nil {
    panic(err)
  }
  return base64.StdEncoding.EncodeToString(randomBytes)[:length]
}

func (km *KubaManager) newUser(id string) bool {
  if len(km.cookieToUser) > 1 {
    // Sorry, the game is full!
    return false
  }

  var color AgentColor
  if len(km.cookieToUser) == 0 {
    color = AgentColor(mrand.Intn(2) + 1)
  } else /* => len(...) == 1 */ {
    for _, v := range km.cookieToUser {
      color = v.color.otherAgent()
    }
  }

  u := User {
    id: id,
    cookie: km.newCookie(id),
    color: color,
  }
  km.cookieToUser[u.cookie] = &u
  return true
}
