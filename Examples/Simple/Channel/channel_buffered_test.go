package channel

import (
	"testing"
	"time"
)

// ================================
// Buffered channel scenarios
// ================================

// TestBufferedFillNoRead: buffered channel full, no reader
func TestBufferedFillNoRead(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2 // now full
	go func() {
		// no reader
		ch <- 3 // L04 (buffered leak no partner)
	}()
	time.Sleep(10 * time.Millisecond)
}

// TestBufferedDrainSlow: buffered channel drained slowly
func TestBufferedDrainSlow(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	go func() {
		time.Sleep(10 * time.Millisecond)
		<-ch
		<-ch
	}()
	// main returns quickly
}
