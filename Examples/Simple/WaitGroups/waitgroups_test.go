package waitgroups

import (
	"sync"
	"testing"
	"time"
)

func Test_WG_A05_NegativeCounter(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()

	defer func() { _ = recover() }() // keep test suite running
	wg.Done()                        // counter goes negative => panic
}

func Test_WG_L09_WaitBlocksForever(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Wait() // leaks forever because Done never called
	}()

	time.Sleep(50 * time.Millisecond) // allow trace capture
}

func Test_WG_Possible_AddAfterWait(t *testing.T) {
	var wg sync.WaitGroup

	startWait := make(chan struct{})
	go func() {
		close(startWait)
		wg.Wait() // may return immediately, but Add happens concurrently afterwards (misuse pattern)
	}()

	<-startWait
	time.Sleep(1 * time.Millisecond)
	wg.Add(1)
	// We purposely never call Done to keep the trace interesting; you can add Done to avoid leaking.
	time.Sleep(50 * time.Millisecond)
}

func Test_WG_Leak_WorkerNeverDone_BlockedOnChan(t *testing.T) {
	var wg sync.WaitGroup
	block := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-block // blocks forever, so Done never runs
	}()

	go func() {
		wg.Wait() // leaks forever
	}()

	time.Sleep(50 * time.Millisecond)
}
