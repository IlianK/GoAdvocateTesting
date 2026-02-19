# Metrics

## Core Metrics
```
(Mode)
Unique_Bugs
Bug_Types
Total_Bugs
Panics
Leaks
Confirmed_Replays
Total_Runs
Total_Time_s
Rec_s
Ana_s
Rep_s
Replays_Written
Replays_Successful
```

---

## Meaning and Source

- **Unique_Bugs**: number of unique bug types detected (best-effort; from stats, with fallback to `results_machine.log`).

- **Bug_Types**: semicolon-separated bug codes detected (e.g. `A06`;`L00`;`P06`), extracted from `output/results_machine.log`.

- **Total_Bugs**: number of bug report files (`bugs_*.md`) produced for the run (a “how many reports” indicator).

- **Panics**: total number of panics observed in analysis (`NrPanicsTotal` from `statsAnalysis_*`).

- **Leaks**: total number of leaks observed in analysis (`NrLeaksTotal` from `statsAnalysis_*`).

- **Confirmed_Replays**: panics/leaks confirmed or resolved via replay (`NrPanicsVerifiedViaReplayTotal + NrLeaksResolvedViaReplayTotal`).

- **Total_Runs**: number of fuzzing mutations/executions (`NrMut` from `statsFuzz_*`; typically 0 for analysis-only runs).

- **Total_Time_s**: total wall-clock time for the run from `times_total_*.csv`.

- **Rec_s**: time spent recording (from `rec_s` or `times_detail_*.csv`).

- **Ana_s**: time spent in analysis (from `ana_s` or `times_detail_*.csv`).

- **Rep_s**: time spent in replay (from `rep_s` or `times_detail_*.csv`).

- **Replays_Written**: total replay files written across bug types (summed from `statsAll_*` replay-written fields).

- **Replays_Successful**: total successful replays across bug types (summed from `statsAll_*` replay-success fields).
