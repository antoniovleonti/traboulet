package main

import (
	"log"
	"net/http"
	"server"
)

func main() {
	router := server.NewRootRouter()
	log.Print("starting server...")
	log.Fatal(http.ListenAndServe(":25565", router))
}
