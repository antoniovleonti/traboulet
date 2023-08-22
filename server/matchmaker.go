package server

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type multigameHandler struct {
	router     *httprouter.Router
	challenges map[string]*challengeHandler
	games      map[string]*gameHandler
	pathGen    *pathGenerator
}

func newMatchmaker() *multigameHandler {
	mm := multigameHandler{
		router:     httprouter.New(),
		challenges: make(map[string]*challengeHandler),
		games:      make(map[string]*gameHandler),
		pathGen:    newPathGenerator(),
	}

	return &mm
}

func (mm *multigameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func newCookieValue(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(randomBytes)[:length]
}
