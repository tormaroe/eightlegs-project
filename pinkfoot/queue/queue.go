package queue

import (
	"github.com/google/uuid"
	"github.com/tormaroe/eightlegs-project/pinkfoot/config"
)

// PersistantQueue ...
type PersistantQueue struct {
	config           config.Config
	pushChan         chan PushRequest
	popChan          chan PopRequest
	waitReceiptChan  chan waitingForReceipt
	receiptChan      chan uuid.UUID
	waitReceiptQueue []waitingForReceipt
	len              *atomicCount
	// TODO: len is bad when starting Queue on a non-empty file
}

// Init initializes a PersistantQueue based on the provided configuration
func Init(conf config.Config) (*PersistantQueue, error) {
	pq := PersistantQueue{
		config:          conf,
		pushChan:        make(chan PushRequest, 20),
		popChan:         make(chan PopRequest, 20),
		waitReceiptChan: make(chan waitingForReceipt, 10),
		receiptChan:     make(chan uuid.UUID, 20),
		len:             &atomicCount{},
	}

	pur, err := pq.pushRoutine()
	if err != nil {
		return nil, err
	}
	go pur()

	por, err := pq.popRoutine()
	if err != nil {
		return nil, err
	}
	go por()

	go pq.receiptRoutine()

	// TODO: Routine to truncate logfile (must pause push and pop)

	return &pq, nil
}

// Len returns the number of unread messages in the queue.
func (pq *PersistantQueue) Len() int {
	return pq.len.val()
}
