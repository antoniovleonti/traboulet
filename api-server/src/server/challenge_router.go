package server

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"kuba"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type deleteChallengeFn func()

type createGameFnT func(
	deleteChallengeFn, kuba.Config, *http.Cookie, *http.Cookie) (string, error)

type challengeRouter struct {
	router     *httprouter.Router
	challenges map[string]*challengeHandler
	pathGen    *nonCryptoStringGen
	createGame createGameFnT
	prefix     string
	mutex      sync.RWMutex
}

func newChallengeRouter(
	prefix string, createGame createGameFnT) *challengeRouter {
	if prefix[0] != '/' || prefix[len(prefix)-1] != '/' {
		panic("expect prefix to begin and end with '/'")
	}
	cr := challengeRouter{
		router:     httprouter.New(),
		challenges: make(map[string]*challengeHandler),
		pathGen:    newNonCryptoStringGen(),
		createGame: createGame,
		prefix:     prefix,
	}

	cr.router.GET("/", cr.getChallenges)
	cr.router.POST("/", cr.postChallenge)

	cr.router.GET("/:id", cr.forwardToHandler)
	cr.router.GET("/:id/*etc", cr.forwardToHandler)
	cr.router.POST("/:id/*etc", cr.forwardToHandler)

	return &cr
}

func (cr *challengeRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If request has come in with no cookie, set cookie in the response.
	cr.router.ServeHTTP(w, r)
}

func (cr *challengeRouter) forwardToHandler(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	id := p.ByName("id")
	etc := httprouter.CleanPath(p.ByName("etc"))

	url, err := url.Parse(etc)
	if err != nil {
		http.Error(w, "error parsing URL", http.StatusInternalServerError)
		return
	}
	r.URL = url

	if ch, ok := cr.challenges[id]; ok && ch != nil {
		ch.ServeHTTP(w, r)
		return
	}

	http.Error(w, "challenge "+id+" not found.", http.StatusNotFound)
	return
}

func (cr *challengeRouter) postChallenge(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// check cookie
	if len(r.Cookies()) == 0 {
		http.Error(w, "No cookie found.", http.StatusUnauthorized)
	}
	cookie := r.Cookies()[0]
	// parse body
	var config kuba.Config
	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		http.Error(w, "Could not parse config: "+err.Error(), http.StatusBadRequest)
		return
	}

	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	// create challenge
	path := cr.pathGen.newString(8)
	bind := func(
		ch *challengeHandler, c kuba.Config, c1, c2 *http.Cookie) (
		string, error) {
		return cr.onChallengeAccepted(ch, path, c, c1, c2)
	}
	challenge, err := newChallengeHandler(cookie, config, bind)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cr.challenges[path] = challenge

	log.Print("Created challenge " + path + ".")

	w.Header().Add("Location", cr.prefix+path)
	w.WriteHeader(http.StatusSeeOther)
	w.Write([]byte("success"))
}

func (cr *challengeRouter) getChallenges(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

  unaccepted := make(map[string]*challengeHandler)
  for id, ch := range cr.challenges {
    if !ch.accepted {
      unaccepted[id] = ch
    }
  }
	b, err := json.Marshal(unaccepted)
	if err != nil {
		http.Error(
			w, "could not marshal challenges: "+err.Error(),
			http.StatusInternalServerError)
	}
	w.Write(b)
}

// Not thread safe-- MUST be synchronized by challenge handler
func (cr *challengeRouter) onChallengeAccepted(
	ch *challengeHandler, id string, config kuba.Config, cookie1,
	cookie2 *http.Cookie) (string, error) {
	if _, ok := cr.challenges[id]; !ok {
		// ???? challenge doesn't exist!
		return "", errors.New("Challenge " + id + " (unexpectedly) does not exist.")
	}

	if cr.createGame == nil {
		return "", errors.New(
			"Game could be created, but no callback to do so was provided.")
	}

	deleteCb := func() {
		delete(cr.challenges, id)
	}

	return cr.createGame(deleteCb, config, cookie1, cookie2)
}
