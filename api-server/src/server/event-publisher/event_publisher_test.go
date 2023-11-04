package evtpub

import (
  "testing"
  "net/url"
  "net/http"
  "net/http/httptest"
	"github.com/julienschmidt/httprouter"
)

func TestExtEventPublisherIsEventPublisher(t *testing.T) {
	var _ EventPublisher = (*ExtEventPublisher)(nil)
}

func TestMockEventPublisherIsEventPublisher(t *testing.T) {
	var _ EventPublisher = (*MockEventPublisher)(nil)
}

type mockProxy struct {
  responder func(http.ResponseWriter, *http.Request)
  t *testing.T
  rtr *httprouter.Router
}

func newMockProxy(t *testing.T) *mockProxy {
  mp := mockProxy{
    rtr: httprouter.New(),
    t: t,
  }
  mp.rtr.POST("/*path", mp.post)
  mp.rtr.DELETE("/*path", mp.post)
  return &mp
}

func (mp *mockProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  mp.rtr.ServeHTTP(w, r)
}

func (mp *mockProxy) setResponder(
  responder func(http.ResponseWriter, *http.Request)) {

  mp.responder = responder
}

func (mp *mockProxy) post(
  w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

  mp.responder(w, r)
}

func (mp *mockProxy) del(
  w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

  mp.responder(w, r)
}

func TestProxyInitPushDelete(t *testing.T) {
  // Set up server to consume message on port 80
  mp := newMockProxy(t)
  proxyServer := httptest.NewServer(mp)
  defer proxyServer.Close()

  pep := NewExtEventPublisher(proxyServer.URL)

  path, err := url.Parse("/test/path/")
  if err != nil {
    t.Error(err)
  }

  mp.setResponder(func(w http.ResponseWriter, r *http.Request) {
    expectedURL, err := url.JoinPath("/test/path", "event-publisher")
    if err != nil {
      t.Error(err)
    }
    if r.URL.String() != expectedURL {
      t.Error("expected url to equal '"+expectedURL+"'; was '"+r.URL.String()+
      ".")
    }
  })

  err = pep.initChannel(path)
  if err != nil {
    t.Error(err)
  }

  if _, ok := pep.channels[path.String()]; !ok {
    t.Error("expected channels["+path.String()+"] to exist.")
  }
  err = pep.push(path, "update", "data")
  if err != nil {
    t.Error(err)
  }
  if _, ok := pep.channels[path.String()]; !ok {
    t.Error("expected channels["+path.String()+"] to exist.")
  }

  err = pep.deleteChannel(path)
  if err != nil {
    t.Error(err)
  }
  if _, ok := pep.channels[path.String()]; ok {
    t.Error("expected channels["+path.String()+"] to be deleted.")
  }
}

func TestExtPushUpdateNoInit(t *testing.T) {
  // Set up server to consume message on port 80
  mp := newMockProxy(t)
  proxyServer := httptest.NewServer(mp)
  defer proxyServer.Close()

  pep := NewExtEventPublisher(proxyServer.URL)


  path, err := url.Parse("/test/path/")
  if err != nil {
    t.Error(err)
  }

  mp.setResponder(func(w http.ResponseWriter, r *http.Request) {
    t.Error(
      "Request should have not been made by client (channel does not exist)")
  })
  err = pep.push(path, "update", "data")
  if err == nil {
    t.Error("expected error; channel does not exist")
  }
  err = pep.deleteChannel(path)
  if err == nil {
    t.Error("expected error; channel does not exist")
  }
}

func TestMockEventPublisherInitPushDelete(t *testing.T) {
  mep := NewMockEventPublisher()

  path, err := url.Parse("/test/path/")
  if err != nil {
    t.Error(err)
  }

  err = mep.initChannel(path)
  if err != nil {
    t.Error(err)
  }

  stream, ok := mep.Channels[path.String()]
  if !ok {
    t.Error("expected channels["+path.String()+"] to exist.")
  }

  err = mep.push(path, "update", "data")
  err = mep.push(path, "update", "data")
  err = mep.push(path, "update", "data")
  err = mep.push(path, "update", "data")
  if err != nil {
    t.Error(err)
  }

  if len(stream.Pushes) != 4 {
    t.Error("expected stream to be of length 4")
  }

  err = mep.deleteChannel(path)
  if err != nil {
    t.Error(err)
  }
  if !stream.Deleted {
    t.Error("expected stream to be deleted")
  }
}

func TestMockEventPublisherNoInit(t *testing.T) {
  mep := NewMockEventPublisher()

  path, err := url.Parse("/test/path/")
  if err != nil {
    t.Error(err)
  }

  err = mep.push(path, "update", "data")
  if err == nil {
    t.Error("shouldn't be able to push without initializing channel")
  }

  err = mep.deleteChannel(path)
  if err == nil {
    t.Error("shouldn't be able to delete channel without initializing it")
  }
}
