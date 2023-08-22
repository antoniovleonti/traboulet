package server

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"kuba"
	"net/http"
	"net/url"
)

type createGameFnT func(kuba.Config, *http.Cookie, *http.Cookie) string

type challengeRouter struct {
	router     *httprouter.Router
	challenges map[string]*challengeHandler
	pathGen    *pathGenerator
	createGame createGameFnT
	prefix     string
}

func newChallengeRouter(
	prefix string, createGame createGameFnT) *challengeRouter {
	if prefix[0] != '/' || prefix[len(prefix)-1] != '/' {
		panic("expect prefix to begin and end with '/'")
	}
	cr := challengeRouter{
		router:     httprouter.New(),
		challenges: make(map[string]*challengeHandler),
		pathGen:    newPathGenerator(),
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
	id := p.ByName("id")
	etc := p.ByName("etc")

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
	// create challenge
	path := cr.pathGen.newPath(8)
	bind := func(c kuba.Config, c1, c2 *http.Cookie) (string, error) {
		return cr.onChallengeAccepted(path, c, c1, c2)
	}
	cr.challenges[path] = newChallengeHandler(cookie, config, bind)

	w.Header().Add("Location", cr.prefix+path)
	w.WriteHeader(http.StatusSeeOther)
	w.Write([]byte("success"))
}

func (cr *challengeRouter) getChallenges(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b, err := json.Marshal(cr.challenges)
	if err != nil {
		http.Error(
			w, "could not marshal challenges: "+err.Error(),
			http.StatusInternalServerError)
	}
	w.Write(b)
}

func (cr *challengeRouter) onChallengeAccepted(
	id string, config kuba.Config, cookie1,
	cookie2 *http.Cookie) (string, error) {
	if _, ok := cr.challenges[id]; !ok {
		// ???? challenge doesn't exist!
		return "", errors.New("Challenge " + id + " (unexpectedly) does not exist.")
	}

	if cr.createGame == nil {
		return "", errors.New(
			"Game could be created, but no callback to do so was provided.")
	}

	delete(cr.challenges, id)
	return cr.createGame(config, cookie1, cookie2), nil
}
