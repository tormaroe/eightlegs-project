package queue

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/tormaroe/eightlegs-project/pinkfoot/config"
)

type PushRequest struct {
	Bytes []byte
	Acc   chan struct{}
}

// PersistantQueue ...
type PersistantQueue struct {
	config   config.Config
	pushChan chan PushRequest
	len      *atomicCount
}

// Init initializes a PersistantQueue based on the provided configuration
func Init(conf config.Config) (*PersistantQueue, error) {
	pq := PersistantQueue{
		config:   conf,
		pushChan: make(chan PushRequest, 10),
		len: &atomicCount{
			mut: &sync.Mutex{},
		},
	}

	pr, err := pq.pushRoutine()
	if err != nil {
		return nil, err
	}
	go pr()

	return &pq, nil
}

// Len returns the number of unread messages in the queue.
func (pq *PersistantQueue) Len() int {
	return pq.len.val()
}

// Push will process a PushRequest concurrently.
// The Acc channel will be closed when request has been fulfilled.
// An error is returned if the queue is full.
func (pq *PersistantQueue) Push(req PushRequest) error {

	if max := pq.config.Persistance.MaxUnreadMessages; pq.len.val() >= max {
		return fmt.Errorf("Queue is full with %d messages", max)
	}

	pq.pushChan <- req
	return nil
}

func (pq *PersistantQueue) pushRoutine() (func(), error) {
	f, err := os.OpenFile(pq.config.Persistance.DataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return func() {}, err
	}

	return func() {
		for {
			req := <-pq.pushChan
			if _, err := f.Write(req.Bytes); err != nil {
				log.Fatal(err)
				return
			}
			pq.len.inc()
			close(req.Acc)
		}
	}, nil
}
