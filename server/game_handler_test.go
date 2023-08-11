package server

import (
  "reflect"
  "kuba"
  "net/http/httptest"
  "encoding/json"
  "testing"
  "net/http"
  "time"
)

func TestNewGameHandler(t *testing.T) {
  gh := newGameHandler("", 0 * time.Second, "white", "black")
  if gh == nil {
    t.Error("game handler is nil")
  }
}

func TestGetState(t *testing.T) {
  gh := newGameHandler("", 0 * time.Second, "white", "black")
  if gh == nil {
    t.Error("game handler is nil")
  }

  req, err := http.NewRequest("GET", "/state", nil)
  if err != nil {
    t.Fatal(err)
  }
  rr := httptest.NewRecorder()
  // Our handlers satisfy http.Handler, so we can call their ServeHTTP method
  // directly and pass in our Request and ResponseRecorder.
  gh.ServeHTTP(rr, req)
  // Check the response body is what we expect.
  var actual kuba.ClientView
  decoder := json.NewDecoder(rr.Body)
  err = decoder.Decode(&actual)
  if err != nil {
    t.Fatal(err)
  }
  expected := gh.km.GetClientView()
	if !reflect.DeepEqual(actual, expected) {
    t.Errorf("handler returned unexpected body:\ngot: %v\nexpected: %v\n",
             actual, expected)
	}
}
