package scenarios

import (
	"sync"
	"testing"
	"time"
)

// ---------------- A00–A07: Absolute Bugs ----------------

func TestA00_UnknownPanic(t *testing.T) {
	panic("triggering unknown panic") // A00
}

func TestA01_SendOnClosed(t *testing.T) {
	ch := make(chan int)
	close(ch)
	ch <- 1 // A01
}

func TestA02_ReceiveOnClosed(t *testing.T) {
	ch := make(chan int)
	close(ch)
	_ = <-ch // A02 (Warning)
}

func TestA03_CloseOnClosed(t *testing.T) {
	ch := make(chan int)
	close(ch)
	close(ch) // A03
}

func TestA04_CloseOnNil(t *testing.T) {
	var ch chan int
	close(ch) // A04
}

func TestA05_NegativeWaitGroup(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()
	wg.Done() // A05
}

func TestA06_UnlockUnlocked(t *testing.T) {
	var mu sync.Mutex
	mu.Unlock() // A06
}

func TestA07_ConcurrentRecv(t *testing.T) {
	ch := make(chan int)
	go func() { _ = <-ch }()
	_ = <-ch // A07 (concurrent recv)
	close(ch)
}

// ---------------- P01–P03: Possible Bugs ----------------

func TestP01_PossibleSendOnClosed(t *testing.T) {
	ch := make(chan int)
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(ch)
	}()
	ch <- 42 // P01 (race with close)
}

func TestP02_PossibleRecvOnClosed(t *testing.T) {
	ch := make(chan int)
	go func() {
		ch <- 1
		close(ch)
	}()
	time.Sleep(10 * time.Millisecond)
	_ = <-ch // P02
	_ = <-ch // P02
}

func TestP03_PossibleNegativeWaitGroup(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		wg.Done() // P03
	}()
	wg.Wait()
}

// ---------------- L00–L10: Leaks ----------------

func TestL00_UnknownLeak(t *testing.T) {
	done := make(chan struct{})
	go func() { <-done }() // never signaled → L00
	time.Sleep(20 * time.Millisecond)
}

func TestL01_UnbufferedLeakWithPartner(t *testing.T) {
	ch := make(chan int)
	go func() {
		time.Sleep(10 * time.Millisecond)
		_ = <-ch // partner exists
	}()
	ch <- 1 // L01 (racing)
}

func TestL02_UnbufferedLeakNoPartner(t *testing.T) {
	ch := make(chan int)
	ch <- 1 // L02 → no receiver
}

func TestL03_BufferedLeakWithPartner(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1
	go func() {
		time.Sleep(10 * time.Millisecond)
		_ = <-ch // L03
	}()
}

func TestL04_BufferedLeakNoPartner(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1 // L04
}

func TestL05_LeakOnNilChan(t *testing.T) {
	var ch chan int
	ch <- 1 // L05 → nil channel send (blocks forever)
}

func TestL06_LeakOnSelectWithPartner(t *testing.T) {
	ch1 := make(chan int)
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch1 <- 42
	}()
	select {
	case <-ch1:
	case <-time.After(50 * time.Millisecond):
	}
}

func TestL07_LeakOnSelectWithoutPartner(t *testing.T) {
	var ch chan int // nil channel
	select {
	case <-ch:
	case <-time.After(50 * time.Millisecond):
	}
}

func TestL08_LeakOnMutex(t *testing.T) {
	var mu sync.Mutex
	mu.Lock()
	go func() {
		time.Sleep(20 * time.Millisecond)
		mu.Unlock()
	}()
	mu.Lock() // L08: blocked
}

func TestL09_LeakOnWaitGroup(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		time.Sleep(100 * time.Millisecond) // never calls Done()
	}()
	wg.Wait() // L09
}

func TestL10_LeakOnCond(t *testing.T) {
	mu := sync.Mutex{}
	cond := sync.NewCond(&mu)
	go func() {
		time.Sleep(50 * time.Millisecond)
		// no cond.Signal()
	}()
	mu.Lock()
	cond.Wait() // L10: waits forever
	mu.Unlock()
}

