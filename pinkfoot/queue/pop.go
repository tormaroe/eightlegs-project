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
	if pq.len.val() == 0 {
		return req.id, false
	}
	pq.popChan <- req
	return req.id, true
}

func (pq *PersistantQueue) popRoutine() (func(), error) {
	f, err := os.OpenFile(pq.config.Persistance.DataFile, os.O_RDONLY, 0644)
	if err != nil {
		return func() {}, err
	}

	// TODO: Add correct offset to f

	return func() {
		for {
			req := <-pq.popChan

			if pq.len.val() == 0 {
				req.Reply <- nil
				continue
			}

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
			pq.len.dec()
			pq.addWaitingForReceipt(req.id)
		}
	}, nil
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
