module server

go 1.21.0

replace game => ../game

require (
	github.com/antoniovleonti/sse v0.0.0-20230904230022-1b089e02c02c
	github.com/julienschmidt/httprouter v1.3.0
	github.com/r3labs/sse/v2 v2.10.0
	game v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.14.0 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
)
