package queue

import (
	"encoding/binary"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
)

type PopRequest struct {
	Reply chan []byte
	id    uuid.UUID
}

func (pq *PersistantQueue) Pop(req PopRequest) (uuid.UUID, bool) {
	req.id = uuid.New()
	pq.popChan <- req
	return req.id, true
}

func (pq *PersistantQueue) popRoutine() (func(), error) {
	log.Println("Opening reader for file", pq.config.Persistance.DataFile)
	f, err := os.OpenFile(pq.config.Persistance.DataFile, os.O_RDONLY, 0644)
	if err != nil {
		return func() {}, err
	}

	return func() {
		log.Println("popRoutine starting!")
		for {
			select {
			case req := <-pq.popChan:
				pq.setCurrOffset(f)
				bytes, err := req.read(f)
				if err == io.EOF {
					req.Reply <- nil
					close(req.Reply)
					continue
				} else if err != nil {
					log.Fatal(err)
					return
				}
				req.Reply <- bytes
				close(req.Reply)
				pq.addWaitingForReceipt(req.id, bytes, pq.currOffset)
			case <-pq.stopChan:
				log.Println("popRoutine stopping!")
				pq.setCurrOffset(f)
				f.Close()
				pq.stopWG.Done()
				return
			}
		}
	}, nil
}

func (pq *PersistantQueue) setCurrOffset(f io.Seeker) {
	currOffset, err := f.Seek(0, 1)
	if err != nil {
		log.Fatal(err)
		return
	}
	pq.currOffset = currOffset
}

func (req PopRequest) read(f *os.File) ([]byte, error) {
	sizeBytes := make([]byte, 8)
	nRead, err := f.Read(sizeBytes)

	if err != nil {
		return nil, err
	}

	if nRead == 0 && err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	size := int(binary.LittleEndian.Uint64(sizeBytes))
	bytes := make([]byte, size)
	if _, err = f.Read(bytes); err != nil {
		return bytes, err
	}

	return bytes, nil
}
