package server

import (
  "testing"
  "net/http/httptest"
  "io"
)

func TestPublisher(t *testing.T) {
  p := publisher{}

  var subs []*httptest.ResponseRecorder
  for i := 0; i < 100; i++ {
    subs = append(subs, httptest.NewRecorder())
    p.addSubscriber(subs[i])
  }

  msg := "message"
  p.publish(msg)

  if len(p.subscribers) != 0 {
    t.Error("subscribers was not cleared after publish")
  }

  for i := 0; i < 100; i++ {
    resp := subs[i].Result()
    body, _ := io.ReadAll(resp.Body)

    if actual := string(body); actual != msg {
      t.Errorf("recorded message \"%s\" did not match expected message \"%s\".",
               actual, msg)
    }
  }
}
