package server

import (
	"io"
	"net/http/httptest"
	"testing"
  "time"
)

func TestPublisher(t *testing.T) {
	p := publisher{}

  type sub struct {
    done <-chan struct{}
    rr *httptest.ResponseRecorder
  }

	var subs []sub
	for i := 0; i < 100; i++ {
    rr := httptest.NewRecorder()
    done := p.subscribe(rr)
    subs = append(subs, sub{done, rr})
	}

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

	if len(p.subscribers) != 0 {
		t.Error("subscribers was not cleared after publish")
	}
}
