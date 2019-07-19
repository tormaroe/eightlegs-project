package queue

import (
	"encoding/binary"
	"log"
	"os"
)

type PushRequest struct {
	Bytes []byte
	Acc   chan struct{}
}

// Push will process a PushRequest concurrently.
// The Acc channel will be closed when request has been fulfilled.
func (pq *PersistantQueue) Push(req PushRequest) error {
	// TODO: Remove error if it can't happen
	pq.pushChan <- req
	return nil
}

func (pq *PersistantQueue) pushRoutine() (func(), error) {
	log.Println("Opening appender for file", pq.config.Persistance.DataFile)
	f, err := os.OpenFile(pq.config.Persistance.DataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return func() {}, err
	}

	return func() {
		log.Println("pushRoutine starting!")
		for {
			select {
			case req := <-pq.pushChan:
				if err := req.write(f); err != nil {
					log.Fatal(err)
					return
				}
				close(req.Acc)
			case <-pq.stopChan:
				log.Println("pushRoutine stopping!")
				f.Close()
				pq.stopWG.Done()
				return
			}
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
