package queue

import (
	"sync"

	"github.com/google/uuid"
	"github.com/tormaroe/eightlegs-project/pinkfoot/config"
)

// PersistantQueue ...
type PersistantQueue struct {
	config config.Config

	pushChan        chan *pushRequest
	popChan         chan PopRequest
	waitReceiptChan chan waitingForReceipt
	receiptChan     chan uuid.UUID
	waitReceiptList map[uuid.UUID]waitingForReceipt
	stopChan        chan struct{}
	stopWG          sync.WaitGroup

	currOffset int64
}

// Init initializes a PersistantQueue based on the provided configuration
func Init(conf config.Config) (*PersistantQueue, error) {
	pq := PersistantQueue{
		config:          conf,
		pushChan:        make(chan *pushRequest, 0),
		popChan:         make(chan PopRequest, 20),
		waitReceiptChan: make(chan waitingForReceipt, 10),
		receiptChan:     make(chan uuid.UUID, 20),
		waitReceiptList: make(map[uuid.UUID]waitingForReceipt),
	}

	err := pq.start()
	go pq.truncate()

	return &pq, err
}

func (pq *PersistantQueue) start() error {

	pq.stopChan = make(chan struct{})

	pur, err := pq.pushRoutine()
	if err != nil {
		return err
	}
	go pur()

	por, err := pq.popRoutine()
	if err != nil {
		return err
	}
	go por()

	go pq.receiptRoutine()
	return nil
}
