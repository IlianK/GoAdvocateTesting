# advocate-runner CLI

This repo provides a small CLI (`advocate-runner`) to simplify running **ADVOCATE's** **analysis** and **fuzzing** mode against Go tests and projects. It structures the original `advocateResult` according to a given configuration and indivdiual run profiles, discovers and matches the restructured results of the same label written in a `results/` folder and compares these runs by generating CSVs under `comparisons/` with metrics extracted from advocate's stat files. The metrics to compare can be selected using an additional config (the `metrics_select.yaml`) and the comparison itself can be done cross-test, per-test, and also generating pivot-tables per selected metric.

---
---

## Quick Start

### Build
To get the binary `./advocate-runner` run at repo root:
```
go test ./...
go build -o advocate-runner ./cmd/advocate-runner
```

### [Config.yaml](./Docs/Config.md)
Build the original `advocate`-bin and set the absolute paths of the following attributes in `config.yaml`. Use relative paths for `advocate_bin` and `patched_go` from below if the project directories `ADVOCATE` and `GoAdvocateTesting` exist at the same level:

```
runtime:
  advocate_bin: "../ADVOCATE/advocate/advocate"
  patched_go: "../ADVOCATE/goPatch/bin/go"
  go_runtime: "../../../Tools/go-runtime/bin/go"
```
Additionally set the name of the desired restructured results folder and if necessary restrict the fuzzing modes to run tests with.

---

## Modes
`advocate-runner test ...` starts the test runs on a given path according to the [configuration](./Docs/Config.md) done in `config.yaml` and `profiles.yaml`.

### [Test Mode](./Docs/Test_Mode.md)
```
./advocate-runner test <path> --config config.yaml --profiles profiles.yaml

./advocate-runner test <path> --recursive --label baseline
```
If `--config` and `--profiles` are not provided it will default to the project rooth paths. 

Add an additional label with `--label baseline` to group runs for later comparison. If no label is provided, the `run-` folder will default to using the current timestamp.

Use `--recursive` to search for `_test.go` files in subdirectories of the given path. Omit it to only run tests directly existing in the given path.


---

### [Compare Mode](./Docs/Compare_Mode.md) (Interactive)
```
./advocate-runner compare <datasetDir> --interactive

./advocate-runner compare <datasetDir>/results --interactive 
```
To compare results either directly provide the path to the `results` folder or the path originally given during running the `test` mode. `--interactive` provides a menu to choose further comparison options.


---

### [Compare Mode](./Docs/Compare_Mode.md) (Non-Interactive)
The compare can be done non-interactively, but it is necessary to provide the `--profile`
```
./advocate-runner compare per-test <datasetDir>/results/<TestName> \
  --kind fuzzing --profile <profile> --config config.yaml

./advocate-runner compare cross-test <datasetDir>/results \
  --kind analysis --profile <profile> --config config.yaml

./advocate-runner compare pivot <datasetDir>/results \
  --profile <profile> --metric Bug_Types --config config.yaml
```



---

### Batch Compare (Everything)
```
./advocate-runner compare all <datasetDir|datasetDir/results> 
```

Include `all-labels` to group compare all runs with respective matching labels (Note: label `-latest` is included per default):
```
./advocate-runner compare all <datasetDir|datasetDir/results> --all-labels 
```

---
---

## Docs
- [Config](./Docs/Config.md)
- [Test Mode](./Docs/Test_Mode.md)
- [Compare Mode](./Docs/Compare_Mode.md)
- [Metrics](./Docs/Metrics.md)