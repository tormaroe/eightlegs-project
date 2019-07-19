package queue

import (
	"log"
	"time"

	"github.com/google/uuid"
)

type waitingForReceipt struct {
	id       uuid.UUID
	poppedAt time.Time
	bytes    []byte
	offset   int64
}

func (pq *PersistantQueue) addWaitingForReceipt(id uuid.UUID, bytes []byte, offset int64) {
	wfr := waitingForReceipt{
		id:       id,
		poppedAt: time.Now(),
		bytes:    bytes,
		offset:   offset,
	}
	pq.waitReceiptChan <- wfr
}

func (pq *PersistantQueue) AddReceipt(id uuid.UUID) {
	log.Println("AddReceipt")
	pq.receiptChan <- id
}

func (pq *PersistantQueue) receiptRoutine() {
	log.Println("receiptRoutine starting!")
	frequency := time.Duration(pq.config.Acknowledgement.SecondsBeforeReInsert / 10)
	for {
		select {
		case wfr := <-pq.waitReceiptChan:
			log.Println("RECEIPT ROUTINE wait for receipt:", wfr)
			pq.waitReceiptList[wfr.id] = wfr
		case id := <-pq.receiptChan:
			log.Println("RECEIPT ROUTINE receipt:", id)
			delete(pq.waitReceiptList, id) // does nothing if id is not in list
		case <-time.After(frequency * time.Second):
			pq.rePushStalePulls()
		case <-pq.stopChan:
			log.Println("receiptRoutine stopping!")
			pq.stopWG.Done()
			return
		}
	}
}

func (pq *PersistantQueue) rePushStalePulls() {
	limit := pq.config.Acknowledgement.SecondsBeforeReInsert
	now := time.Now()
	for k, v := range pq.waitReceiptList {
		waitDuration := now.Sub(v.poppedAt)
		if int(waitDuration.Seconds()) > limit {
			log.Println("Re-inserting stale pull")
			ack := make(chan struct{})
			pq.Push(PushRequest{
				Acc:   ack,
				Bytes: v.bytes,
			})
			<-ack
			delete(pq.waitReceiptList, k)
		}
	}
}
