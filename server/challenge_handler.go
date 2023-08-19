package server

import (
	"github.com/julienschmidt/httprouter"
	"kuba"
	"net/http"
	"sync"
)

// Assumption is that this is a function which will create a new game with
// these cookies and config, publish a redirect to the new game (or a signal to
// refresh or whatever-- details of what is communicated to the client is kind
// of none of this file's business), and probably dispose of this handler.
type challengeAcceptedCb func(
	*publisher, kuba.Config, *http.Cookie, *http.Cookie)

type challengeHandler struct {
	router   *httprouter.Router
	creator  *http.Cookie
	config   kuba.Config
	pub      publisher
	acceptCb challengeAcceptedCb
	accepted bool
	mutex    sync.Mutex
}

func newChallengeHandler(
	c *http.Cookie, config kuba.Config,
	acceptCb challengeAcceptedCb) *challengeHandler {
	ch := challengeHandler{
		router:   httprouter.New(),
		creator:  c,
		config:   config,
		pub:      publisher{},
		acceptCb: acceptCb,
		accepted: false,
	}

	ch.router.GET("/update", ch.getUpdate)
	ch.router.POST("/join", ch.postJoin)

	return &ch
}

func (ch *challengeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ch.router.ServeHTTP(w, r)
}

func (ch *challengeHandler) getUpdate(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ch.pub.addSubscriber(w)
}

func (ch *challengeHandler) postJoin(
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
	if c.Value == ch.creator.Value {
		http.Error(
			w, "You cannot accept your own challenge.", http.StatusBadRequest)
		return
	}

	w.Write([]byte("success"))
	if ch.acceptCb != nil {
		ch.acceptCb(&ch.pub, ch.config, ch.creator, c)
	}
	ch.accepted = true
}
