# Compare Mode

---

## Purpose

Compare mode turns the structured `results/` into **CSV summaries** under `<datasetDir>/comparisons/` to make runs comparable at a glance:

- **cross-test**: compare metrics **across many tests** (most useful for analysis runs)
- **per-test**: compare metrics **within one test across fuzz modes** (most useful for fuzzing runs)
- **pivot**: pivot **one chosen metric across tests and fuzz modes** (“matrix view”)

---

## Concepts

### Dataset directory

A dataset directory is the folder that contains `results/` (and later `comparisons/`), e.g.:

```
<datasetDir>/
  results/          // created with advocate-runner test
  comparisons/      // created with advocate-runner compare
```

Most commands accept either:
- `<datasetDir>`
- `<datasetDir>/results`

The CLI will resolve both correctly.


### Kind
- **analysis**: one analysis run per test (no fuzzing mode dimension)
- **fuzzing**: many runs per test across fuzz modes (e.g. Flow, GFuzz, …)


### Label
A label groups runs (e.g. `baseline` or config specific names for identification)
- If you pass no `--label`, comparisons use latest runs per grouping.
- If you pass `--label X`, comparisons use that label.

In output folders, label is reflected as:
- `label-latest` (when no label specified)
- `label-<yourlabel>`

---

## Run
### Interactive Menu 
The **`--interactive`** mode shows menus to pick:
- test
- kind (analysis / fuzzing)
- profile
- label (or “latest”)
- metric (for pivot)


Run with `<datasetDir>` or pass results directly with `<datasetDir>/results`
```
./advocate-runner compare <datasetDir> --interactive --config config.yaml

./advocate-runner compare <datasetDir>/results --interactive --config config.yaml
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
  --profile <profile> \
  --config config.yaml

# cross-test (all tests, one kind/profile)
./advocate-runner compare cross-test <datasetDir>/results \
  --kind analysis \
  --profile <profile> \
  --config config.yaml

# pivot (fuzzing only)
./advocate-runner compare pivot <datasetDir>/results \
  --profile <profile> \
  --metric Bug_Types \
  --config config.yaml
```

Add a label (optional), if omitted, comparisons are generated for latest runs:
```
--label baseline
```

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
./advocate-runner compare all <datasetDir|datasetDir/results> \
  --config config.yaml
```

This generates for every `(kind, profile)`:
- `cross-test` (latest)
- `per-test` for every test (latest)

For fuzzing profiles:
- `pivot` for the core metrics (latest)

Alternatively **include all explicit labels** (not only latest):

```
./advocate-runner compare all <datasetDir|datasetDir/results> \
  --all-labels \
  --config config.yaml
```

This generates the same structure for every discovered label as well as `label-latest`.
