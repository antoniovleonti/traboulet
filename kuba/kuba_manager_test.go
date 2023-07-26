package kuba

import (
  "testing"
  "time"
)

func TestGetRandBase64String(t *testing.T) {
  if len(getRandBase64String(32)) != 32 {
    t.Error("length did not match expectation")
  }
}

func TestNewUser(t *testing.T) {
  for i := 0; i < 1000; i++ {
    km := newKubaManager(0 * time.Second, "abcdefg")
    if !km.newUser("p1") {
      t.Error("couldn't create first player")
    }
    if !km.newUser("p2") {
      t.Error("couldn't create second player")
    }

    if len(km.cookieToUser) != 2 {
      t.Error("num users did not match expectation")
    }

    used := make(map[AgentColor]bool)
    for _, v := range km.cookieToUser {
      used[v.color] = true
    }

    if !(used[agentWhite] && used[agentBlack]) {
      t.Error("did not assign players to both colors!")
    }
  }
}
