package scenarios

import (
	"sync"
	"testing"
	"time"
)

// ---------------- Mixed scenarios ----------------

// TestMixed1: A01 (send on closed channel) & L02 (unbuffered leak without partner)
func TestMixed1(t *testing.T) {
	// Unbuffered channel leak
	ch1 := make(chan int)
	go func() {
		// no receiver for ch1, this goroutine will leak
		ch1 <- 1 // L02
	}()

	// Closed channel panic
	ch2 := make(chan int)
	close(ch2)
	ch2 <- 5 // A01
}

// TestMixed2: A05 (negative wait group), A06 (unlock unlocked mutex) & L09 (waitgroup leak)
func TestMixed2(t *testing.T) {
	var wg sync.WaitGroup
	// Negative waitgroup counter
	wg.Done() // A05

	// Missing lock before unlock
	var mu sync.Mutex
	mu.Unlock() // A06

	// Leak on waitgroup (never Done for this Add)
	wg.Add(1)
	go func() {
		// busy work
		time.Sleep(20 * time.Millisecond)
		// forgot wg.Done()
	}()
	wg.Wait() // L09
}

// TestMixed3: P01 (possible send on closed), P03 (possible negative waitgroup), & L08 (mutex leak)
func TestMixed3(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	// Possible negative waitgroup inside goroutine
	go func() {
		wg.Done()
		wg.Done() // P03
	}()

	// Possible send on closed
	ch := make(chan int)
	go func() {
		close(ch)
	}()
	ch <- 10 // P01

	// Leak on mutex: lock twice without unlock
	var mu sync.Mutex
	mu.Lock()
	go func() {
		// no unlock for outer lock
		mu.Unlock()
	}()
	mu.Lock() // L08 if scheduling orders leak
	mu.Unlock()
}

// TestMixed4: A07 (concurrent recv), L03 (buffered leak with partner), P02 (possible recv on closed)
func TestMixed4(t *testing.T) {
	// Buffered channel with partner
	ch := make(chan int, 1)
	ch <- 1 // buffered
	go func() {
		_ = <-ch // partner exists (L03)
	}()

	// Concurrent receives on unbuffered channel
	u := make(chan int)
	go func() { u <- 2 }()
	go func() { _ = <-u }() // A07 concurrently
	_ = <-u

	// Possible recv on closed
	pc := make(chan int)
	close(pc)
	_ = <-pc // P02
}
