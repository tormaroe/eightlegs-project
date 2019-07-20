package api

import (
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/tormaroe/eightlegs-project/pinkfoot/queue"
)

// Handler will handle the queue API
// Http method POST will enqueue the request body
// Http method GET will dequeue a message
// Http method PUT will acknowledge reception of message
type Handler struct {
	Queue *queue.PersistantQueue
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.handlePush(w, r)
	case "GET":
		h.handlePop(w, r)
	case "PUT":
		h.handleReceipt(w, r)
	default:
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

	done := h.Queue.Push(bytes)
	done.Wait() // blocks until committed to storage

	// TODO: Timeout ??
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handlePop(w http.ResponseWriter, r *http.Request) {
	log.Printf("POP\n")

	reply := make(chan []byte)
	req := queue.PopRequest{
		Reply: reply,
	}

	id, hasMessages := h.Queue.Pop(req)
	if !hasMessages {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	bytes := <-reply
	if bytes == nil || len(bytes) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("X-Correlation-ID", id.String())
	w.Write(bytes)
}

func (h *Handler) handleReceipt(w http.ResponseWriter, r *http.Request) {
	log.Printf("RECEIPT\n")

	idStr := r.Header.Get("X-Correlation-ID")

	if len(idStr) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	h.Queue.AddReceipt(id)

	// Can do this concurrently (respond directly) since a receipt failure is no big deal
	w.WriteHeader(http.StatusNoContent)
}
