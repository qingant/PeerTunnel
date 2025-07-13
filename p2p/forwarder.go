package p2p

import (
	"io"
	"log"
	"sync"
)

// Forwarder handles the bidirectional copying of data between two connections.
func Forwarder(from io.ReadWriteCloser, to io.ReadWriteCloser) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer from.Close()
		defer to.Close()
		_, err := io.Copy(from, to)
		if err != nil {
			log.Printf("Error copying from 'to' to 'from': %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		defer from.Close()
		defer to.Close()
		_, err := io.Copy(to, from)
		if err != nil {
			log.Printf("Error copying from 'from' to 'to': %v", err)
		}
	}()

	wg.Wait()
}
