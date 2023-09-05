package server

import (
	"fmt"
	"net/http"
	"sync"
  "container/list"
)

type subscriber struct {
  done chan<- struct{}
  w http.ResponseWriter
}

// Streamlines responding to long poll requests (e.g. chat, game updates).
type publisher struct {
	subscribers list.List
	mutex       sync.Mutex
}

// Returns when no more data will be written to writer
func (p *publisher) subscribe(
  w http.ResponseWriter, cancel <-chan struct{}) {
	p.mutex.Lock()
  done := make(chan struct{})
	handle := p.subscribers.PushBack(subscriber{done, w})
  p.mutex.Unlock()

  select {
    case <-cancel:
      p.mutex.Lock()
      p.subscribers.Remove(handle)
      p.mutex.Unlock()
    case <-done:
  }
}

func (p *publisher) publish(msg string) {
  p.do(func(w http.ResponseWriter) {
		fmt.Fprint(w, msg)
	}, true)
}

// For when you need more control over the response than just the body text
func (p *publisher) do(f func(http.ResponseWriter), final bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for el := p.subscribers.Front(); el != nil; el = el.Next() {
    sub, ok := el.Value.(subscriber)
    if !ok {
      panic("non-subscriber pushed to subcriber list")
    }
		f(sub.w)
    if final {
      sub.done <- struct{}{}
    }
	}
  if final {
    p.subscribers.Init()
  }
}
