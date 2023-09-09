package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"kuba"
	"net/http"
	"sync"
	"time"
)

// Assumption is that this is a function which will create a new game with
// these cookies and config, publish a redirect to the new game (or a signal to
// refresh or whatever-- details of what is communicated to the client is kind
// of none of this file's business), and probably dispose of this handler.
type challengeAcceptedCb func(
	kuba.Config, *http.Cookie, *http.Cookie) (string, error)

type challengeHandler struct {
	router              *httprouter.Router
	creator             *http.Cookie
	timestamp           time.Time
	config              kuba.Config
	pub                 publisher
	onChallengeAccepted challengeAcceptedCb
	mutex               sync.RWMutex
	accepted            bool
	gamePath            *string
}

type challengeHandlerView struct {
	Config    kuba.Config `json:"config"`
	CreatorID string      `json:"creatorID"`
}

func newChallengeHandler(
	c *http.Cookie, config kuba.Config,
	onChallengeAccepted challengeAcceptedCb) (*challengeHandler, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	ch := challengeHandler{
		router:              httprouter.New(),
		timestamp:           time.Now(),
		creator:             c,
		config:              config,
		pub:                 publisher{},
		onChallengeAccepted: onChallengeAccepted,
		accepted:            false,
	}

	ch.router.GET("/", ch.getChallenge)
	ch.router.GET("/update", ch.getUpdate)
	ch.router.POST("/accept", ch.postAccept)

	return &ch, nil
}

func (ch *challengeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch.redirectToGameOrRoute(w, r)
}

func (ch *challengeHandler) redirectToGameOrRoute(
	w http.ResponseWriter, r *http.Request) {
	ch.mutex.RLock()
	if ch.accepted && ch.gamePath != nil {
		w.Header().Add("Location", *ch.gamePath)
		w.WriteHeader(http.StatusSeeOther)
		w.Write([]byte(
			"Game has started; check header Location field for game path."))

		ch.mutex.RUnlock()
		return
	}
	ch.mutex.RUnlock()
	ch.router.ServeHTTP(w, r)
}

func (ch *challengeHandler) getChallenge(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ch.mutex.RLock()
	defer ch.mutex.RUnlock()

	b, err := json.Marshal(ch)
	if err != nil {
		http.Error(w, "could not generate json: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func (ch *challengeHandler) getUpdate(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ch.pub.subscribe(w, r.Context().Done())
}

func (ch *challengeHandler) postAccept(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	if ch.accepted {
		http.Error(
			w, "Challenge has already been accepted.", http.StatusInternalServerError)
	}
	// This probably needs to be syncronized to make sure 2 ppl can't join at once
	// check cookie exists
	if len(r.Cookies()) == 0 {
		http.Error(w, "No cookies found.", http.StatusUnauthorized)
		return
	}
	c := r.Cookies()[0] // just grab first cookie, idc if they have multiple.
	// check cookie is not the same as the creator
	if c.Name == ch.creator.Name && c.Value == ch.creator.Value {
		http.Error(
			w, "You cannot accept your own challenge.", http.StatusBadRequest)
		return
	}

	if ch.onChallengeAccepted == nil {
		http.Error(
			w, "Challenge could be accepted, but no callback to do so was provided.",
			http.StatusInternalServerError)
		return
	}

	gamePath, err := ch.onChallengeAccepted(ch.config, ch.creator, c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ch.accepted = true
	ch.gamePath = &gamePath

	w.Header().Add("Location", gamePath)
	w.WriteHeader(http.StatusSeeOther)
	w.Write([]byte("Success; check header Location field for game path."))

	// Notify subscribers of new game
	ch.pub.do(func(w http.ResponseWriter) {
		w.Header().Add("Location", gamePath)
		w.WriteHeader(http.StatusSeeOther)
		w.Write([]byte(
			"Game has started; check header Location field for game path."))
	}, true)
}

func (ch *challengeHandler) MarshalJSON() ([]byte, error) {
	return json.Marshal(challengeHandlerView{
		Config:    ch.config,
		CreatorID: ch.creator.Name,
	})
}
