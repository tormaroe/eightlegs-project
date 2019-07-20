package queue

import (
	"encoding/binary"
	"log"
	"os"
	"sync"
)

type pushRequest struct {
	bytes []byte
	wg    sync.WaitGroup
}

// Push will ...
func (pq *PersistantQueue) Push(b []byte) *sync.WaitGroup {
	req := pushRequest{
		bytes: b,
		wg:    sync.WaitGroup{},
	}
	req.wg.Add(1)
	pq.pushChan <- &req
	return &req.wg
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
				req.wg.Done()
			case <-pq.stopChan:
				log.Println("pushRoutine stopping!")
				f.Close()
				pq.stopWG.Done()
				return
			}
		}
	}, nil
}

func (req pushRequest) write(f *os.File) error {
	size := len(req.bytes)
	sizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBytes, uint64(size))

	if _, err := f.Write(sizeBytes); err != nil {
		return err
	}

	if _, err := f.Write(req.bytes); err != nil {
		return err
	}
	return nil
}
