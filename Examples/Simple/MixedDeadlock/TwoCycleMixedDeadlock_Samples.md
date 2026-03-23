# TwoCycleMixedDeadlock - Testing

This document describes the mixed deadlock test cases defined in `TwoCycleMixedDeadlock_test.go` and attempts their classification as defined in `doc/TwoCycleMixedDeadlock_Definition.md`

These tests cover **Two-cycle MD2-X patterns**
that arise from lock-channel interaction under different structural and temporal arrangements. It is described:
- whether a **working trace** exists,
- whether a **deadlocking trace** exists,
- and whether a **predictive reordering** can expose a mixed deadlock.

There are also test cases included to test the precision of the analysis and are scenarios that:
- Should **not** be classified as MD = **false negatives (`_FN`)**
- Expose the detectability limits = **false positives (`_FP`)** 


---

## Notation and Classification
Each test corresponds to one theoretical MD-case:

- **MD2-1**: **Both operations inside critical sections**  
Sender and receiver perform their channel operations while holding the shared lock  
  (**symmetric**).

- **MD2-2**: **Sender inside CS, receiver after CS**  
  Sender performs channel operation while holding shared lock, Receiver performs channel operation after completing their CS involving the shared lock.   
  (**asymmetric**).

- **MD2-3**: **Sender after CS, receiver inside CS**  
  Receiver performs channel operation while holding shared lock, Sender performs channel operation after completing their CS involving the shared lock.  
  (**asymmetric**).

---

### Close Variants
**MD2-X-Close** Variants where a `close(c)` operation replaces a `send(c)` operation. Close–receive interactions are treated as communication pairs for the purpose of mixed deadlock detection.

---

## Channel Variants
Each MD case is tested with different channel configurations:
- `_U`: Unbuffered channel
- `_B`: Buffered channel

---

## Lock Variants
The test suite distinguishes between different lock interaction patterns:
- **READ / READ** – both goroutines acquire a read lock (`RLock`)
- **READ / WRITE** – one goroutine holds a read lock, the other a write lock
- **WRITE / WRITE** – both goroutines acquire a write lock (`Lock`)

---

## Predictive Analysis 
All tests are analyzed under a **trace-based predictive analysis model**:
- No static program analysis is assumed.

- Dependencies are constructed exclusively from the observed execution trace. As a consequence, some semantically valid mixed-deadlock cases may not be detectable if the execution immediately deadlocks and no working trace exists.
Such cases are annotated with the suffix `_NoDec` (not detectable).

- The `_NoDec` category is explicitly distinguished from `_FN` cases, where a working trace does exist, but the mixed deadlock is still not detected due to inherent limitations of the trace-based analysis (e.g., non standard critical sections).

- In several tests, `sleep()` statements are intentionally used to influence scheduling and **allow the program to produce a working trace**. This is necessary to make communication events observable and thereby enable predictive analysis. The use of `sleep()` does not change the semantics of the program, but only affects which executions become observable in the trace.

- Potential reorderings are also restricted by **WMHB (Weak Must-Happen-Before)** constraints, which intentionally exclude some reorderings that would lead to a mixed deadlock, in order to avoid false positives. These also include a suffic `_WMHB`.


---


## Basic MD Test Cases

### `TestMixedDeadlock_MD2_1U_NoDec`
Both sender and receiver perform their channel operations inside their `CS` guarded by the same mutex `m` and the shared channel `c` is `unbuffered`

Because the channel is unbuffered, a send and receive must synchronize. However, mutual exclusion via `m` prevents the two goroutines from ever reaching the channel operation concurrently.

As a result, every possible execution leads to an **immediate deadlock**, and **no working trace exists**. This represents an **Actual Mixed Deadlock**, which is currently not detected by the analysis.


```go
func TestMixedDeadlock_MD2_1U((t *testing.T) {
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

#### Deadlocking Trace 1: Sender First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    acq(x)?
e2.    acq(x)  
e3.    snd(y)?										
e4.    					acq(x)?		
```

