package api

import (
	"io"
	"log"
	"net/http"

	"github.com/tormaroe/eightlegs-project/pinkfoot/queue"
)

// Handler will handle the queue API
// Http method POST will enqueue the request body
// Http method GET will dequeue a message
type Handler struct {
	Queue *queue.PersistantQueue
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.handlePush(w, r)
	} else if r.Method == "GET" {
		log.Printf("POP\n")
		w.Write([]byte("....."))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handlePush(w http.ResponseWriter, r *http.Request) {
	log.Printf("PUSH %d bytes\n", r.ContentLength)

	bytes := make([]byte, r.ContentLength)
	_, err := r.Body.Read(bytes)
	if err != io.EOF {
		log.Printf("Failed reading body: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	acc := make(chan struct{})
	err = h.Queue.Push(queue.PushRequest{
		Acc:   acc,
		Bytes: bytes,
	})

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	<-acc // Wait for completion
	// TODO: Timeout ??
	defer log.Printf("Queue length is %d\n", h.Queue.Len())
	w.WriteHeader(http.StatusNoContent)
}
