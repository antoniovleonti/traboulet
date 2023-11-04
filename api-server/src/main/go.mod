module main

go 1.21.0

replace game => ../game

replace server => ../server

replace evtpub => ../server/event-publisher

require server v0.0.0-00010101000000-000000000000

require (
	evtpub v0.0.0-00010101000000-000000000000 // indirect
	game v0.0.0-00010101000000-000000000000 // indirect
	github.com/antoniovleonti/sse v0.0.0-20230904230022-1b089e02c02c // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
)
