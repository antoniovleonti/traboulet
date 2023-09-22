package server

import "testing"

func TestNewPath(t *testing.T) {
	pg := newNonCryptoStringGen()

	p := pg.newString(8)
	if len(p) != 8 {
		t.Errorf("p had unexpected len %d; expected %d", len(p), 8)
	}
}
