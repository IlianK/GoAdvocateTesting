# WaitGroup 
The `WaitGroup` is used to wait for a collection of goroutines to finish executing. These test cases explore common problems such as double calls to `Done()`, missing `Done()` calls, and nested `WaitGroup` operations.

---

## Double Done Call Without Add
This test demonstrates the scenario where the `Done()` method is called twice without a corresponding `Add()` call to match the counter. This leads to a negative counter in the `WaitGroup`, which can cause unexpected behavior.

The first call to `Done()` decreases the counter, but the second call leads to a negative counter, which is invalid. This could cause the `WaitGroup` to behave incorrectly, such as signaling when it shouldn't.

```go
func TestWGDoubleDone(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()
	wg.Done() // Negative counter
}
```

---

## Missing Done Call After Add
This test shows a scenario where `Add()` is called to increment the `WaitGroup` counter, but the corresponding `Done()` call is forgotten in a goroutine. This results in the `Wait()` method blocking indefinitely.

The main goroutine calls `wg.Wait()`, but since the goroutine fails to call `Done()`, the `WaitGroup` counter never reaches zero, causing the main goroutine to block forever.

```go
func TestWGMissingDone(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// forgot wg.Done()
		time.Sleep(20 * time.Millisecond)
	}()
	wg.Wait() // Leak
}
```

---

## Nested Add and Done
This test demonstrates nested `Add()` and `Done()` calls within goroutines. The counter is adjusted within multiple nested goroutines, and the `WaitGroup` must correctly handle these nested operations.

The first `Add()` increments the counter by 2, and two goroutines each call `Done()`, decrementing the counter by 1 each. The `WaitGroup` should wait until both `Done()` calls are made, ensuring synchronization across the nested goroutines.

```go
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
```

**Comparison**:
- [Bug Types](./results/comparison_pivot_Bug_Types.csv)

- [Total Time](./results/comparison_pivot_Total_Time_s.csv)