#### Deadlocking Trace 2: Receiver First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e1.    					acq(x)?
e2.    					acq(x)  
e3.    					rcv(y)?										
e4.    acq(x)?		
```

---

### `TestMixedDeadlock_MD2_1B`
Both sender and receiver perform their channel operations **inside their critical sections** guarded by the same mutex `m`.  
The shared channel `c` is **buffered with capacity 1**.

The test has **both working and deadlocking executions** and represents a **Potential Mixed Deadlock (P06)**.

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
		time.Sleep(50 * time.Millisecond) // let sender go first
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
e1.    acq(x)?
e2.    acq(x)
e3.    					acq(x)?       
e4.    snd(y)?
e5.    snd(y)
e6.    rel(x)
e7.    					acq(x)
e8.    					rcv(y)?
e9.    					rcv(y)
e10.   					rel(x)
```

The buffered send completes while holding `m`, and the receiver later consumes the buffered value after acquiring the lock. The communication pair `(snd(y), rcv(y))` is observable in the trace, **enabling predictive reordering** to the following deadlocking trace.


#### Deadlocking Trace: Receiver First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e3.    					acq(x)?
e7.    					acq(x)
e1.    acq(x)?                        
e8.    					rcv(y)?      
```

Here, the receiver blocks on the receive operation while holding `m`, preventing the sender from acquiring the lock and performing the send, resulting in a **deadlock**.



---


### `TestMixedDeadlock_MD2_2U`
The sender performs its channel operation inside its `CS` guarded by the mutex `m`, while the receiver performs its channel operation after completing its `CS`, corresponding to the `MD2-2` pattern. The shared channel `c` is **unbuffered**.

The test has **both working and deadlocking executions** and represents a **Potential Mixed Deadlock (P06)**.


```go
func TestMixedDeadlock_MD2_2U(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		time.Sleep(50 * time.Millisecond) // let receiver complete PCS
		m.Lock()
		c <- 1 // send inside CS
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		<-c // receive after PCS
	}

	run2(sender, receiver)
}
```

#### Working Trace: Receiver First
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.                             acq(x)?
e2.                             acq(x)
e3.    acq(x)?
e4.                             rel(x)
e5.    acq(x)
e6.                             rcv(y)?
e7.    snd(y)?
e8.                             rcv(y)
e9.    snd(y)
e10.   rel(x)
```

The receiver completes its `PCS` before the sender enters its `CS`.
The **send and receive synchronize successfully**, and the communication pair (`snd(y), rcv(y)`) is observable in the trace, enabling predictive reordering.


#### Deadlocking Trace: Sender First
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e3.    acq(x)?
e5.    acq(x)
e7.    snd(y)?
e1.                             acq(x)?
```
Here, the sender blocks on the unbuffered send while holding `m`.
The receiver cannot acquire the mutex to complete its `PCS`, resulting in a deadlock.


---


### `TestMixedDeadlock_MD2_2B_FP`
The sender performs its channel operation inside its `CS` guarded by mutex `m`, while the receiver performs its channel operation after completing its `PCS`, corresponding to the `MD2-2` pattern.
The shared channel `c` is buffered with capacity `1`.

The test has **both only working executions** and represents a `_FP` test case, which is currently falsely detected as as `MD` by the analysis.

```go
func TestMixedDeadlock_MD2_2B(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		// no sleep, always non-blocking due to buffer
		m.Lock()
		c <- 1 // send inside CS
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		<-c // receive after PCS
	}

	run2(sender, receiver)
}
```

#### Working Trace 1: Receiver First
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.                             acq(x)?
e2.                             acq(x)
e3.    acq(x)?
e4.                             rel(x)
e5.    acq(x)
e6.                             rcv(y)?
e7.    snd(y)?
e8.    snd(y)
e9.                             rcv(y)
e10.   rel(x)
```

The receiver completes its `PCS` before the sender enters its `CS`.
The sender then performs a non-blocking buffered send while holding `m`.
The communication pair `(snd(y), rcv(y))` is fully observable in the trace, enabling predictive analysis.


