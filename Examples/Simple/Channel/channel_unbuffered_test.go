package channel

import (
	"sync"
	"testing"
	"time"
)

func Test_Unbuffered_Leak_SendWithoutPartner(t *testing.T) {
	c := make(chan int) // unbuffered

	go func() {
		c <- 1 // blocks forever (no receiver)
	}()
	time.Sleep(50 * time.Millisecond)
}

func Test_Unbuffered_PossibleSendOnClosed(t *testing.T) {
	c := make(chan int) // unbuffered

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	// Receiver exists so send can succeed in some schedules
	go func() {
		defer wg.Done()
		<-start
		select {
		case <-c:
		case <-time.After(20 * time.Millisecond):
		}
	}()

	// Sender races with close
	go func() {
		defer wg.Done()
		<-start
		// Try to send; if close wins, this would panic
		select {
		case c <- 1:
		case <-time.After(10 * time.Millisecond):
		}
	}()

	close(start)

	// Close happens concurrently with the above operations
	time.Sleep(1 * time.Millisecond)
	close(c)

	wg.Wait()
}
