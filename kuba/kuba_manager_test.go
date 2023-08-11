package kuba

import (
  "reflect"
  "encoding/json"
  "testing"
  "time"
  // "fmt"
)

func TestGetRandBase64String(t *testing.T) {
  if len(getRandBase64String(32)) != 32 {
    t.Error("length did not match expectation")
  }
}

func TestNewKubaManager(t *testing.T) {
  path := "testpath"
  km := NewKubaManager(path, 0 * time.Second, "white", "black")

  if len(km.cookieToUser) != 2 {
    t.Error("num users did not match expectation")
  }

  if km.path != path {
    t.Error("path did not match")
  }

  used := make(map[AgentColor]bool)
  for _, v := range km.cookieToUser {
    used[v.color] = true
  }

  if !(used[agentWhite] && used[agentBlack]) {
    t.Error("did not assign players to both colors!")
  }
}

func TestTryMove(t *testing.T) {
  km := NewKubaManager("path", 0 * time.Second, "white", "black")

  type testCase struct {
    m Move
    a AgentColor
    valid bool
  }
  testCases := []testCase{
    testCase{ m: Move{ x: 0, y: 0, d: DirDown }, a: agentWhite, valid: true },
    testCase{ m: Move{ x: 0, y: 0, d: DirDown }, a: agentWhite, valid: false },
    testCase{ m: Move{ x: 0, y: 1, d: DirDown }, a: agentWhite, valid: false },
    testCase{ m: Move{ x: 6, y: 0, d: DirDown }, a: agentBlack, valid: true },
    testCase{ m: Move{ x: 0, y: 1, d: DirDown }, a: agentWhite, valid: true },
  }
  for idx, tc := range testCases {
    actual := km.tryMove(tc.m, km.colorToUser[tc.a].cookie)
    if actual != tc.valid {
      t.Errorf("testCases[%d]: expected %t, got %t\n", idx, tc.valid, actual)
    }
  }
}

func TestMarshalJSON(t *testing.T) {
  km := NewKubaManager("path", 600 * time.Second, "white", "black")
  b, err := json.Marshal(km)
  if err != nil {
    t.Errorf("marshal json error: %s", err.Error())
  }
  var actual ClientView
  err = json.Unmarshal(b, &actual)
  if err != nil {
    t.Fatal(err)
  }
  expected := km.GetClientView()
	if !reflect.DeepEqual(actual, expected) {
    t.Errorf("handler returned unexpected body:\ngot: %v\nexpected: %v\n",
             actual, expected)
	}
}
