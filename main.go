package main

import (
	"net/http"

	"go-sse/server"
)

func main() {
	// new streaming server
	stream := server.NewServer()

	// handle server
	http.HandleFunc("/sse", stream.ServHTTP)

	// handle client
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./client/index.html")
	})

	http.ListenAndServe(":8080", nil)
}
