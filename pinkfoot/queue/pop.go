package queue

import (
	"encoding/binary"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
)

// Message has the bytes of a popped message from the queue,
// as well as an ID that must be used in order to acknowledge
// receipt of the message. If the receipt is not acknowledged,
// the message will be re-inserted into the queue after some time.
type Message struct {
	Bytes []byte
	ID    uuid.UUID
}

// Pop a message from the queue. A pointer to the message will be sent through
// the returned channel.
// If there are no messages in the queue, the channel will simply be closed.
func (pq *PersistantQueue) Pop() chan *Message {
	c := make(chan *Message)
	pq.popChan <- c
	return c
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
			case replyChan := <-pq.popChan:
				pq.setCurrOffset(f)
				bytes, err := readMessage(f)
				if err == io.EOF {
					close(replyChan)
					continue
				} else if err != nil {
					log.Fatal(err)
					return
				}
				id := uuid.New()
				replyChan <- &Message{
					ID:    id,
					Bytes: bytes,
				}
				close(replyChan)
				pq.addWaitingForReceipt(id, bytes, pq.currOffset)
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

func readMessage(f *os.File) ([]byte, error) {
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
