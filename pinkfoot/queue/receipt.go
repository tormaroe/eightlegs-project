package queue

import (
	"log"
	"time"

	"github.com/google/uuid"
)

type waitingForReceipt struct {
	id       uuid.UUID
	poppedAt time.Time
	// TODO: Add message bytes
}

func (pq *PersistantQueue) addWaitingForReceipt(id uuid.UUID) {
	wfr := waitingForReceipt{
		id:       id,
		poppedAt: time.Now(),
	}
	pq.waitReceiptChan <- wfr
}

func (pq *PersistantQueue) AddReceipt(id uuid.UUID) {
	log.Println("AddReceipt")
	pq.receiptChan <- id
}

func (pq *PersistantQueue) receiptRoutine() {
	for {
		select {
		case wfr := <-pq.waitReceiptChan:
			log.Println("RECEIPT ROUTINE wait for receipt:", wfr)
			// TODO: Add wfr to list
		case res := <-pq.receiptChan:
			log.Println("RECEIPT ROUTINE receipt:", res)
			// TODO: Remove res from list
			// TODO: Advanse start offset if possible
		case <-time.After(10 * time.Second):
			log.Println("RECEIPT ROUTINE time to check for re-inserts")
			// TODO: Re-push messages that's been in wait list for a long time (duration from config)
		}
	}
}
