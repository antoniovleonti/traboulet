package server

import (
	"encoding/base32"
	"math/rand"
	"time"
)

type nonCryptoStringGen rand.Rand

func newNonCryptoStringGen() *nonCryptoStringGen {
	return (*nonCryptoStringGen)(rand.New(rand.NewSource(time.Now().UnixNano())))
}

func (pg *nonCryptoStringGen) read(bytes []byte) (int, error) {
	return (*rand.Rand)(pg).Read(bytes)
}

func (pg *nonCryptoStringGen) newString(length int) string {
	randomBytes := make([]byte, length)
	_, err := pg.read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:length]
}
