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

func (h *Handler) handlePop(w http.ResponseWriter, r *http.Request) {
	log.Printf("POP\n")

	reply := make(chan []byte)
	req := queue.PopRequest{
		Reply: reply,
	}
	// TODO: Refactor nesting (error cases first)
	if id, hasMessages := h.Queue.Pop(req); hasMessages {
		bytes := <-reply
		if bytes == nil || len(bytes) == 0 {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.Header().Add("X-Correlation-ID", id.String())
			w.Write(bytes)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
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
