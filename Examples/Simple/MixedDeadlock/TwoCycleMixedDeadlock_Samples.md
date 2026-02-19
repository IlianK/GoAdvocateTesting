# Mixed Deadlock 

Mixed deadlocks scenarios show how combining synchronization and communication primitives introduces new dependency cycles that cannot occur with either mechanism alone.


| Case         | Communication | Lock Relation                 | Description                            | Symmetry   |
| ------------ | ------------- | ----------------------------- | -------------------------------------- | ---------- |
| **MD2-1**    | Send–Receive  | Both inside critical sections | Both hold lock during communication    | Symmetric  |
| **MD2-2**    | Send–Receive  | Sender inside, receiver after | Sender waits for Receiver action & Receiver waits for held lock by Sender | Asymmetric |
| **MD2-3**    | Send–Receive  | Sender after, receiver inside | Receiver wait for Sender action & Receiver waits for held lock by Receiver | Asymmetric |


---


## MD2-1: Both Operations Inside Critical Sections

Here, both the sender and receiver execute their channel operation inside their respective CS (still holding the lock). The unbuffered variant leads to a deterministic deadlock, while the buffered variant may succeed, given the sender acquires the lock first.

**Symmetric MD**: `s ∈ CS(acq(m)s)` and `r ∈ CS(acq(m)r)`

### MD2-1U: Unbuffered Variant

```go
func TestMixedDeadlock_MD2_1U(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		m.Lock()
		c <- 1 // send inside CS (deterministic deadlock if sender first)
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c // receive inside CS (deterministic deadlock if receiver first)
		m.Unlock()
	}

	run2(sender, receiver)
}
```

#### Deadlocking Trace 1: Sender First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    req(x)
e2.    acq(x)  
e3.    snd(y)							<-- waits for receiver
e4.    					req(x)			<-- waits for lock x
```

#### Deadlocking Trace 2: Receiver First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e4.    					req(x)
e5.    					acq(x)  
e6.						rcv(y)			<-- waits for sender
e1.    req(x)							<-- waits for lock x
```

---

### MD2-1U: Buffered Variant

```go
func TestMixedDeadlock_MD2_1B(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered

	sender := func() {
		m.Lock()
		c <- 1 // send inside CS
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}
```

#### Working Trace: Sender First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    req(x)
e2.    acq(x)  
e3.    					req(x)		<-- waits for lock x
e4.    snd(y)						<-- sender fills buffer
e5.    rel(x)
e6.						acq(x)	
e7.   					rcv(y)		<-- receiver takes from buffer
e8.						rel(x)
```

#### Deadlocking Trace: Receiver First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e3.    					req(x)
e6.    					acq(x)  
e1.    req(x)						<-- waits for lock x
e7.						rcv(y)		<-- waits for empty buffer
```

---


## MD2-2: Sender Inside, Receiver After Critical Section
Here, the sender executes its channel operation while still holding the lock, while the receiver performs its channel operation immediately after leaving its critical section.

If the sender blocks on the unbuffered channel before releasing m, the receiver can never proceed further if it later needs the same lock, forming a partial dependency cycle.

**Asymmetric MD**: `s ∈ CS(acq(m)s)` and `CS(acq(m)r) ≫ r`


### MD2-2U: Unbuffered Variant
```go
func TestMixedDeadlock_MD2_2U(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		time.Sleep(50 * time.Millisecond) // sleep to let receiver complete PCS
		m.Lock()
		c <- 1 // unbuffered send inside CS (blocks until rcv)
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		<-c // receive after PCS
	}

	run2(sender, receiver)
}
```

#### Working Trace: Receiver First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    					req(x)
e2.    					acq(x)  
e3.    req(x)							<-- waits for lock x
e4.						rel(x)
e5.    					rcv(y)			<-- waits for sender
e6.    acq(x)
e7.    snd(y)							<-- sender & receiver synchronize
e8.    rel(x)
```

#### Deadlocking Trace: Sender First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e3.    req(x)
e6.    acq(x)  
e1.		 				req(x)			<-- waits for lock x
e7.    snd(y)							<-- waits for receiver
```



### MD2-2B: Buffered Variant
```go
func TestMixedDeadlock_MD2_2B(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		time.Sleep(50 * time.Millisecond) // sleep to let receiver complete PCS 
		m.Lock()
		c <- 1 // buffered send inside CS
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		<-c // receive after PCS
	}

	run2(sender, receiver)
}
```

#### Working Trace 1: Receiver First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    					req(x)
e2.    					acq(x)  
e3.	   req(x) 						<-- waits for lock x
e4.          			rel(x)			
e5.    					rcv(y)		<-- waits for buffer item
e6.    acq(x)
e7.    snd(y)						<-- sender fills buffer 
e8.    rel(x)
```

#### Working Trace 2: Sender First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e3.    req(x)
e6.    acq(x)  
e1.		 				req(x)		<-- wait for lock x
e7.    snd(y)						<-- sender fills buffer 
e8.    rel(x)
e2.    					acq(x)
e4.  					rel(x)
e5.						rcv(y)		<-- receiver takes item from buffer
```



