package main

import (
	"log"
	"net/http"
	"server"
  "evtpub"
  "flag"
)

func main() {
  // Default is what I use for testing. For production replace this flag with
  // the IP of the nginx container & the admin server's port.
  proxyHostname := flag.String(
    "P", "http://localhost:25566", "destination for push stream events")
  flag.Parse()

  evpub := evtpub.NewExtEventPublisher(*proxyHostname)
	router := server.NewRootRouter(evpub)
	log.Print("starting server...")
	log.Fatal(http.ListenAndServe(":25565", router))
}
