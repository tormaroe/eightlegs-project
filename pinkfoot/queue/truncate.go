package queue

import (
	"io"
	"log"
	"os"
	"sync"
	"time"
)

func (pq *PersistantQueue) truncate() {

	time.Sleep(30 * time.Second) // TODO: Decide some other way

	pq.stopWG = sync.WaitGroup{}
	pq.stopWG.Add(3)
	close(pq.stopChan)
	pq.stopWG.Wait()

	defer pq.start()

	readOffset := pq.currOffset
	log.Println("Current offset", readOffset)

	if readOffset == 0 {
		log.Println("No need to truncate")
		goto done
	}

	for _, v := range pq.waitReceiptList {
		if v.offset < readOffset {
			readOffset = v.offset
		}
	}

	log.Println("Current offset from stale pops", readOffset)

	if readOffset == 0 {
		log.Println("No need to truncate")
		goto done
	}

	pq.truncateTo(readOffset)

done:
	go pq.truncate()
}

func (pq *PersistantQueue) truncateTo(offset int64) {
	tmpFilename := pq.config.Persistance.DataFile + ".swp"
	log.Println("Creating temp file", tmpFilename)
	fTmp, err := os.OpenFile(tmpFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	f, err := os.OpenFile(pq.config.Persistance.DataFile, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = f.Seek(offset, 0)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("About to copy")
	n, err := io.Copy(fTmp, f)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("Copied %d bytes\n", n)
	f.Close()
	fTmp.Close()

	err = os.Rename(tmpFilename, pq.config.Persistance.DataFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Appendfile truncated")
}
