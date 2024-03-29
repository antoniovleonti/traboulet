package server

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/url"
	"time"
  "evtpub"
)

type rootRouter struct {
	router       *httprouter.Router
	challengeRtr *challengeRouter
	gameRtr      *gameRouter
	nameGen      *nonCryptoStringGen
  evPub     evtpub.EventPublisher
}

func NewRootRouter(evPub evtpub.EventPublisher) *rootRouter {
  gameRtrURLBase, err := url.Parse("/games/")
  if err != nil {
    panic(err)
  }
	rr := rootRouter{
		router:  httprouter.New(),
		nameGen: newNonCryptoStringGen(),
    evPub:   evPub,
		gameRtr: newGameRouter(gameRtrURLBase, evPub),
	}

  challengeRtrURLBase, err := url.Parse("/challenges/")
  if err != nil {
    panic(err)
  }
  rr.challengeRtr =
    newChallengeRouter(challengeRtrURLBase, rr.gameRtr.addGame, evPub)

	rr.router.GET("/games", rr.fwdToGameRouter)
	rr.router.POST("/games", rr.fwdToGameRouter)
	rr.router.GET("/games/*etc", rr.fwdToGameRouter)
	rr.router.POST("/games/*etc", rr.fwdToGameRouter)

	rr.router.GET("/challenges", rr.fwdToChallengeRouter)
	rr.router.POST("/challenges", rr.fwdToChallengeRouter)
	rr.router.GET("/challenges/*etc", rr.fwdToChallengeRouter)
	rr.router.POST("/challenges/*etc", rr.fwdToChallengeRouter)

	go rr.challengeRtr.PeriodicallyDeleteOldChallenges(10 * time.Minute)
	go rr.gameRtr.PeriodicallyDeleteGamesOlderThan(10 * time.Minute)

	return &rr
}

func (rr *rootRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If request has come in with no cookie, set cookie in the response.
	if len(r.Cookies()) == 0 {
		c := http.Cookie{
			Name:  rr.nameGen.newString(8),
			Value: newCookieValue(8),
			Path:  "/",
		}
		http.SetCookie(w, &c)
	}

	rr.router.ServeHTTP(w, r)
}

func (rr *rootRouter) fwdToGameRouter(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	etc := httprouter.CleanPath(p.ByName("etc"))

	url, err := url.Parse(etc)
	if err != nil {
		http.Error(w, "error parsing URL", http.StatusInternalServerError)
		return
	}
	r.URL = url

	rr.gameRtr.ServeHTTP(w, r)
}

func (rr *rootRouter) fwdToChallengeRouter(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	etc := httprouter.CleanPath(p.ByName("etc"))

	url, err := url.Parse(etc)
	if err != nil {
		http.Error(w, "error parsing URL", http.StatusInternalServerError)
		return
	}
	r.URL = url

	rr.challengeRtr.ServeHTTP(w, r)
}

func newCookieValue(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(randomBytes)[:length]
}
