package server

import (
	"encoding/base32"
	"math/rand"
	"time"
)

type pathGenerator rand.Rand

func newPathGenerator() *pathGenerator {
	return (*pathGenerator)(rand.New(rand.NewSource(time.Now().UnixNano())))
}

func (pg *pathGenerator) Read(bytes []byte) (int, error) {
	return (*rand.Rand)(pg).Read(bytes)
}

func (pg *pathGenerator) newPath(length int) string {
	randomBytes := make([]byte, length)
	_, err := pg.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:length]
}
