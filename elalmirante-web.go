package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"gitlab.com/ozkar99/middleware"
)

const (
	configFile = "config.yml"
)

func main() {
	// parse port flag
	port := flag.String("port", "3005", "The port to bind the server.")
	flag.Parse()

	// error if config file is not present
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatal("Cannot find " + configFile + " file!")
	}

	// server
	mux := createMux()
	log.Printf("Starting server... listening on http://0.0.0.0:%s \n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, middleware.Lowercase(mux)))
}
