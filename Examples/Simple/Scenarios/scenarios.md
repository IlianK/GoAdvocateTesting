# Scenarios

The `Scenarios` directory contains test cases to cover each scenario Advocate is able to cover via analysis. 

## Modes Detection Scope

| **Test Function**                    | **Description**                                | **GFuzz** | **GoPie (incl. +)** | **HB-Based Modes** |
| ------------------------------------ | ---------------------------------------------- | :-------: | :-----------------: | :----------------: |
| `TestA00_UnknownPanic`               | Explicit `panic(...)`                          |     ✓     |          ✓          |          ✓         |
| `TestA01_SendOnClosed`               | Send on a closed channel → runtime panic       |     ✓     |          ✓          |          ✓         |
| `TestA02_ReceiveOnClosed`            | Receive from closed channel returns zero       |           |                     |          ✓         |
| `TestA03_CloseOnClosed`              | Closing a channel twice → runtime panic        |     ✓     |          ✓          |          ✓         |
| `TestA04_CloseOnNil`                 | Closing a nil channel → runtime panic          |     ✓     |          ✓          |          ✓         |
| `TestA05_NegativeWaitGroup`          | Too many `wg.Done()` calls → runtime panic     |     ✓     |          ✓          |          ✓         |
| `TestA06_UnlockUnlocked`             | Unlocking an unlocked mutex → runtime panic    |     ✓     |          ✓          |          ✓         |
| `TestA07_ConcurrentRecv`             | Two concurrent `<-ch` ops, no panic or hang    |           |                     |          ✓         |
| `TestP01_PossibleSendOnClosed`       | Racing `ch <-` vs. `close(ch)` (no panic)      |           |                     |          ✓         |
| `TestP02_PossibleRecvOnClosed`       | Racing receives after close (no panic)         |           |                     |          ✓         |
| `TestP03_PossibleNegativeWaitGroup`  | Two `Done()` in goroutine (no panic)           |           |                     |          ✓         |
| `TestL00_UnknownLeak`                | Goroutine blocks on unclosed channel           |     ✓     |          ✓          |          ✓         |
| `TestL01_UnbufferedLeakWithPartner`  | Send paired later by goroutine → leak at exit  |     ✓     |          ✓          |          ✓         |
| `TestL02_UnbufferedLeakNoPartner`    | Send on unbuffered channel with no receiver    |     ✓     |          ✓          |          ✓         |
| `TestL03_BufferedLeakWithPartner`    | Buffered send consumed too late → leak at exit |     ✓     |          ✓          |          ✓         |
| `TestL04_BufferedLeakNoPartner`      | Send on full buffered channel with no reader   |     ✓     |          ✓          |          ✓         |
| `TestL05_LeakOnNilChan`              | Send on nil channel → blocks forever           |     ✓     |          ✓          |          ✓         |
| `TestL06_LeakOnSelectWithPartner`    | Select waiting on channel, matched later       |     ✓     |          ✓          |          ✓         |
| `TestL07_LeakOnSelectWithoutPartner` | Select on nil channel, fallback via timeout    |     ✓     |          ✓          |          ✓         |
| `TestL08_LeakOnMutex`                | `mu.Lock()` blocks, `Unlock()` delayed         |     ✓     |          ✓          |          ✓         |
| `TestL09_LeakOnWaitGroup`            | `wg.Wait()` never unblocks                     |     ✓     |          ✓          |          ✓         |
| `TestL10_LeakOnCond`                 | `cond.Wait()` without `Signal()`               |     ✓     |          ✓          |          ✓         |


**Comparison**:
- [Bug Types](./results/comparison_pivot_Bug_Types.csv)

- [Total Time](./results/comparison_pivot_Total_Time_s.csv)