#### Working Trace 2: Sender First
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e3.    acq(x)?
e5.    acq(x)
e1.                             acq(x)?
e7.    snd(y)?
e8.    snd(y)
e10.   rel(x)
e2.                             acq(x)
e4.                             rel(x)
e6.                             rcv(y)?
e9.                             rcv(y)
```

The sender completes its `CS` by sending the non-blocking buffered send while holding `m`. Then the receiver enters its `PCS` and after completing it, performs its receive. The communication pair `(snd(y), rcv(y))` is again fully observable.


---


### `TestMixedDeadlock_MD2_3U`
The receiver performs its channel operation inside its `CS` guarded by mutex `m`, while the sender performs its channel operation after completing its `CS`, corresponding to the `MD2-3` pattern.
The shared channel `c` is **unbuffered**.

The test has **both working and deadlocking executions** and represents a Potential Mixed Deadlock (P06).

```go
func TestMixedDeadlock_MD2_3U(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		c <- 1 // send after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) //let sender complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}
```

#### Working Trace: Sender First
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    acq(x)?
e2.    acq(x)
e3.                             acq(x)?
e4.    rel(x)
e5.    snd(y)?
e6.                             acq(x)
e7.                             rcv(y)?
e8.    snd(y)
e9.                             rcv(y)
e10.                            rel(x)
```

The sender completes its `PCS` before performing the send.
The receiver then acquires the mutex and performs the receive inside its `CS`.
The **unbuffered send and receive synchronize successfully**, and the communication pair `(snd(y), rcv(y))` is fully observable in the trace, enabling predictive analysis.


#### Deadlocking Trace 2: Receiver First

```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e3.                             acq(x)?
e6.                             acq(x)
e7.                             rcv(y)?   
e1.    acq(x)?                  
   
```
Here, the receiver acquires the mutex first and blocks on the unbuffered receive while holding `m`, waiting to synchronize.
The sender is unable to acquire the mutex to complete its `PCS`, and therefore cannot reach the send operation, resulting in an **deadlock**.



### `TestMixedDeadlock_MD2_3B`
The receiver performs its channel operation inside its `CS` guarded by mutex `m`, while the sender performs its channel operation after completing its `PCS`, corresponding to the `MD2-3` pattern.
The shared channel `c` is buffered with capacity `1`.

The test has **both working and deadlocking executions** and represents a Potential Mixed Deadlock (P06).

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
		time.Sleep(50 * time.Millisecond) // let sender complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}
```

#### Working Trace: Sender First
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    acq(x)?
e2.    acq(x)
e3.                             acq(x)?
e4.    rel(x)
e5.                             acq(x)
e6.    snd(y)?          
e7.    snd(y)
e8.                             rcv(y)?
e9.                             rcv(y)
e10.                            rel(x)
```
The sender completes its `PCS` and performs a non-blocking buffered send. The receiver later acquires the mutex and performs the receive inside its `CS`. The communication pair `(snd(y), rcv(y))` is fully observable in the trace, enabling predictive analysis.


#### Deadlocking Trace 2: Receiver First

```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e3.                             acq(x)?
e5.                             acq(x)
e8.                             rcv(y)?
e1.    acq(x)?
```

Like in `2_3U`, if the receiver acquires the mutex first it blocks on the receive. But here the reason is, that there is no buffered item to consume. The sender is unable to acquire the mutex to complete its `PCS`, and therefore cannot reach the send operation, resulting in an **deadlock**.




---
---


## `TestMixedDeadlock_MD2_*_Close*`-Variants
`close()` is a **non-blocking** operation in Go. This changes the mixed deadlock analysis:
- `close(ch)` returns immediately, regardless of channel state
- A receive on a closed channel returns the zero value immediately
- Only receives that are **already waiting** are affected by `close()`


### `TestMixedDeadlock_MD2_1_Close*`
Both closer and receiver perform their channel operations inside their `CS` guarded by the same mutex `m`.

The test has **both working and deadlocking executions** and represents a **Potential Mixed Deadlock (P06)**.

```go
func TestMixedDeadlock_MD2_1_CloseU(t *testing.T) {
	var m sync.Mutex
	c := make(chan int) // unbuffered

	closer := func() {
		m.Lock()
		close(c) // close inside CS
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // let closer go first
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(closer, receiver)
}
```

```go
func TestMixedDeadlock_MD2_1_CloseB(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered
	// ...
```

#### Working Trace 1: Closer First
```
       T1 (Closer)      T2 (Receiver)
----------------------------------------
e1.    acq(x)?
e2.    acq(x)  
e3.    					acq(x)?	
e4.    close(y)					
e5.    rel(x)			
e6.  					acq(x)
e7. 					rcv(y?)
e8.						rcv(y)
e9.					    rel(x)
```

The closer acquires the shared mutex and closes the channel within its `CS`.
The receiver is blocked on trying to acquire the shared mutex.
Since the close is **non-blocking** it releases the lock and the receiver may acquire it.

#### Deadlocking Trace 2: Receiver First
```
       T1 (Closer)      T2 (Receiver)
----------------------------------------
e3.    					acq(x)?
e6.    					acq(x)  
e7.    					rcv(y)?										
e1.    acq(x)?		
```

The receiver acquires the shared mutex and is blocking on its channel operation. The receiver is waiting on a synchronize (`unbuffered` case) or a channel item (`buffered` case) to consume. Since the receiver is blocked on receive and the closer is blocked on trying to acquire the shared mutex, this causes a deadlock.


---


### `TestMixedDeadlock_MD_2_2Close*`
The closer performs its channel operation inside its `CS` guarded by the mutex `m`, while the receiver performs its channel operation after completing its `CS`, corresponding to the `MD2-2` pattern. 

The test has **only working executions**, beacuse the close is non-blocking and the in the `MD2-2` pattern the receiver has a `PCS`, that always completes and then unlocks the shared mutex in case of the receiver being first in execution order.


```go
func TestMixedDeadlock_MD_2_2CloseU(t *testing.T) {
	var m sync.Mutex
	c := make(chan int) // unbuffered

	closer := func() {
		// no sleep, recv is blocking until close
		m.Lock()
		close(c) // close in CS
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		<-c // recv with PCS
	}

	run2(receiver, closer)
}
```

```go
func TestMixedDeadlock_MD_2_2CloseB(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered
	// ...
```

#### Working Trace 1: Receiver First
```
       T1 (Closer)              T2 (Receiver)
-----------------------------------------------
e1.                             acq(x)?
e2.                             acq(x)
e3.    acq(x)?
e4.                             rel(x)
e5.    acq(x)
e6.                             rcv(y)?
e7.    close(y)
e8.    							rcv(y)	<- zero val
e9.    rel(x)
```

The receiver completes its `PCS` and is then waiting on a synchronize (`unbuffered` case) or a channel item (`buffered` case). The closer enters its `CS`. The **non-blocking close** closes the channel and thus unblocks the receiver, letting it receive a zero val. 


#### Working Trace 2: Closer First
```
       T1 (Closer)              T2 (Receiver)
-----------------------------------------------
e3.    acq(x)?
e5.    acq(x)
e1.                             acq(x)?
e7.    close(y)
e9.    rel(x)
e2.                             acq(x)
e4.                             rel(x)
e6.                             rcv(y)?
e8.    							rcv(y)	<- zero val

```
If the closer gets the mutex first it successfully closes the channel again, then letting the receiver complete the `PCS` and receive the zero value. 


---


### `TestMixedDeadlock_MD_2_3Close*`
The receiver performs its channel operation inside its `CS` guarded by mutex `m`, while the closer performs its channel operation after completing its `CS`, corresponding to the `MD2-3` pattern.

The test has **both working and deadlocking executions** and represents a **Potential Mixed Deadlock (P06)**.


```go
func TestMixedDeadlock_MD_2_3CloseU(t *testing.T) {
	var m sync.Mutex
	c := make(chan int) // unbuffered

	closer := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		close(c) // close after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // let closer complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(receiver, closer)
}
```

```go
func TestMixedDeadlock_MD_2_3CloseB(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered
	// ...
```


#### Working Trace 1: Closer First
```
       T1 (Closer)              T2 (Receiver)
-----------------------------------------------
e1.    acq(x)?
e2.    acq(x)
e3.                             acq(x)?
e4.    rel(x)
e5.    close(y)
e6.                             acq(x)
e7.                             rcv(y)?  <- zero val
e8.                             rcv(y)
e9.                             rel(x)
```

The closer completes its `PCS` before closing the channel.
The receiver then acquires the mutex and performs the receive inside its `CS`, receiving a zero value, since the channel is closed.


#### Deadlocking Trace 2: Receiver First

```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e3.                             acq(x)?
e6.                             acq(x)
e7.                             rcv(y)?   
e1.    acq(x)?                  
   
```
Here, the receiver acquires the mutex first and blocks on the receive while holding `m`, waiting on a synchronize (`unbuffered` case) or a channel item (`buffered` case).
The closer is unable to acquire the mutex to complete its `PCS`, and therefore cannot reach the close operation, which would unblock the receiver, resulting in an **deadlock**.



---
---



## `TestMixedDeadlock_MD2_**_RW*`-Variants
These tests cover Mixed Deadlock `MD2-X` patterns where the shared lock is a `sync.RWMutex` and the interaction is between a writer (`Lock`) and a reader (`RLock`).

The structural MD patterns are **identical to the mutex-based cases**.
However, `RWMutex` introduces asymmetric blocking semantics:
- A writer blocks while **any reader holds `RLock`**
- A reader blocks only **if a writer already holds `Lock`**


---


### `TestMixedDeadlock_MD_2_1B_RW`

```go
func TestMixedDeadlock_MD_2_1B_RW(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int, 1)

	writer := func() {
		rw.Lock()
		c <- 1 // send inside CS
		rw.Unlock()
	}

	reader := func() {
		time.Sleep(50 * time.Millisecond)
		rw.RLock()
		<-c // receive inside CS
		rw.RUnlock()
	}

	run2(reader, writer)
}

```

A communication pair is observable, enabling predictive analysis.
A reordered execution where the reader acquires the read lock first leads to a writer-blocked deadlock, yielding a Potential Mixed Deadlock (P06).

#### Working Trace: Writer First
```
       T1 (Writer)              T2 (Reader)
-----------------------------------------------
e1.    acqW(x)?
e2.    acqW(x)
e3.                             acqR(x)?
e4.    snd(y)?
e5.    snd(y)
e6.    relW(x)
e7.                             acqR(x)
e8.                             rcv(y)?
e9.                             rcv(y)
e10.                            relR(x)
```

#### Deadlocking Trace: Reader First
```
       T1 (Sender)      T2 (Receiver)
----------------------------------------
e3.                             acqR(x)?
e7.                             acqR(x)
e8.                             rcv(y)?
e1.    acqW(x)?	
```


---


### `TestRWMutex_DoubleRLockCounter_MD2_1*`

This test uses a `RWMutex` with multiple read-lock acquisitions and a buffered channel.

Both goroutines perform their channel operations inside a `CS` protected by the same `rw`, but the reader goroutine holds the lock via two nested `RLock()` calls, creating a **non-standard critical-section** structure.

The shared channel `c` is buffered with capacity `1`, ensuring that a working trace exists.

```
func TestRWMutex_DoubleRLockCounter_MD2_1B(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int, 1) // buffered für MD2-1

	reader := func() {
		time.Sleep(50 * time.Millisecond) // let sender go first
		rw.RLock()
		rw.RLock()
		rw.RUnlock() // Reader Count: 2 -> 1
		time.Sleep(30 * time.Millisecond)
		// Receive still in CS
		<-c

		rw.RUnlock()
	}

	sender := func() {
		rw.Lock()
		c <- 1 // Send in CS
		rw.Unlock()
	}

	run2(reader, sender)
}
```

#### Working Trace: Sender First

```
       T1 (Reader)               T2 (Sender)
-----------------------------------------------
e1.                             acqW(x)?
e2.                             acqW(x)
e3.    acqR(x)?
e4.                             snd(y)?
e5.                             snd(y)
e6.                             relW(x)
e7.    acqR(x)
e8.    acqR(x)?		
e9.    acqR(x)
e10.   relR(x)
e11.   rcv(y)?
e12.   rcv(y)
e13.   relR(x)
```

The sender acquires the write lock and performs a buffered send inside its `CS`.
The reader then acquires the read lock twice, partially releases it once, and performs the receive while still holding the lock.
The communication pair `(snd(y), rcv(y))` is observable, enabling predictive analysis.

### Deadlocking Trace: Sender First
```
       T1 (Reader)               T2 (Sender)
-----------------------------------------------
e3.    acqR(x)?
e7.    acqR(x)
e8.    acqR(x)?
e9.    acqR(x)
e11.   rcv(y)?        
e1.                             acqW(x)?   
```

In the reordered execution, the reader acquires the read lock twice and blocks on the receive while still holding the lock (reader `count > 0`).
The sender attempts to acquire the write lock but is blocked until all readers have released the lock, which never happens because the reader is waiting on the channel.

The earlier implementation of MD-detection did not include a read-counter, and thus produced a `FN`.


---
---


## Multiple CS MD Tests
The following tests include scenarios where a single goroutine executes more than one `CS` involving the same mutex, to ensure that the mixed-deadlock analysis correctly reasons about which specific `CS` a channel operation belongs to.

A naive analysis previously incorrectly assumed that once a goroutine holds a lock at some point, all channel operations in that goroutine are implicitly associated with that lock.


---


### `TestMixedDeadlock_DoubleCS_Send`

In this test, the sender executes two distinct critical sections, but the send occurs in only one of them.

```
func TestMixedDeadlock_DoubleCS_Send(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		c <- 1
		m.Unlock()
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c
		m.Unlock()
	}

	run2(sender, receiver)
}
```

#### Working Trace: Sender First & Sender again / Receiver after

```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    acq(x)?
e2.    acq(x)
e3.                             acq(x)?
e4.    snd(y)?
e5.    snd(y)
e6.    rel(x)
e7.    acq(x)?
e8.    acq(x)
e9.    rel(x)
e10.                            acq(x)
e11.                            rcv(y)?
e12.                            rcv(y)
e13.                            rel(x)
```

```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    acq(x)?
e2.    acq(x)
e3.                             acq(x)?
e4.    snd(y)?
e5.    snd(y)
e6.    rel(x)
e10.                            acq(x)
e11.                            rcv(y)?
e12.                            rcv(y)
e13.                            rel(x)
e6.    acq(x)?
e7.    acq(x)
e8.    rel(x)
```

The send belongs to the first `CS_1` from `e1-e5` 
The second `CS_2` from `e6-e8` is independent and does not participate in any lock–channel dependency, which is why once the Sender completes its first `CS_1` it will be successfull whether the Sender acquires the lock directly again for the `sleep` or if the receiver acquires it.

For these examples the notation will be simplified, meaning for test `TestMixedDeadlock_DoubleCS_Send` we have the working traces:
```

       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    CS_1	 // with snd
e2.	   CS_2   
e3.								CS
```

or

```

       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    CS_1	// with snd
e3.								CS
e2.	   CS_2   
```


#### Deadlocking Trace: Receiver First 
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e3.                             acq(x)?
e10.                            acq(x)
e1.    acq(x)?
e11.                            rcv(y)?
```

or simplified as:

```

       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e3.								CS?
e1.    CS_1?
```

If the receiver acquires first the same pattern can be seen as in previous tests. Due to the empty buffer (or in case of unbuffered channel, due to the missing communication partner to synchronize) a deadlock occurs.


---


### `TestMixedDeadlock_DoubleCS_Send_2`
For `TestMixedDeadlock_DoubleCS_Send_2` the problem just delays but produces the same pattern in the end. 

```
func TestMixedDeadlock_DoubleCS_Send_2(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		m.Lock()
		c <- 1
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c
		m.Unlock()
	}

	run2(sender, receiver)
}
```

#### Working Trace: Sender First & Sender again / Receiver after
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    CS_1
e2.	   CS_2 //with snd
e3.								CS
```


