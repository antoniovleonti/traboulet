package server

import (
	"fmt"
	"net/http"
	"sync"
)

// Streamlines responding to long poll requests (e.g. chat, game updates).
type publisher struct {
	subscribers []http.ResponseWriter
	mutex       sync.Mutex
}

func (p *publisher) addSubscriber(w http.ResponseWriter) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.subscribers = append(p.subscribers, w)
}

func (p *publisher) publish(msg string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, w := range p.subscribers {
		fmt.Fprint(w, msg)
	}
	p.subscribers = nil
}

// For when you need more control over the response than just the body text
// Assumption: f will always write a response to EVERY writer
func (p *publisher) do(f func(http.ResponseWriter)) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, w := range p.subscribers {
		f(w)
	}
	p.subscribers = nil
}
