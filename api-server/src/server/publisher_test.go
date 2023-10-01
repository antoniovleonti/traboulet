package server

import (
	"io"
	"log"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublisher(t *testing.T) {
	p := publisher{}

	type sub struct {
		done <-chan struct{}
		rr   *httptest.ResponseRecorder
	}

	var subs []sub
	for i := 0; i < 100; i++ {
		rr := httptest.NewRecorder()
		done := make(chan struct{})
		go func() {
			p.subscribe(rr, make(chan struct{}))
			done <- struct{}{}
		}()
		subs = append(subs, sub{done, rr})
	}
	time.Sleep(time.Millisecond)

	msg := "message"
	go p.publish(msg)

	for _, s := range subs {
		<-s.done
		resp := s.rr.Result()
		body, _ := io.ReadAll(resp.Body)

		if actual := string(body); actual != msg {
			t.Errorf(
				"recorded message \"%s\" did not match expected message \"%s\".",
				actual, msg)
		}
	}
	time.Sleep(time.Millisecond)

	if p.subscribers.Len() != 0 {
		t.Error("subscribers was not cleared after publish")
	}
}

func TestCancel(t *testing.T) {
	p := publisher{}

	type sub struct {
		done   <-chan struct{}
		cancel chan<- struct{}
		rr     *httptest.ResponseRecorder
	}

	var doneSubs []sub
	var cancelSubs []sub

	appendSub := func(s *[]sub) {
		rr := httptest.NewRecorder()
		done := make(chan struct{})
		cancel := make(chan struct{})
		go func() {
			p.subscribe(rr, cancel)
			done <- struct{}{}
		}()
		*s = append(*s, sub{done, cancel, rr})
	}

	for i := 0; i < 100; i++ {
		appendSub(&doneSubs)
		appendSub(&cancelSubs)
	}
	time.Sleep(5 * time.Millisecond)

	// cancel all the cancel subscribers
	log.Print("cancelling all cancelSubs")
	for _, s := range cancelSubs {
		s.cancel <- struct{}{}
		<-s.done
	}

	msg := "message"
	go p.publish(msg)

	// check cancelSubs didn't get written to
	for _, s := range cancelSubs {
		resp := s.rr.Result()
		body, _ := io.ReadAll(resp.Body)

		if actual := string(body); actual != "" {
			t.Errorf(
				"recorded message \"%s\" did not match expected message \"%s\".",
				actual, "")
		}
	}

	for _, s := range doneSubs {
		<-s.done
		resp := s.rr.Result()
		body, _ := io.ReadAll(resp.Body)

		if actual := string(body); actual != msg {
			t.Errorf(
				"recorded message \"%s\" did not match expected message \"%s\".",
				actual, msg)
		}
	}
	time.Sleep(time.Millisecond)

	if p.subscribers.Len() != 0 {
		t.Error("doneSubs was not cleared after publish")
	}

}