#### Deadlocking Trace: Receiver First
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e3.								CS?
e1.    CS_1?
```

or

```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    CS_1
e3.								CS
e2.	   CS_2?
```


---



### `TestMixedDeadlock_DoubleCS_Recv`
In these tests, the receiver executes two distinct critical sections, but the receive occurs in only one of them.
The same pattern as in `Double_CS_Send` applies:

```
func TestMixedDeadlock_DoubleCS_Recv(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		c <- 1
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c
		m.Unlock()
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	run2(sender, receiver)
}
```


#### Working Trace: Sender First 
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    CS
e2.	   							CS_1 //with rcv
e3.								CS_2
```

#### Deadlocking Trace: Receiver First 
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e2.	   							CS_1? //with rcv
e1.    CS?
```


---


### `TestMixedDeadlock_DoubleCS_Recv_2`
```
func TestMixedDeadlock_DoubleCS_Recv_2(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		c <- 1
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		m.Lock()
		<-c
		m.Unlock()
	}

	run2(sender, receiver)
}
```

#### Working Trace: Sender First 
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    CS
e2.	   							CS_1 
e3.								CS_2 // with rcv
```

or

```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e2.	   							CS_1 
e1.    CS
e3.								CS_2 // with rcv
```

#### Deadlocking Trace: Receiver First 
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e2.	   							CS_1 
e3.								CS_2 // with rcv
e1.    CS
```


---
---



## False Positive Tests
These tests were used to confirm that selection of communication candidates for a mixed deadlock scenario were following the requirements of:
- A shared lock
- A shared channel
- At least one channel operation overlapping with a critical section
- A cyclic dependency between lock acquisition and channel synchronization


### `TestMixedDeadlock_MD_Read_FP`
Both goroutines acquire a read lock (`RLock`) on the same `sync.RWMutex` and perform their channel operations inside their `CS`, corresponding structurally to an `MD2-1` pattern with a `READ/READ` lock interaction. The shared channel `c` is unbuffered.

Despite the superficial similarity to other `MD2-1` cases, this test is not a mixed deadlock.

```
func TestMixedDeadlock_MD_Read_FP(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int)

	reader_1 := func() {
		rw.RLock()
		<-c
		rw.RUnlock()
	}

	reader_2 := func() {
		rw.RLock()
		c <- 1
		rw.RUnlock()
	}

	run2(reader_1, reader_2)
}
```

#### Working Trace: Either Reader First

```
       T1 (Reader 1)             T2 (Reader 2)
