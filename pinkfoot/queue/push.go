package queue

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

type PushRequest struct {
	Bytes []byte
	Acc   chan struct{}
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
			if err := req.write(f); err != nil {
				log.Fatal(err)
				return
			}
			pq.len.inc()
			close(req.Acc)
		}
	}, nil
}

func (req PushRequest) write(f *os.File) error {
	size := len(req.Bytes)
	sizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBytes, uint64(size))

	if _, err := f.Write(sizeBytes); err != nil {
		return err
	}

	if _, err := f.Write(req.Bytes); err != nil {
		return err
	}
	return nil
}
