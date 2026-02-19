package channel

import (
	"advocate"
	"testing"
	"time"
)

// ================================
// Unbuffered channel scenarios
// ================================

// TestUnbufferedSendRecv: simple send/recv
func TestUnbufferedSendRecv(t *testing.T) {
	ch := make(chan int)
	go func() { ch <- 1 }()
	_ = <-ch
}

// TestUnbufferedLeakNoRecv: send blocks forever
func TestUnbufferedLeakNoRecv(t *testing.T) {
	ch := make(chan int)
	go func() {
		// no recv
		ch <- 1 // L02
	}()
	time.Sleep(10 * time.Millisecond)
}

// TestUnbufferedRecvNoSend: recv blocks forever
func TestUnbufferedRecvNoSend(t *testing.T) {
	// ======= Preamble Start =======
  advocate.InitTracing(5)
  defer advocate.FinishTracing()
  // ======= Preamble End =======
	ch := make(chan int)
	// no send
	_ = <-ch // L01 or deadlock
}