-----------------------------------------------
e1.    acqR(x)?
e2.    acqR(x)
e3.                             acqR(x)?
e4.                             acqR(x)
e5.                             snd(y)?
e6.                             snd(y)
e7.    rcv(y)?
e8.    rcv(y)
e9.    relR(x)
e10.                            relR(x)
```
Both goroutines are able to simultaneously hold the read lock, as `RWMutex` allows multiple readers to coexist.
The unbuffered send and receive synchronize successfully while both goroutines remain inside their `CS`.

---


### `TestMixedDeadlock_No_MD_BeforeCS`
Both goroutines perform their channel operations before entering any `CS`.
Although a shared mutex m and an unbuffered channel `c` are used, no channel operation occurs inside a `CS`, and therefore no mixed lock–channel dependency can arise.

```
func TestMixedDeadlock_No_MD_BeforeCS(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		c <- 1 // before CS
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	receiver := func() {
		<-c // before CS
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	run2(sender, receiver)
}
```

#### Working Trace

```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    snd(y)?
e2.                             rcv(y)?
e3.                             rcv(y)
e4.    snd(y)
e5.    acq(x)?
e6.    acq(x)
e7.    rel(x)
e8.                             acq(x)?
e9.                             acq(x)
e10.                            rel(x)
```

The send and receive synchronize before any lock is acquired.
All lock operations occur after communication has completed, so no circular dependency involving the channel and the lock is possible.


### `TestMixedDeadlock_No_MD_AfterPCS`
Both goroutines perform their channel operations after completing their `PCS`.
Although the lock `m` is used by both goroutines, it does not overlap with any channel operation.

```
func TestMixedDeadlock_No_MD_AfterPCS(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		c <- 1 // after PCS
	}

	receiver := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		<-c // after PCS
	}

	run2(sender, receiver)
}
```

#### Working Trace
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    acq(x)?
e2.    acq(x)
e3.    rel(x)
e4.                             acq(x)?
e5.                             acq(x)
e6.                             rel(x)
e7.    snd(y)?
e8.                             rcv(y)?
e9.                             rcv(y)
e10.   snd(y)
```
Both goroutines release the mutex before performing their channel operations. As a result, the channel communication is completely decoupled from mutual exclusion, and no lock–channel cycle can form.


