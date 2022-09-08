package server

import (
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	Message     chan string
	NewClient   chan chan string
	CloseClient chan chan string
	Clients     map[chan string]bool
}

func NewServer() *Server {
	serv := &Server{
		Message:     make(chan string),
		NewClient:   make(chan chan string),
		CloseClient: make(chan chan string),
		Clients:     make(map[chan string]bool),
	}

	go serv.listen()
	serv.send()

	return serv
}

// send a message every 10s
func (s *Server) send() {
	go func() {
		for {
			time.Sleep(time.Second * 10)
			now := time.Now().Format("2006-01-02 15:04:05")
			s.Message <- fmt.Sprintf("%v", now)
		}
	}()
}

// listen client connect or disconnect
func (s *Server) listen() {
	for {
		select {
		case c := <-s.NewClient:
			// add a client to list
			s.Clients[c] = true
		case c := <-s.CloseClient:
			// delete a client to list
			delete(s.Clients, c)
			close(c)
		case msg := <-s.Message:
			// send a message to every clients
			for c := range s.Clients {
				c <- msg
			}
		}
	}
}

func (s *Server) ServHTTP(w http.ResponseWriter, r *http.Request) {
	/// make sure flusher is ok
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// set event-stream header
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// make a message channel
	clientChan := make(chan string)

	// add client to list
	s.NewClient <- clientChan

	// close message channel and remove client from list
	defer func() {
		s.CloseClient <- clientChan
	}()

	doneNotify := r.Context().Done()

	for {
		select {
		case <-doneNotify:
			// listen client close the website
			return
		case msg := <-clientChan:
			// send a message to the client
			fmt.Fprintf(w, "data: %s\n\n", msg)

			// flush sends any buffered data to the client
			flusher.Flush()
		}
	}
}
