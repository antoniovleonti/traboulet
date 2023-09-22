module main

go 1.21

replace kuba => ../kuba

replace server => ../server

require server v0.0.0-00010101000000-000000000000

require (
	github.com/antoniovleonti/sse v0.0.0-20230904230022-1b089e02c02c // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	kuba v0.0.0-00010101000000-000000000000 // indirect
)