### `TestMixedDeadlock_No_MD_DifferentLocks`
The sender and receiver use different mutexes to guard their respective channel operations.
Although both channel operations occur inside `CS`, the lock dependency is absent.

```
func TestMixedDeadlock_No_MD_DifferentLocks(t *testing.T) {
	var m1, m2 sync.Mutex
	c := make(chan int)

	sender := func() {
		m1.Lock()
		c <- 1
		m1.Unlock()
	}

	receiver := func() {
		m2.Lock()
		<-c
		m2.Unlock()
	}

	run2(sender, receiver)
}
```

#### Working Trace
```
       T1 (Sender)              T2 (Receiver)
-----------------------------------------------
e1.    acq(x1)?
e2.    acq(x1)
e3.                             acq(x2)?
e4.                             acq(x2)
e5.    snd(y)?
e6.                             rcv(y)?
e7.                             rcv(y)
e8.    snd(y)
e9.    rel(x1)
e10.                            rel(x2)
```

Although the channel communication synchronizes successfully, there is no shared lock between the two goroutines.
Therefore, no cyclic dependency involving a common mutex can be constructed.


### `TestMixedDeadlock_No_MD_ForkMustOrder`
This test ensures that fork-based must-happen-before (MHB) constraints prevent false mixed-deadlock reports.
Although the structure superficially resembles an `MD2-1` pattern, the execution order imposed by the fork operation makes any deadlocking reordering illegal.

