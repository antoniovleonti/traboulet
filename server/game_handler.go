package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
  "github.com/antoniovleonti/sse"
	"kuba"
  "time"
  "log"
	"net/http"
  "sync"
)

// Handles requests for endpoints related to a particular game. Is not aware
// that any other games exist (accepts requests for endpoints like "/state/").
// This allows for things like writing a single-game server for testing.
type gameHandler struct {
	km     *kuba.KubaManager
	router *httprouter.Router
	pub    publisher
  completionTime *time.Time
  timeMutex sync.Mutex
}

func newGameHandler(
	config kuba.Config, white, black *http.Cookie) (*gameHandler, error) {
	gh := gameHandler{
		router: httprouter.New(),
		pub:    publisher{},
	}
	km, err :=
		kuba.NewKubaManager(config, white, black, gh.publishUpdate, gh.markComplete)
  if err != nil {
    return nil, err
  }
  gh.km = km

	// gh.router.Handler("GET", "/state-stream", gh.pub)
	gh.router.GET("/state-stream", gh.getStateStream)
	gh.router.GET("/state", gh.getState)
	gh.router.POST("/move", gh.postMove)
	gh.router.POST("/resignation", gh.postResignation)

	return &gh, nil
}

func (gh *gameHandler) publishUpdate() {
  b, err := json.Marshal(gh.km)
  if err != nil {
    panic("couldn't marshal game state")
  }
  event := sse.Event{
    Data: string(b),
  }

	gh.pub.do(func(w http.ResponseWriter) {
		err := event.Render(w)
    if err != nil {
      log.Printf("Error writing event: %v\n", err)
    }
  }, false)
}

// Convenience method
func (gh *gameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gh.router.ServeHTTP(w, r)
}

func (gh *gameHandler) getStateStream(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
  gh.pub.subscribe(w, r.Context().Done())
}

func (gh *gameHandler) getState(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gh.km)
}

func (gh *gameHandler) postMove(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Parse body.
	var move kuba.Move
	err := json.NewDecoder(r.Body).Decode(&move)
	if err != nil {
		http.Error(w, "Could not parse move: "+err.Error(), http.StatusBadRequest)
		return
	}

	c := r.Cookies()
	if len(c) == 0 {
		http.Error(w, "No cookies provided.", http.StatusUnauthorized)
		return
	}

	if err = gh.km.TryMove(move, c[0]); err != nil {
		http.Error(w, "Could not execute move: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	w.Write([]byte("success"))
	gh.publishUpdate()
}

func (gh *gameHandler) postResignation(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := r.Cookies()
	if len(c) == 0 {
		http.Error(w, "No cookies provided.", http.StatusUnauthorized)
		return
	}

	if !gh.km.TryResign(c[0]) {
		http.Error(w, "Could not resign.", http.StatusBadRequest)
		return
	}

	w.Write([]byte("success"))
	gh.publishUpdate()
}

func (gh *gameHandler) markComplete() {
  gh.timeMutex.Lock()
  defer gh.timeMutex.Unlock()
  t := time.Now()
  gh.completionTime = &t
}

func (gh *gameHandler) DurationSinceCompletion() *time.Duration {
  gh.timeMutex.Lock()
  defer gh.timeMutex.Unlock()
  if gh.completionTime == nil {
    return nil
  }
  d := time.Since(*gh.completionTime)
  return &d
}
