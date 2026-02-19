package waitgroups

import (
	"sync"
	"testing"
	"time"
)

// ================================
// WaitGroup scenarios
// ================================

// TestWGDoubleDone: wg.Done() called twice without Add
func TestWGDoubleDone(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()
	wg.Done() // Negative counter
}

// TestWGMissingDone: wg.Add() without corresponding Done
func TestWGMissingDone(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// forgot wg.Done()
		time.Sleep(20 * time.Millisecond)
	}()
	wg.Wait() // Leak
}

// TestWGNested: nested adds and dones
func TestWGNested(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		wg.Done()
		go func() {
			wg.Done()
		}()
	}()
	wg.Wait()
}