```
func TestMixedDeadlock_No_MD_ForkMustOrder(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)
	done := make(chan struct{})

	// Parent acquires BEFORE forking
	m.Lock()

	// Child after
	go func() {
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
		close(done)
	}()

	// Parent sends while STILL in CS (would look like MD2-1 if reordering was allowed),
	// but acquires are must-ordered by the fork, so the detector should NOT flag it.

	c <- 1 // send inside CS, non-blocking due to buffer

	m.Unlock()
	<-done
}
```

#### Working Trace (Must Ordered)

```
       T1 (Parent)              T2 (Child)
-----------------------------------------------
e1.    acq(x)?
e2.    acq(x)
e3.    fork(T2)
e4.    snd(y)?
e5.    snd(y)
e6.    rel(x)
e7.                             acq(x)?
e8.                             acq(x)
e9.                             rcv(y)?
e10.                            rcv(y)
e11.                            rel(x)
```

The parent goroutine acquires the mutex before forking the child.
As a result, the child’s `acq(x)` is must-happen-after the parent’s `acq(x)` and cannot be reordered before it.



---
---



## False Negatives 
The following tests demonstrate that recording only a single (first or last) dependency per goroutine is insufficient for sound mixed-deadlock detection.

### `TestMixedDeadlock_MultiDep_LastOnly_FN`
In this test, each goroutine creates multiple dependencies, but only an earlier one participates in a cycle.

