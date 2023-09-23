package server

import (
	"encoding/json"
	"game"
	"github.com/antoniovleonti/sse"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"sync"
	"time"
)

// Handles requests for endpoints related to a particular game. Is not aware
// that any other games exist (accepts requests for endpoints like "/state/").
// This allows for things like writing a single-game server for testing.
type gameHandler struct {
	gm                *game.GameManager
	router            *httprouter.Router
	pub               publisher
	completionTime    *time.Time
	timeMutex         sync.Mutex
	deleteChallengeCb deleteChallengeFn
}

func newGameHandler(
	deleteChallengeCb deleteChallengeFn, config game.Config,
	white, black *http.Cookie) (*gameHandler, error) {
	gh := gameHandler{
		router:            httprouter.New(),
		pub:               publisher{},
		deleteChallengeCb: deleteChallengeCb,
	}
	gm, err :=
		game.NewGameManager(config, white, black, gh.publishUpdate, gh.markComplete)
	if err != nil {
		return nil, err
	}
	gh.gm = gm

	gh.router.GET("/event-stream", gh.getEventStream)
	gh.router.GET("/state", gh.getState)
	gh.router.POST("/move", gh.postMove)
	gh.router.POST("/resignation", gh.postResignation)

	go gh.periodicallySendKeepAlive()

	return &gh, nil
}

func (gh *gameHandler) publishUpdate() {
	b, err := json.Marshal(gh.gm)
	if err != nil {
		panic("couldn't marshal game state")
	}
	event := sse.Event{
		Event: "state-push",
		Data:  string(b),
	}

	gh.pub.do(func(w http.ResponseWriter) {
		err := event.Render(w)
		if err != nil {
			log.Printf("Error writing event: %v\n", err)
		}
	}, false)
}

func (gh *gameHandler) periodicallySendKeepAlive() {
	for {
		time.Sleep(1 * time.Minute)
		gh.sendKeepAlive()
	}
}

func (gh *gameHandler) sendKeepAlive() {
	event := sse.Event{
		Event: "keep-alive",
		Data:  "",
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

func (gh *gameHandler) getEventStream(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	gh.pub.subscribe(w, r.Context().Done())
}

func (gh *gameHandler) getState(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gh.gm)
}

func (gh *gameHandler) postMove(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Parse body.
	var move game.Move
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

	if err = gh.gm.TryMove(move, c[0]); err != nil {
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

	if !gh.gm.TryResign(c[0]) {
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

func (gh *gameHandler) DeleteChallenge() {
	if gh.deleteChallengeCb != nil {
		gh.deleteChallengeCb()
	}
}
