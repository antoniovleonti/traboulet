package server

import (
	"fmt"
	"net/http"
	"sync"
)

type subscriber struct {
  c chan struct{}
  w http.ResponseWriter
}

// Streamlines responding to long poll requests (e.g. chat, game updates).
type publisher struct {
	subscribers []subscriber
	mutex       sync.Mutex
}

func (p *publisher) subscribe(w http.ResponseWriter) <-chan struct{}  {
	p.mutex.Lock()
  c := make(chan struct{})
	p.subscribers = append(p.subscribers, subscriber{c, w})
  p.mutex.Unlock()
  return c
}

func (p *publisher) publish(msg string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, s := range p.subscribers {
		fmt.Fprint(s.w, msg)
    s.c <- struct{}{}
	}
	p.subscribers = nil
}

// For when you need more control over the response than just the body text
// Assumption: f will always write a response to EVERY writer
func (p *publisher) do(f func(http.ResponseWriter)) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, s := range p.subscribers {
		f(s.w)
    s.c <- struct{}{}
	}
	p.subscribers = nil
}
