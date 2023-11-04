package server

import (
	"encoding/json"
	"game"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"sync"
	"time"
  "evtpub"
  "net/url"
)

// Assumption is that this is a function which will create a new game with
// these cookies and config, publish a redirect to the new game (or a signal to
// refresh or whatever-- details of what is communicated to the client is kind
// of none of this file's business), and probably dispose of this handler.
type challengeAcceptedCb func(
	*challengeHandler, game.Config, *http.Cookie, *http.Cookie) (*url.URL, error)

type challengeHandler struct {
	router              *httprouter.Router
	creator             *http.Cookie
	timestamp           time.Time
	config              game.Config
	onChallengeAccepted challengeAcceptedCb
	mutex               sync.RWMutex
	accepted            bool
	gamePath            *url.URL
  channelPub          *evtpub.ChannelPublisher
}

type challengeHandlerView struct {
	Config    game.Config `json:"config"`
	CreatorID string      `json:"creatorID"`
}

func newChallengeHandler(
	c *http.Cookie,
  config game.Config,
	onChallengeAccepted challengeAcceptedCb,
  channelPub *evtpub.ChannelPublisher,
)(
  *challengeHandler, error,
){
	if err := config.Validate(); err != nil {
		return nil, err
	}
	ch := challengeHandler{
		router:              httprouter.New(),
		timestamp:           time.Now(),
		creator:             c,
		config:              config,
		onChallengeAccepted: onChallengeAccepted,
		accepted:            false,
    channelPub:            channelPub,
	}

	ch.router.GET("/", ch.getChallenge)
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
		w.Header().Add("Location", ch.gamePath.String())
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

	gamePath, err := ch.onChallengeAccepted(ch, ch.config, ch.creator, c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ch.accepted = true
	ch.gamePath = gamePath

  // Respond to this request.
	w.Header().Add("Location", gamePath.String())
	w.WriteHeader(http.StatusSeeOther)
	w.Write([]byte("Success; check header Location field for game path."))

  err = ch.channelPub.Push("game-created", gamePath.String())
}

func (ch *challengeHandler) MarshalJSON() ([]byte, error) {
	return json.Marshal(challengeHandlerView{
		Config:    ch.config,
		CreatorID: ch.creator.Name,
	})
}

func (ch *challengeHandler) TearDown() {
  ch.channelPub.Delete()
}
