package server

import (
	"github.com/julienschmidt/httprouter"
	"kuba"
	mrand "math/rand"
	"net/http"
	"net/url"
)

type gameRouter struct {
	router     *httprouter.Router
	challenges map[string]*challengeHandler
	games      map[string]*gameHandler
	pathGen    *pathGenerator
	prefix     string
}

func newGameRouter(prefix string) *gameRouter {
	if prefix[0] != '/' || prefix[len(prefix)-1] != '/' {
		panic("expect prefix to begin and end with '/'")
	}
	gr := gameRouter{
		router:  httprouter.New(),
		games:   make(map[string]*gameHandler),
		pathGen: newPathGenerator(),
		prefix:  prefix,
	}

	gr.router.GET("/:id", gr.forwardToHandler)
	gr.router.GET("/:id/*etc", gr.forwardToHandler)
	gr.router.POST("/:id/*etc", gr.forwardToHandler)

	return &gr
}

func (gr *gameRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gr.router.ServeHTTP(w, r)
}

func (gr *gameRouter) forwardToHandler(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	etc := p.ByName("etc")

	url, err := url.Parse(etc)
	if err != nil {
		http.Error(w, "error parsing URL", http.StatusInternalServerError)
		return
	}
	r.URL = url

	if gh, ok := gr.games[id]; ok && gh != nil {
		gh.ServeHTTP(w, r)
		return
	}

	http.Error(w, "game "+id+" not found.", http.StatusNotFound)
	return
}

func (gr *gameRouter) addGame(
	config kuba.Config, cookie1, cookie2 *http.Cookie) string {
	// Randomize who plays white
	if mrand.Intn(2) == 0 {
		cookie1, cookie2 = cookie2, cookie1
	}

	id := gr.pathGen.newPath(8)
	gr.games[id] = newGameHandler(config, cookie1, cookie2)

	return gr.prefix + id
}
