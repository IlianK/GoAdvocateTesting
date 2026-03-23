package selectblock

import (
	"context"
	"testing"
	"time"
)

func Test_Select_Leak_WithoutPartner_Send(t *testing.T) {
	c := make(chan int) // unbuffered, but key is: no receiver exists

	go func() {
		select {
		case c <- 1: // blocks forever => select leak without partner
		}
	}()

	time.Sleep(50 * time.Millisecond)
}

func Test_Select_Leak_WithPartner_Send(t *testing.T) {
	c := make(chan int) // unbuffered
	blockRecv := make(chan struct{})

	// Possible partner exists but is prevented from running.
	go func() {
		<-blockRecv
		<-c
	}()

	go func() {
		select {
		case c <- 1: // blocks (receiver exists but is blocked elsewhere) => select leak with possible partner
		}
	}()

	time.Sleep(50 * time.Millisecond)
}

func Test_Select_Buffered_SendBlocksOnFullBuffer_WithPartner(t *testing.T) {
	c := make(chan int, 1)
	c <- 1 // buffer full

	blockRecv := make(chan struct{})
	go func() {
		<-blockRecv
		<-c // partner exists but never runs
	}()

	go func() {
		select {
		case c <- 2: // blocks forever because buffer is full and nobody drains
		}
	}()

	time.Sleep(50 * time.Millisecond)
}

func Test_Select_Leak_OnContextDone(t *testing.T) {
	ctx := context.Background() // Done() is nil channel

	go func() {
		select {
		case <-ctx.Done(): // blocks forever (nil channel) => context/select leak
		}
	}()

	time.Sleep(50 * time.Millisecond)
}