If the analysis records only the last dependency per goroutine, the cyclic dependency is lost.

```
func TestMixedDeadlock_MultiDep_LastOnly_FN(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)
	s := make(chan int, 1)

	receiver := func() { // G2
		// D3
		m.Lock()
		<-c
		m.Unlock()

		s <- 1
		<-s

		// D4
		m.Lock()
		<-c
		m.Unlock()
	}

	sender := func() { // G1
		// D1
		m.Lock()
		c <- 1
		m.Unlock()

		<-s

		// D2
		m.Lock()
		s <- 1
		c <- 1
		m.Unlock()
	}

	run2(receiver, sender)
}
```

#### Working Trace: Sender - Receiver - Sender - Receiver

```
       G1 (Sender)              G2 (Receiver)
-----------------------------------------------
e0.    							 rcv(s)?  // forces G1 to go first
e1.    acq(x)?
e2.    acq(x)
e3.    snd(c)?
e4.    snd(c)
e5.    rel(x)        // D1
e6.    snd(s)?		 // sync with receiver (trigger empty buffer)
e7.    snd(s)

e8.    							 rcv(s)
e9.                              acq(x)?
e10.                             acq(x)
e11.                             rcv(c)?
e12.                             rcv(c)
e13.                             rel(x)        // D3

e14.    acq(x)?
e15.    acq(x)
e16.    snd(s)?
e17.    snd(s)
e18.    snd(c)?
e19.    snd(c)
e20.    rel(x)        // D2

e21.                            acq(x)?
e22.                            acq(x)
e23.                            rcv(c)?
e24.                            rcv(c)
e25.                            rel(x)        // D4
```



#### Deadlocking Trace: Sender - Receiver - Receiver

```
       G1 (Sender)              G2 (Receiver)
-----------------------------------------------
e0.    							 rcv(s)?  // forces G1 to go first
e1.    acq(x)?
e2.    acq(x)
e3.    snd(c)?
e4.    snd(c)
e5.    rel(x)        // D1
e6.    snd(s)?		 // sync with receiver (trigger empty buffer)
e7.    snd(s)

e8.    							 rcv(s)
e9.                              acq(x)?
e10.                             acq(x)
e11.                             rcv(c)?
e12.                             rcv(c)
e13.                             rel(x)        // D3

e21.                             acq(x)?
e22.                             acq(x)
e23.                             rcv(c)?
```

