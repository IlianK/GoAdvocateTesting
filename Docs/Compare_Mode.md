# Compare Mode

---

## Purpose

Compare mode turns the structured `results/` into **CSV summaries** under `<datasetDir>/comparisons/` to make runs comparable at glance:

- **cross-test**: compare metrics **across many tests** (most useful for analysis runs)
- **per-test**: compare metrics **within one test across fuzz modes** (most useful for fuzzing runs)
- **pivot**: one chosen metric **across tests and fuzz modes** (“matrix view”)

---

## Concepts

### Dataset directory

A dataset directory is the folder that contains `results/` (and later `comparisons/`), e.g.:

```
<datasetDir>/
  results/          // created with advocate-runner test
  comparisons/      // created with advocate-runner compare
```


### Kind
- **analysis**: one analysis run per test
- **fuzzing**: many runs per test across fuzz modes (listed in `config.yaml` e.g. Flow, GFuzz, …)


### Label
A label groups runs (e.g. `baseline` or config specific names for identification)
- If no `--label` is passed, comparisons use latest runs for grouping.
- If `--label X` is passed, comparisons use that label.

In output folders, label is reflected as:
- `run-latest` (when no label specified)
- `run-<label>`

---

## Run
### Interactive Menu 
The **`--interactive`** mode shows menus to select:
- `kind`: `analysis` / `fuzzing` results to compare
- `compare-action`: `cross-test`/ `per-test`
- `profile`
- `label` (or “latest”)
- (`metric` only for pivot)


Run providing the dataset path or pass results directly. The CLI will correctly resolve paths as either:
- `<datasetDir>`
- `<datasetDir>/results`

```
./advocate-runner compare <datasetDir|datasetDir/results>  --interactive 
```

### Non-Interactive
**`Non-interactive`** mode uses explicit subcommands:
- `per-test`: To compare all metrics of all fuzz modes for one test
- `cross-test`: To compare all metrics across all tests
- `pivot`: To pivot one metric across all tests for all fuzz modes

**Examples**:
```
# per-test (all modes within one test)
./advocate-runner compare per-test <datasetDir>/results/<TestName> \
  --kind fuzzing \
  --profile <profile> 

# cross-test (all tests, one kind/profile)
./advocate-runner compare cross-test <datasetDir>/results \
  --kind analysis \
  --profile <profile> 

# pivot (fuzzing only)
./advocate-runner compare pivot <datasetDir>/results \
  --profile <profile> \
  --metric Bug_Types 
```

Add a label (optional), if omitted, comparisons are generated for latest runs.

---

## Comparison Folder Structure (`comparisons/`)
### Cross-test outputs
```
comparisons/
  cross-test/
    kind-analysis/
      profile-<profile>/
        label-latest/
          cross_test.csv
          inputs.json
    kind-fuzzing/
      profile-<profile>/
        label-latest/
          cross_test.csv
          inputs.json
```

### Per-test outputs
```
comparisons/
  per-test/
    <TestName>/
      kind-analysis/
        profile-<profile>/
          label-latest/
            compare.csv
            inputs.json
      kind-fuzzing/
        profile-<profile>/
          label-latest/
            compare.csv
            inputs.json
```

### Pivot outputs (fuzzing only)
```
comparisons/
  pivot/
    kind-fuzzing/
      profile-<profile>/
        label-latest/
          pivot_<Metric>.csv
          inputs.json     
```

---

## Batch Compare: Generate All Comparisons
Batch compare is the “one command” way to generate the **full suite** of comparisons.

Run:
```
./advocate-runner compare all <datasetDir|datasetDir/results> 
```

This generates in the provided path for every `kind` and every `profile`:
- `cross-test` results over all tests
- `per-test` results for every test 
- `pivot` results for all metrics listed in `metrics_select.yaml`

Alternatively **include all explicit labels** (not only latest):

```
./advocate-runner compare all <datasetDir|datasetDir/results> \
  --all-labels \
```

This generates the same structure for every discovered label as well as `label-latest`.
