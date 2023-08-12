package server

import (
  "bytes"
  "strings"
  "encoding/json"
  "kuba"
  "net/http"
  "net/http/httptest"
  "reflect"
  "testing"
  "time"
)

func TestNewGameHandler(t *testing.T) {
  // Make sure gameHandler implements the http.Handler interface
  var _ http.Handler = (*gameHandler)(nil)

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

func TestPostValidMove(t *testing.T) {
  gh := newGameHandler("", 0 * time.Second, "white", "black")
  if gh == nil {
    t.Error("game handler is nil")
  }

  // Create the body
  move := kuba.Move{ X: 0, Y: 0, D: kuba.DirRight }
  b, err := json.Marshal(move)
  if err != nil {
    t.Fatal(err)
  }
  body := bytes.NewReader(b)

  // Build request
  req, err := http.NewRequest("POST", "/move", body)
  if err != nil {
    t.Fatal(err)
  }
  req.AddCookie(gh.km.GetWhiteCookie())

  // Run the request through our handler
  rr := httptest.NewRecorder()
  gh.ServeHTTP(rr, req)

  // Check the response body is what we expect.
	if rr.Code != http.StatusOK {
    t.Errorf("handler returned unexpected status:\ngot: %d\nexpected: %d\n",
             rr.Code, http.StatusOK)
    t.Errorf("returned body: %s\n", rr.Body.String())
	}
}

func TestPostInvalidMove(t *testing.T) {
  gh := newGameHandler("", 0 * time.Second, "white", "black")
  if gh == nil {
    t.Error("game handler is nil")
  }

  // Build request
  req, err := http.NewRequest("POST", "/move", strings.NewReader("blah blah"))
  if err != nil {
    t.Fatal(err)
  }
  req.AddCookie(gh.km.GetWhiteCookie())

  // Run the request through our handler
  rr := httptest.NewRecorder()
  gh.ServeHTTP(rr, req)

  // Check the response body is what we expect.
	if rr.Code != http.StatusBadRequest {
    t.Errorf("handler returned unexpected status:\ngot: %d\nexpected: %d\n",
             rr.Code, http.StatusBadRequest)
    t.Errorf("returned body: %s\n", rr.Body.String())
	}
}

func TestPostMoveNoCookie(t *testing.T) {
  gh := newGameHandler("", 0 * time.Second, "white", "black")
  if gh == nil {
    t.Error("game handler is nil")
  }

  // Create the body
  move := kuba.Move{ X: 0, Y: 0, D: kuba.DirRight }
  b, err := json.Marshal(move)
  if err != nil {
    t.Fatal(err)
  }
  body := bytes.NewReader(b)

  // Build request
  req, err := http.NewRequest("POST", "/move", body)
  if err != nil {
    t.Fatal(err)
  }

  // Run the request through our handler
  rr := httptest.NewRecorder()
  gh.ServeHTTP(rr, req)

  // Check the response body is what we expect.
	if rr.Code != http.StatusUnauthorized {
    t.Errorf("handler returned unexpected status:\ngot: %d\nexpected: %d\n",
             rr.Code, http.StatusUnauthorized)
    t.Errorf("returned body: %s\n", rr.Body.String())
	}
}
