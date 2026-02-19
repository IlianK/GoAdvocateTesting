package selectblock

import (
	"testing"
	"time"
)

// ================================
// Select scenarios
// ================================

// TestSelectNoPartner: select on nil channel without partner
func TestSelectNoPartner(t *testing.T) {
	var ch chan int // nil channel
	select {
	case <-ch:
	case <-time.After(10 * time.Millisecond):
	}
}

// TestSelectWithPartner: select with buffered partner
func TestSelectWithPartner(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 1
	select {
	case ch <- 2: // partner exists
	case <-time.After(10 * time.Millisecond):
	}
}

// TestSelectMultiple: multiple cases
func TestSelectMultiple(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)
	go func() { ch1 <- 1 }()
	time.Sleep(5 * time.Millisecond)
	select {
	case <-ch1:
	case <-ch2:
	case <-time.After(20 * time.Millisecond):
	}
}
