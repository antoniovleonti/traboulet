package server

import (
	"kuba"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewChallengeHandler(t *testing.T) {
	var _ http.Handler = (*challengeHandler)(nil)

	cb := func(p *publisher, c kuba.Config, w *http.Cookie, b *http.Cookie) {}

	ch := newChallengeHandler(fakeWhiteCookie(), kuba.Config{}, cb)
	if ch == nil {
		t.Error("nil challenge handler")
	}
}

func TestPostJoinValid(t *testing.T) {
	white := fakeWhiteCookie()
	black := fakeBlackCookie()

	callbackCalled := false
	cb := func(p *publisher, c kuba.Config, w *http.Cookie, b *http.Cookie) {
		if w.Value != white.Value {
			t.Error("white cookie did not match expectation")
		}
		if b.Value != black.Value {
			t.Error("black cookie did not match expectation")
		}
		p.publish("success")
		callbackCalled = true
	}

	ch := newChallengeHandler(white, kuba.Config{}, cb)
	if ch == nil {
		t.Error("nil challenge handler")
	}

	// Build request
	postJoinReq, err := http.NewRequest("POST", "/join", nil)
	if err != nil {
		t.Fatal(err)
	}
	postJoinReq.AddCookie(black)

	// Add a subscriber to game updates
	getUpdateReq, err := http.NewRequest("GET", "/update", nil)
	if err != nil {
		t.Fatal(err)
	}
	getUpdateResp := httptest.NewRecorder()
	ch.ServeHTTP(getUpdateResp, getUpdateReq)

	// Run the request through our handler
	postJoinResp := httptest.NewRecorder()
	ch.ServeHTTP(postJoinResp, postJoinReq)

	if !callbackCalled {
		t.Error("callback was never called")
	}

	if getUpdateResp.Body.String() != "success" {
		t.Errorf("update resp (\"%s\") did not match expectation (\"success\")",
			getUpdateResp.Body.String())
	}
}

func TestPostJoinNoCookie(t *testing.T) {
	white := fakeWhiteCookie()

	callbackCalled := false
	cb := func(p *publisher, c kuba.Config, w *http.Cookie, b *http.Cookie) {
		callbackCalled = true
	}

	ch := newChallengeHandler(white, kuba.Config{}, cb)
	if ch == nil {
		t.Error("nil challenge handler")
	}

	// Build request
	postJoinReq, err := http.NewRequest("POST", "/join", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Run the request through our handler
	postJoinResp := httptest.NewRecorder()
	ch.ServeHTTP(postJoinResp, postJoinReq)

	if callbackCalled {
		t.Error("callback was called when it shouldn't have been")
	}

	if postJoinResp.Code != http.StatusUnauthorized {
		t.Errorf("update resp code (%d) did not match expectation (%d)",
			postJoinResp.Code, http.StatusUnauthorized)
	}
}

func TestPostJoinSameCookie(t *testing.T) {
	white := fakeWhiteCookie()

	callbackCalled := false
	cb := func(p *publisher, c kuba.Config, w *http.Cookie, b *http.Cookie) {
		callbackCalled = true
	}

	ch := newChallengeHandler(white, kuba.Config{}, cb)
	if ch == nil {
		t.Error("nil challenge handler")
	}

	// Build request
	postJoinReq, err := http.NewRequest("POST", "/join", nil)
	if err != nil {
		t.Fatal(err)
	}
	postJoinReq.AddCookie(white)

	// Run the request through our handler
	postJoinResp := httptest.NewRecorder()
	ch.ServeHTTP(postJoinResp, postJoinReq)

	if callbackCalled {
		t.Error("callback was called when it shouldn't have been")
	}

	if postJoinResp.Code != http.StatusBadRequest {
		t.Errorf("update resp code (%d) did not match expectation (%d)",
			postJoinResp.Code, http.StatusBadRequest)
	}
}

func TestPostJoinAgain(t *testing.T) {
	white := fakeWhiteCookie()
	black := fakeBlackCookie()

	callbackCalled := false
	cb := func(p *publisher, c kuba.Config, w *http.Cookie, b *http.Cookie) {
		callbackCalled = true
	}

	ch := newChallengeHandler(white, kuba.Config{}, cb)
	if ch == nil {
		t.Error("nil challenge handler")
	}

	// Build request
	postJoinReq, err := http.NewRequest("POST", "/join", nil)
	if err != nil {
		t.Fatal(err)
	}
	postJoinReq.AddCookie(black)

	// Run the request through our handler
	postJoinResp := httptest.NewRecorder()
	ch.ServeHTTP(postJoinResp, postJoinReq)

	if !callbackCalled {
		t.Error("callback was never called")
	}

	// just do it again
	postJoinResp2 := httptest.NewRecorder()
	ch.ServeHTTP(postJoinResp2, postJoinReq)

	if postJoinResp2.Code != http.StatusInternalServerError {
		t.Errorf("update resp code (%d) did not match expectation (%d)",
			postJoinResp2.Code, http.StatusInternalServerError)
	}
}