---


## MD2-3: Sender After, Receiver Inside Critical Section
This case mirrors MD2-2, but roles are reversed.
The receiver holds the lock while performing the channel operation and the sender performs its channel operation after leaving its critical section.

**Asymmetric MD**: `CS(acq(m)s) ≫ s` and `r ∈ CS(acq(m)r)`


### MD2-3U: Unbuffered Variant
```go
func TestMixedDeadlock_MD2_3U(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		time.Sleep(50 * time.Millisecond)
		c <- 1 // send after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // sleep to let sender complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}
```

#### Working Trace: Sender First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    req(x)
e2.    acq(x)  
e3.						req(x)		<-- waits for lock x
e4.    rel(x)
e5.    snd(y)						<-- waits for receiver
e6.    					acq(x)
e7.    					rcv(y)		<-- sender & receiver synchronize
e8.    					rel(x)
```

#### Deadlocking Trace 2: Receiver First

```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    					req(x)
e2.    					acq(x)  
e3.	   req(x)						<-- waits for lock x
e4.    					rcv(y)		<-- waits for sender
```


### MD2-3B: Buffered Variant

```go
func TestMixedDeadlock_MD2_3B(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		c <- 1 // send after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // sleep to let sender complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}
```

#### Working Trace: Sender First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    req(x)
e2.    acq(x)  
e3.						req(x)		<-- waits for lock x
e4.    rel(x)
e5.    snd(y)						<-- sender fills buffer
e6.    					acq(x)		
e7.    					rcv(y)		<-- receiver takes item from buffer
e8.    					rel(x)
```

#### Deadlocking Trace 2: Receiver First

```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    					req(x)
e2.    					acq(x)  
e3.	   req(x)						<-- waits for lock x
e4.    					rcv(x)		<-- waits for buffer item
```

---


## MD-Close: Close–Receive Mixed Deadlock
Close variants are possible for all the above MD cases, but their behavior depends heavily on whether the channel is buffered or unbuffered.

The following example mirrors the MD-2-2 Case, meaning the close operation replaces a send operation. 

The closing goroutine may perform the close(c) after leaving a critical section, while another goroutine attempts to receive from the same channel while holding the lock.

Although close() normally unblocks receivers, timing and mutual exclusion via m can cause a circular dependency, similar to a send–receive case.

### MD2-CloseU: Unbuffered Variant

```go
func TestMixedDeadlock_MD_CloseU(t *testing.T) {
	var m sync.Mutex
	c := make(chan int) // unbuffered

	receiver := func() {
		m.Lock()
		<-c // receive inside CS (blocked until close)
		m.Unlock()
	}

	closer := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		time.Sleep(50 * time.Millisecond)
		close(c) // close after PCS
	}

	run2(receiver, closer)
}
```

#### Working Trace: Receiver First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    					req(x)
e2.    					acq(x)  
e3.    req(x)							<-- waits for lock x
e4.						rel(x)
e5.    					rcv(y)			<-- waits for close (or sender)
e6.    acq(x)
e7.    close(y)							<-- closer & receiver synchronize
e8.    rel(x)
```

#### Deadlocking Trace: Sender First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e3.    req(x)
e6.    acq(x)  
e1.		 				req(x)			<-- waits for lock x
e7.    close(y)							<-- waits for receiver
```

### MD2-CloseB: Buffered Variant

```go
func TestMixedDeadlock_MD_CloseU(t *testing.T) {
	var m sync.Mutex
	c := make(chan int) // unbuffered

	receiver := func() {
		m.Lock()
		<-c // receive inside CS (blocked until close)
		m.Unlock()
	}

	closer := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		time.Sleep(50 * time.Millisecond)
		close(c) // close after PCS
	}

	run2(receiver, closer)
}
```

#### Working Trace 1: Receiver First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    					req(x)
e2.    					acq(x)  
e3.	   req(x) 						<-- waits for lock x
e4.          			rel(x)			
e5.    					rcv(y)		<-- waits for close (or buffer item)
e6.    acq(x)
e7.    close(y)						<-- closer closes channel 
e8.    rel(x)
```

#### Working Trace 2: Sender First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e3.    req(x)
e6.    acq(x)  
e1.		 				req(x)		<-- waits for lock x
e7.    close(y)						<-- closer closes channel 
e8.    rel(x)
e2.    					acq(x)
e4.  					rel(x)
e5.						rcv(y)		<-- receiver gets zero value from channel
```
