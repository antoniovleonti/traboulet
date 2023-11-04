package server

import (
	"encoding/json"
	"game"
	"github.com/julienschmidt/httprouter"
  "evtpub"
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
	completionTime    *time.Time
	timeMutex         sync.Mutex
	deleteChallengeCb deleteChallengeFn
  channelPub evtpub.ChannelPublisher
}

func newGameHandler(
	deleteChallengeCb deleteChallengeFn,
  channelPub evtpub.ChannelPublisher,
  config game.Config,
  white, black *http.Cookie,
)(
  *gameHandler,
  error,
){
	gh := gameHandler{
		router:            httprouter.New(),
		deleteChallengeCb: deleteChallengeCb,
    channelPub: channelPub,
	}
	gm, err := game.NewGameManager(
    config, white, black, gh.publishUpdate, gh.markComplete,
    gh.undoMarkComplete)
	if err != nil {
		return nil, err
	}
	gh.gm = gm

	gh.router.GET("/state", gh.getState)
	gh.router.POST("/move", gh.postMove)
	gh.router.POST("/resignation", gh.postResignation)
	gh.router.POST("/rematch-offer", gh.postRematchOffer)

	return &gh, nil
}

func (gh *gameHandler) publishUpdate() {
	b, err := json.Marshal(gh.gm)
	if err != nil {
		panic("couldn't marshal game state")
	}

  gh.channelPub.Push("state-push", string(b))
}

// Convenience method
func (gh *gameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gh.router.ServeHTTP(w, r)
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

func (gh *gameHandler) postRematchOffer(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
  c := r.Cookies()

	if len(c) == 0 {
		http.Error(w, "No cookies provided.", http.StatusUnauthorized)
		return
	}

  _, err := gh.gm.OfferRematch(c[0])
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("success"))
  // Even if rematch was not started, we can notify other player that a rematch
  // was requested.
	gh.publishUpdate()
}

func (gh *gameHandler) markComplete() {
	gh.timeMutex.Lock()
	defer gh.timeMutex.Unlock()
	t := time.Now()
	gh.completionTime = &t
}

func (gh *gameHandler) undoMarkComplete() {
	gh.timeMutex.Lock()
	defer gh.timeMutex.Unlock()
	gh.completionTime = nil
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

func (gh *gameHandler) TearDown() {
	if gh.deleteChallengeCb != nil {
		gh.deleteChallengeCb()
	}
  gh.channelPub.Delete()
}
