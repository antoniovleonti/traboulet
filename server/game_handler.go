package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"kuba"
	"net/http"
)

// Handles requests for endpoints related to a particular game. Is not aware
// that any other games exist (accepts requests for endpoints like "/state/").
// This allows for things like writing a single-game server for testing.
type gameHandler struct {
	km      *kuba.KubaManager
	router  *httprouter.Router
	pub     publisher
	asyncCh chan struct{}
}

func newGameHandler(
	config kuba.Config, white, black *http.Cookie) *gameHandler {
	gh := gameHandler{
		router: httprouter.New(),
		pub:    publisher{},
	}
	gh.km = kuba.NewKubaManager(config, white, black, gh.publishUpdate)

	gh.router.GET("/state", gh.getState)
	gh.router.GET("/update", gh.getGameUpdate)
	gh.router.POST("/move", gh.postMove)
	gh.router.POST("/resignation", gh.postResignation)

	return &gh
}

func (gh *gameHandler) publishUpdate() {
	b, err := json.Marshal(gh.km)
	if err != nil {
		panic("couldn't marshal game state")
	}
	gh.pub.publish(string(b))
}

// Convenience method
func (gh *gameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gh.router.ServeHTTP(w, r)
}

func (gh *gameHandler) getGameUpdate(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	gh.pub.addSubscriber(w)
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
