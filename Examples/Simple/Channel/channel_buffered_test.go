package channel

import (
	"sync"
	"testing"
	"time"
)

func Test_Buffered_Leak_SendBlocksOnFullBuffer(t *testing.T) {
	c := make(chan int, 1)

	// fill buffer
	c <- 1

	go func() {
		c <- 2 // blocks forever because buffer is full and nobody drains
	}()

	time.Sleep(50 * time.Millisecond)
}

func Test_Buffered_PossibleRecvOnClosed(t *testing.T) {
	c := make(chan int, 1)
	c <- 42 // ensure recv before close

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() { // receiver
		defer wg.Done()
		<-start
		_ = <-c // gets value first
	}()

	go func() { // closer (races)
		defer wg.Done()
		<-start
		time.Sleep(1 * time.Millisecond)
		close(c)
	}()

	close(start)
	wg.Wait()
	time.Sleep(10 * time.Millisecond)
}
