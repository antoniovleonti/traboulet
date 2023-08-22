package server

import (
  "math/rand"
  "time"
  "encoding/base32"
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
