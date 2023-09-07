package server

import (
	"github.com/julienschmidt/httprouter"
	"kuba"
	"log"
	mrand "math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type gameRouter struct {
	router     *httprouter.Router
	challenges map[string]*challengeHandler
	games      map[string]*gameHandler
	pathGen    *nonCryptoStringGen
	prefix     string
	mutex      sync.RWMutex
}

func newGameRouter(prefix string) *gameRouter {
	if prefix[0] != '/' || prefix[len(prefix)-1] != '/' {
		panic("expect prefix to begin and end with '/'")
	}
	gr := gameRouter{
		router:  httprouter.New(),
		games:   make(map[string]*gameHandler),
		pathGen: newNonCryptoStringGen(),
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
	gr.mutex.RLock()
	defer gr.mutex.RUnlock()

	id := p.ByName("id")
	etc := httprouter.CleanPath(p.ByName("etc"))

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
	config kuba.Config, cookie1, cookie2 *http.Cookie) (string, error) {
	gr.mutex.Lock()
	defer gr.mutex.Unlock()

	// Randomize who plays white
	if mrand.Intn(2) == 0 {
		cookie1, cookie2 = cookie2, cookie1
	}

	game, err := newGameHandler(config, cookie1, cookie2)
	if err != nil {
		return "", err
	}
	id := gr.pathGen.newString(8)
	gr.games[id] = game

	log.Print("Created game " + id + ".")

	return gr.prefix + id, nil
}

func (gr *gameRouter) PeriodicallyDeleteGamesOlderThan(d time.Duration) {
	for {
		time.Sleep(d)
		gr.deleteGamesOlderThan(d)
	}
}

func (gr *gameRouter) deleteGamesOlderThan(d time.Duration) {
	gr.mutex.Lock()
	defer gr.mutex.Unlock()

	count := 0
	for id, game := range gr.games {
		actual := game.DurationSinceCompletion()
		if actual == nil {
			continue
		}
		if *actual > d {
			delete(gr.games, id)
			count++
		}
	}
	if count > 0 {
		log.Printf(
			"Cleaned up %d completed games (%d games remain).", count, len(gr.games))
	}
}
