package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMatchmaker(t *testing.T) {
	var _ http.Handler = (*multigameHandler)(nil)

	mm := newMatchmaker()
	if mm == nil {
		t.Error("multigameHandler was nil")
	}
}


func TestServeHTTPSetsCookie(t *testing.T) {
	mm := newMatchmaker()

	req, err := http.NewRequest("GET", "/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	mm.ServeHTTP(rr, req)

	if rr.Header().Get("Set-Cookie") == "" {
		t.Error("cookie was not set")
	}
}

