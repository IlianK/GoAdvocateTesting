
# Concurrency Analysis Tools

## GFuzz
[GFuzz](https://github.com/system-pclub/GFuzz), is a concurrency analysis tool focused on detecting channel-related bugs in Go programs by systematically mutating select case executions.
According to the [GFuzz paper](https://songlh.github.io/paper/gfuzz.pdf), GFuzz found 184 previously unknown bugs in real-world Go projects 
<a id="cite2"></a>[<a href="#ref2">2</a>]. 

### Mutation strategy 
GFuzz assigns unique identifiers to each select and its cases. In each test run, it enforces a preferred case for every select. After execution, it checks whether new behaviors or operations were explored. If so, it mutates the preferred cases to guide the next run. This targeted message reordering increases the likelihood of uncovering concurrency bugs that only occur under specific interleavings <a id="cite1"></a>[<a href="#ref1">1</a>].

### Run Time Statistics
The mutation strategy is driven by runtime statistics such as the number of send/receive operations, created or closed channels, and buffer states. These metrics are used to calculate mutation scores.

### Order Enforcement
GFuzz modifies the program code to enforce execution order. It instruments each select with a wrapper that initially tries to execute only the preferred case. If the case is not selectable (due to a missing sender/receiver), it falls back to the original select with a timeout, to ensure progress without getting stuck.

### Bug detection
GFuzz focuses on identifying blocking goroutines. If a goroutine is stuck in a channel operation with no possible partner it flags it as a blocking bug.

---
---

## GoPie
[GoPie](https://github.com/CGCL-codes/GoPie) extends GFuzz’s ideas beyond select statements. It targets a broader class of concurrency bugs by directly influencing the interleaving of goroutine scheduling across synchronization primitives like channels and mutexes.
The theory is explained in detail in the [GoPie paper](https://chao-peng.github.io/publication/ase23/ase23.pdf).

### Fragmentation
To manage the complexity and avoid exploring all possible interleavings, GoPie breaks the execution trace into smaller fragments called scheduling chains (SCs). Those are sequences of operations from different goroutines, and only one chain is mutated at a time. So GoPie analysis focus is only on key interaction points between goroutines <a id="cite3"></a>[<a href="#ref3">3</a>].

### Relationships
GoPie identifies potential mutation points by analyzing two types of relationships between operations. These relationships reveal which operations are meaningfully connected and suitable for mutation <a id="cite4"></a>[<a href="#ref4">4</a>].
- Rel1: operations that occur one after another within the same goroutine
- Rel2: operations on the same synchronization primitive (channel or mutex) across different goroutines.

### Mutation Strageties
Based on these relationships, GoPie applies different mutation strategies to explore new (potentially bug-revealing) interleavings <a id="cite3"></a>[<a href="#ref3">3</a>]. . 

- Abridge: shortens the scheduling chain by trimming head or tail operations.
- Flip: reverses operation order in a chain segment.
- Substitute: replaces operations with related ones from Rel1.
- Augment: adds related operations from Rel2 to the chain.

(Note: While the Flip strategy is described in the GoPie paper, it is not implemented in the current version of the tool.)

### Feedback Collection
GoPie selectively chooses which test runs to mutate further, based on feedback from previous executions. A run is considered successful if it completes without exceeding time limits or triggering timeouts on synchronization operations. If a run fails due to timeouts or if runs are too long (over 7 minutes), it may be excluded from future mutations. This mechanism helps to focus resources on promising interleavings and avoid wasting time on ineffective paths <a id="cite4"></a>[<a href="#ref4">4</a>].

### Order Reinforcement
GoPie also enforces execution order based on the mutated schedule. This means that each operation runs only when it is its turn in the planned schedule. If it’s not the next expected operation, it simply waits. 

So by combining selective mutation with schedule-based execution control, GoPie can find concurrency bugs without relying on random scheduling or trying every possible combination.

--- 
---

## Advocate
[ADVOCATE](https://github.com/ErikKassubek/ADVOCATE) builds on GFuzz and GoPie, operating by recording execution traces of programs and analyzing them to identify potential concurrency issues, also using happens-before relations and vector clocks. Advocate supports several modes  <a id="cite5"></a>[<a href="#ref5">5</a>]:
1. **Record**: Captures the execution trace of a program or test.
2. **Replay**: Re-executes a program following a recorded trace to reproduce specific behaviors.
3. **Analysis**: Analyzes recorded traces to detect potential concurrency bugs like channel misuse, leaks and deadlocks.
4. **Fuzzing**: Includes and extends GFuzz (select-based) and GoPie (order-based) strategies.


### Runtime
Advocate uses a custom-modified Go runtime to implement recording, replay, and fuzzing. Instrumentation in this context refers to the insertion of extra logic to monitor or control execution. Traditional tools like GFuzz and GoPie instrument source code directly. For example, the original GFuzz directly instruments each select in the source code to enforce case selection <a id="cite7"></a>[<a href="#ref7">7</a>], and GoPie modifies the source to enable recording and replay <a id="cite9"></a>[<a href="#ref9">9</a>]. Advocate instruments the Go runtime itself, embedding logic into low-level concurrency primitives like select, channel operations, mutexes, and atomics <a id="cite10"></a>[<a href="#ref10">10</a>] <a id="cite11"></a>[<a href="#ref11">11</a>].

This runtime-level instrumentation allows Advocate to:
- Log execution traces during normal runs
- Enforce specific execution orders during replay or fuzzing
- Mutate schedules without modifying application logic

Although a small preamble and import are temporarily added to the program’s main or test function to initiate tracing, these are automatically removed after fuzzing. This minimizes developer overhead and ensures the original code remains untouched <a id="cite11"></a>[<a href="#ref11">11</a>].



### Fuzzing Modes
Advocate's Fuzzing supports several modes, each for exploring different aspects of concurrency interleavings and bug types <a id="cite6"></a>[<a href="#ref6">6</a>]:

#### GFuzz
Reimplements the original GFuzz by mutating select statements to prefer specific cases. Unlike the original, which directly modifies source code, Advocate enforces case preference through its runtime. It uses recorded traces to guide mutations and avoids unnecessary instrumentation by relying on internal scheduling control as seen in <a id="cite7"></a>[<a href="#ref7">7</a>]

#### GFuzzHB
Extends GFuzz by integrating happens-before (HB) analysis. This allows Advocate to detect possible but unexecuted bugs, reducing the number of required runs. It also enhances scoring by incorporating select-case partner availability and biases mutation toward useful interleavings.
<a id="cite7"></a>[<a href="#ref7">7</a>]

#### Flow
The Flow mode uses trace-based HB relations alone (without mutations) to identify concurrency bugs. This means it relies entirely on static trace analysis instead of runtime execution of mutations, which offers a low-overhead bug detection pass.
<a id="cite8"></a>[<a href="#ref8">8</a>]

#### GFuzzHBFlow
This mode combines GFuzzHB's mutation strategy with Flow’s static HB-based analysis. Mutated schedules are executed once, and their traces are analyzed to detect both executed and theoretical bugs.
<a id="cite8"></a>[<a href="#ref8">8</a>]

#### GoPie
Reimplements GoPie’s order-based fuzzing. It mutates the execution order of operations within a scheduling chain derived from the trace. And again, unlike the original, which requires source code changes, Advocate integrates this logic into the runtime and improves replay accuracy.
<a id="cite9"></a>[<a href="#ref9">9</a>]

#### GoPie+
Improves upon GoPie by multiple features:
1. Allowing mutation of all operations, not just channels and mutexes, 
2. Introducing new chains from each mutation
3. Implementing the flip strategy missing in the original
4. Mutating even incomplete (leaking) operations
5. Enforcing more consistent replay.
<a id="cite9"></a>[<a href="#ref9">9</a>]

#### GoPieHB
Extends GoPie+ with HB analysis. This enables detection of theoretically possible bugs without requiring those interleavings to be executed, which improves both precision and coverage in concurrency testing.
<a id="cite9"></a>[<a href="#ref9">9</a>]


| **Note**: The fuzzing modes GFuzz, GoPie, and GoPie+ detect only bugs that result in observable effects, such as panics or goroutine leaks. They cannot detect silent or theoretical concurrency issues like scenarios A02, A07, or P01–P03, which require reasoning over happens-before relations. To detect these the analysis-based modes like Flow, GFuzzHB, GFuzzHBFlow, and GoPieHB should be used. which has HB-based detection logic.

### Error Classification System
| Code  | Type             | Meaning                                             |
| ----- | ---------------- | --------------------------------------------------- |
| `Axx` | Actual bug/panic | Happened during execution                           |
| `Rxx` | Runtime fatal    | Unexpected or forced termination (e.g., timeout)    |
| `Pxx` | Potential bug    | Detected by HB analysis, didn’t occur yet           |
| `Lxx` | Leak             | Operation blocked or left incomplete at program end |



### Result artefacts
Each test produces results for a selected fuzzing engine (e.g., GFuzz, GoPie, GoPie+) including log files (output.log, results_machine.log, results_readable.log) and a bug summary (bug_*.md). Also csv-based statistics reports are created when flag `-stats` is set (statsAll, statsAnalysis, statsFuzzing) <a id="cite12"></a>[<a href="#ref12">12</a>].

---


## References
[↩](#cite1) <a id="ref1">[1]</a> Erik Kassubek, GFuzz Summary, AdvocateGo Project, https://github.com/ErikKassubek/ADVOCATE/blob/main/doc_proj/relatedWorks/PaperAndTools/Fuzzing/GFuzz.md

[↩](#cite2) <a id="ref2">[2]</a> Z. Liu, S. Xia, Y. Liang, L. Song, and H. Hu. Who goes first? Detecting Go concurrency bugs via message reordering. ASPLOS 2022, https://songlh.github.io/paper/gfuzz.pdf

[↩](#cite3)  <a id="ref3">[3]</a> Z. Jiang, M. Wen, Y. Yang, C. Peng, P. Yang, and H. Jin. Effective Concurrency Testing for Go via Directional Primitive-Constrained Interleaving Exploration. ASE 2023, https://chao-peng.github.io/publication/ase23/ase23.pdf

[↩](#cite4)  <a id="ref4">[4]</a> Erik Kassubek, GoPie Summary, AdvocateGo Project, https://github.com/ErikKassubek/ADVOCATE/blob/main/doc_proj/relatedWorks/PaperAndTools/Fuzzing/GoPie.md

[↩](#cite5)  <a id="ref5">[5]</a> Erik Kassubek, GoPie Usage, AdvocateGo Project, https://github.com/ErikKassubek/ADVOCATE/blob/main/doc_proj/usage.md

[↩](#cite6)  <a id="ref6">[6]</a> Erik Kassubek, GoPie Fuzzing, AdvocateGo Project, https://github.com/ErikKassubek/ADVOCATE/tree/main/doc/fuzzing

[↩](#cite7)  <a id="ref7">[7]</a> Erik Kassubek, Advocate GFuzz Mode, AdvocateGo Project, https://github.com/ErikKassubek/ADVOCATE/blob/main/doc/fuzzing/GFuzz.md

[↩](#cite8)  <a id="ref8">[8]</a> Erik Kassubek, Advocate Flow Mode, AdvocateGo Project, https://github.com/ErikKassubek/ADVOCATE/blob/main/doc/fuzzing/GoPie.md

[↩](#cite9)  <a id="ref9">[9]</a> Erik Kassubek, Advocate GoPie Mode, AdvocateGo Project, https://github.com/ErikKassubek/ADVOCATE/blob/main/doc/fuzzing/GoPie.md

[↩](#cite10)  <a id="ref10">[10]</a> Erik Kassubek, Advocate Runtime, AdvocateGo Project, https://github.com/ErikKassubek/ADVOCATE/blob/main/doc_proj/runtime.md

[↩](#cite11)  <a id="ref11">[11]</a> Erik Kassubek, Advocate Recording, AdvocateGo Project, https://github.com/ErikKassubek/ADVOCATE/blob/main/doc_proj/recording.md

[↩](#cite12)  <a id="ref12">[12]</a> Erik Kassubek, Advocate Statistics, AdvocateGo Project, https://github.com/ErikKassubek/ADVOCATE/blob/main/doc/stats/stats.md