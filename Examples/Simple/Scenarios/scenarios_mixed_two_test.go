package scenarios

import (
	"testing"
	"time"
)

// ----------

func TestA01_Mixed(t *testing.T) {
	ch := make(chan int)
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(ch) // closes while another routine sends
	}()
	ch <- 1 // A01: may send on closed
}

func TestA02_Mixed(t *testing.T) {
	ch := make(chan int)
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(ch) // close before receive
	}()
	_ = <-ch // A02: might receive from closed
}

func TestA03_Mixed(t *testing.T) {
	ch := make(chan int)
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(ch)
	}()
	close(ch) // A03: may panic if closes cross
}

func TestA04_Mixed(t *testing.T) {
	var ch chan int
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch = make(chan int) // init too late
	}()
	close(ch) // A04: panic if nil when closed
}
