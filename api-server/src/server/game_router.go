package server

import (
	"game"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
	mrand "math/rand"
  "errors"
  "evtpub"
)

type createGameFnT func(
	deleteChallengeFn, game.Config, *http.Cookie, *http.Cookie) (*url.URL, error)

type gameRouter struct {
	router     *httprouter.Router
	games      map[string]*gameHandler
	pathGen    *nonCryptoStringGen
	urlBase    *url.URL
	mutex      sync.RWMutex
  evpub      evtpub.EventPublisher
}

func newGameRouter(urlBase *url.URL, evpub evtpub.EventPublisher) *gameRouter {
	gr := gameRouter{
		router:  httprouter.New(),
		games:   make(map[string]*gameHandler),
		pathGen: newNonCryptoStringGen(),
		urlBase:  urlBase,
    evpub: evpub,
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
	etc := httprouter.CleanPath(p.ByName("etc"))

	url, err := url.Parse(etc)
	if err != nil {
		http.Error(w, "error parsing URL", http.StatusInternalServerError)
		return
	}
	r.URL = url

	gr.mutex.RLock()
	gh, ok := gr.games[id]
	gr.mutex.RUnlock()

	if ok && gh != nil {
		gh.ServeHTTP(w, r)
		return
	}

	http.Error(w, "game "+id+" not found.", http.StatusNotFound)
	return
}

func (gr *gameRouter) addGame(
	deleteChallengeCb deleteChallengeFn, config game.Config,
	cookie1, cookie2 *http.Cookie) (*url.URL, error) {

	gr.mutex.Lock()
	defer gr.mutex.Unlock()

	if len(gr.games) >= 100 {
		return nil, errors.New("Too many games in play; try again later.")
	}

	// Randomize who plays white
	if mrand.Intn(2) == 0 {
		cookie1, cookie2 = cookie2, cookie1
	}

	id := gr.pathGen.newString(8)
  fullPath := gr.urlBase.JoinPath(id)
  chpub, err := gr.evpub.NewChannelPublisher(fullPath.String())
  if err != nil {
    return nil, err
  }
	game, err :=
    newGameHandler(deleteChallengeCb, *chpub, config, cookie1, cookie2)
	if err != nil {
		return nil, err
	}
	gr.games[id] = game

	log.Print("Created game " + id + ".")

	return fullPath, nil
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
			gr.deleteGame(id)
			count++
		}
	}
	if count > 0 {
		log.Printf(
			"Cleaned up %d completed games (%d games remain).", count, len(gr.games))
	}
}

func (gr *gameRouter) deleteGame(id string) {
  gr.games[id].TearDown()
	delete(gr.games, id)
}
