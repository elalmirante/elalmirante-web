package main

import (
	"flag"
	"net/http"
)

func main() {
	// parse port flag
	port := flag.Int("port", 3005, "The port to bind the server.")
	flag.Parse()

	// setup routes
	mux := http.NewServeMux()

	http.ListenAndServe("*:"+string(*port), mux)
}
