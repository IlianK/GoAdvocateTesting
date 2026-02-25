# Bug: Actual Send on Closed Channel

During the execution of the program, a send on a closed channel occurred.
The occurrence of a send on closed leads to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestP01_PossibleSendOnClosed
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/Scenarios/scenarios_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/Scenarios/scenarios_test.go:65
```go
54 ...
55 
56 
57 // ---------------- P01–P03: Possible Bugs ----------------
58 
59 func TestP01_PossibleSendOnClosed(t *testing.T) {
60 	ch := make(chan int)
61 	go func() {
62 		time.Sleep(10 * time.Millisecond)
63 		close(ch)
64 	}()
65 	ch <- 42 // P01 (race with close)           // <-------
66 }
67 
68 func TestP02_PossibleRecvOnClosed(t *testing.T) {
69 	ch := make(chan int)
70 	go func() {
71 		ch <- 1
72 		close(ch)
73 	}()
74 	time.Sleep(10 * time.Millisecond)
75 	_ = <-ch // P02
76 
77 ...
```


###  Channel: Close
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/Scenarios/scenarios_test.go:63
```go
52 ...
53 
54 	close(ch)
55 }
56 
57 // ---------------- P01–P03: Possible Bugs ----------------
58 
59 func TestP01_PossibleSendOnClosed(t *testing.T) {
60 	ch := make(chan int)
61 	go func() {
62 		time.Sleep(10 * time.Millisecond)
63 		close(ch)           // <-------
64 	}()
65 	ch <- 42 // P01 (race with close)
66 }
67 
68 func TestP02_PossibleRecvOnClosed(t *testing.T) {
69 	ch := make(chan int)
70 	go func() {
71 		ch <- 1
72 		close(ch)
73 	}()
74 
75 ...
```


## Replay
**Replaying was not run**.

