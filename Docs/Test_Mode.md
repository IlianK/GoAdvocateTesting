# Test Mode

---

## Purpose

Test mode runs ADVOCATE based on the config in `config.yaml` and `profiles.yaml` and organizes the output of the runs into a consistent, queryable structure under `results/`.

Instead of working with a single `advocateResult` folder, runs are stored:

- grouped by **TestName**
- separated by **kind** (`analysis` / `fuzzing`)
- separated by **profile** (from `profiles.yaml`)
- optionally grouped by a **label** (e.g. `baseline`), if omitted defaulting to a timestamp
- uniquely identified by **runId**

This makes it easy to keep many runs, rerun experiments, and later compare them automatically.
The basic structure is: `results/<TestName>/<kind>/<profile>/run-<label>`

---

## Run

```bash
./advocate-runner test <path> --config config.yaml --profiles profiles.yaml
```

If `<path>` contains multiple `*_test.go` files, the runner will offer a selection in the menu, if `Run on ONE test` is chosen, else all tests found in the given directoriy will be run with the selected run `kind`.

If `--config` and `--profile` are omitted, they default to using the files provided in the root dir.

---

## Recursive discovery
If tests are in subdirectories, `--recursive` should be used to find them (used for testing projects):

```
./advocate-runner test <path> --recursive 
```

---

## Interactive Menu Flow
The interactive menu should look like the following:
1. Select run kind
- Analysis
- Fuzzing

2. If Analysis
- Select analysis profile
- Select scope:
    - Run on ALL tests
    - Run on ONE test

3. If Fuzzing
- Select fuzz profile
- Select a fuzz plan:
    - Run ALL modes on ALL tests
    - Run ALL modes on ONE test
    - Run ONE mode on ALL tests
    - Run ONE mode on ONE test

If `ONE test` is chosen the menu will offer a selection. Likewise, when `ONE mode` is chosen during the fuzz plan selection, the menu will offer a selection of the available fuzzing modes.

---

## Flags
- `--config config.yaml`: path to config (defaulting to `./config.yaml`)
- `--profiles profiles.yaml`: path to profiles (defaulting to `./profiles.yaml`)
- `--recursive`: discover tests recursively (directories containing `*_test.go`)
- `--label <name>`: group results e.g. `baseline`, `.../<profile>/run-baseline` (defaulting to using timestamp)
- `--keep-raw`: do not delete the raw `advocateResult` folder after moving results (debug)

---

## Results Folder Structure ('results/`)
A typical run layout (one run folder) looks like:

```
results/<TestName>/<kind>/<profile>/<runId>/
  meta.json
  output/
    output.log
    results_machine.log
    results_readable.log
  stats/
    statsAll_...
    statsAnalysis_...
    statsFuzz_...        
    statsTrace_...
  traces/
    trace_*.log
    trace_info.log
  times/                 
    times_total_*.csv
    times_detail_*.csv
  bugs/
    bugs_1.md           
```

**Notes**:
- `meta.json` is created by this project and used for discovery/compare.
- `results_machine.log` is used for `Bug_Types` extraction.
- `statsAll_*` / `statsAnalysis_*` / `statsFuzz_*` are used for numeric metrics.
