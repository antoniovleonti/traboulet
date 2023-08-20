package server

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"kuba"
	mrand "math/rand"
	"net/http"
	"net/url"
	"time"
)

type pathGenerator mrand.Rand

func newPathGenerator() *pathGenerator {
	return (*pathGenerator)(mrand.New(mrand.NewSource(time.Now().UnixNano())))
}

func (pg *pathGenerator) Read(bytes []byte) (int, error) {
	return (*mrand.Rand)(pg).Read(bytes)
}

type matchmaker struct {
	router     *httprouter.Router
	challenges map[string]*challengeHandler
	games      map[string]*gameHandler
	pathGen    *pathGenerator
}

func newMatchmaker() *matchmaker {
	mm := matchmaker{
		router:     httprouter.New(),
		challenges: make(map[string]*challengeHandler),
		games:      make(map[string]*gameHandler),
		pathGen:    newPathGenerator(),
	}

	mm.router.POST("/games", mm.postGame)
	mm.router.GET("/games", mm.getChallenges)

	mm.router.GET("/games/:id/*etc", mm.forwardGameRequest)
	mm.router.POST("/games/:id/*etc", mm.forwardGameRequest)

	return &mm
}

func (mm *matchmaker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If request has come in with no cookie, set cookie in the response.
	if len(r.Cookies()) == 0 {
		c := http.Cookie{
			Name:  "Anonymous",
			Value: newCookieValue(8),
			Path:  "/",
		}
		http.SetCookie(w, &c)
	}

	mm.router.ServeHTTP(w, r)
}

func (mm *matchmaker) forwardGameRequest(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	etc := p.ByName("etc")

	url, err := url.Parse(etc)
	if err != nil {
		http.Error(w, "error parsing URL", http.StatusInternalServerError)
		return
	}
	r.URL = url

	if ch, ok := mm.challenges[id]; ok && ch != nil {
		ch.ServeHTTP(w, r)
		return
	}
	if gh, ok := mm.games[id]; ok && gh != nil {
		gh.ServeHTTP(w, r)
		return
	}

	http.Error(w, id+" not found.", http.StatusNotFound)
	return
}

func (mm *matchmaker) postGame(
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
	path := mm.pathGen.newPath(8)
	bind := func(p *publisher, c kuba.Config, c1, c2 *http.Cookie) {
		mm.moveChallengeToGame(path, p, c, c1, c2)
	}
	mm.challenges[path] = newChallengeHandler(cookie, config, bind)

	w.Header().Add("Location", "/games/"+path)
	w.WriteHeader(http.StatusSeeOther)
	w.Write([]byte("success"))
}

func (mm *matchmaker) getChallenges(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b, err := json.Marshal(mm.challenges)
	if err != nil {
		http.Error(
			w, "could not marshal challenges: "+err.Error(),
			http.StatusInternalServerError)
	}
	w.Write(b)
}

func (mm *matchmaker) moveChallengeToGame(
	path string, donePub *publisher, config kuba.Config, cookie1,
	cookie2 *http.Cookie) {
	if _, ok := mm.challenges[path]; !ok {
		// ???? challenge doesn't exist. Do nothing because idk what else to do.
		return
	}
	if _, ok := mm.games[path]; ok {
		// ???? game already exists. Do nothing.
		return
	}

	delete(mm.challenges, path)

	// Randomize who plays white
	if mrand.Intn(2) == 0 {
		cookie1, cookie2 = cookie2, cookie1
	}

	mm.games[path] = newGameHandler(config, cookie1, cookie2)

	if donePub != nil {
		donePub.do(func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusResetContent)
			w.Write([]byte("game has started at same URL; please refresh"))
		})
	}
}

func (pg *pathGenerator) newPath(length int) string {
	randomBytes := make([]byte, length)
	_, err := pg.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:length]
}

func newCookieValue(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(randomBytes)[:length]
}
