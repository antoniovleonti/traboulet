package server

import (
  "github.com/julienschmidt/httprouter"
  "kuba"
  "net/http"
  "time"
  "encoding/json"
)

// Handles requests for endpoints related to a particular game. Is not aware
// that any other games exist (accepts requests for endpoints like "/state/").
// This allows for things like writing a single-game server for testing.
type gameHandler struct {
  km *kuba.KubaManager
  router *httprouter.Router
  statePublisher publisher
  path string
}

func newGameHandler(path string, t time.Duration, idWhite, idBlack string) (
    *gameHandler) {
  gh := gameHandler {
    km: kuba.NewKubaManager(path, t, idWhite, idBlack),
    router: httprouter.New(),
    statePublisher: publisher{},
    path: path,
  }
  gh.router.GET("/state", gh.state)

  return &gh
}

// Convenience method
func (gh *gameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  gh.router.ServeHTTP(w, r)
}

func (gh *gameHandler) state(w http.ResponseWriter, r *http.Request,
                             p httprouter.Params) {
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(gh.km)
}
