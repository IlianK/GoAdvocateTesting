# advocate-runner CLI

This repo provides a small CLI (`advocate-runner`) to simplify running **ADVOCATE** the **analysis** / **fuzzing** modes against Go tests and projects. It structures the original `advocateResult` according to a configuration and indivdiual run profiles, discovers and matches results written to a `results/` folder of the same label and compares the runs by generating CSVs under `comparisons/` with extracted metrics. 

---
---

## Quick Start

### Build
To get the binary `./advocate-runner` run at repo root:
```
go test ./...
go build -o advocate-runner ./cmd/advocate-runner
```

---

### Test (Interactive)
```
./advocate-runner test <path> --config config.yaml --profiles profiles.yaml
./advocate-runner test <path> --recursive --config config.yaml --profiles profiles.yaml
```

---

### Compare (Interactive)
```
./advocate-runner compare <datasetDir> --interactive --config config.yaml
./advocate-runner compare <datasetDir>/results --interactive --config config.yaml
```

---

### Compare (Non-Interactive)
```
./advocate-runner compare per-test <datasetDir>/results/<TestName> \
  --kind fuzzing --profile <profile> --config config.yaml

./advocate-runner compare cross-test <datasetDir>/results \
  --kind analysis --profile <profile> --config config.yaml

./advocate-runner compare pivot <datasetDir>/results \
  --profile <profile> --metric Bug_Types --config config.yaml
```

Add a label (optional):
```
--label baseline
```

---

### Batch Compare (Everything)
```
./advocate-runner compare all <datasetDir|datasetDir/results> \
  --config config.yaml
```

Include all explicit labels:
```
./advocate-runner compare all <datasetDir|datasetDir/results> \
  --all-labels \
  --config config.yaml
```

---
---

## Docs
- [Config](./Docs/Config.md)
- [Test Mode](./Docs/Test_Mode.md)
- [Compare Mode](./Docs/Compare_Mode.md)
- [Metrics](./Docs/Metrics.md)