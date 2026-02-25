# Configuration

`advocate-runner` reads uses three config files:
- `config.yaml`: runner defaults (results root, comparisons root, runtime/mode info, etc.)
- `profiles.yaml`: profiles used for running and comparing
- `metrics_select.yaml`: metrics to compare results cross-test or per-test (see [Metrics](./Docs/Metrics.md))
---

## `config.yaml`
`config.yaml` is read by the CLI to locate defaults such as:
- where results are written (default: `results/`)
- where the `advocate_bin` and `patched_go` runtime lives (runtime information)
- which fuzzing modes are available during testing 

Provide it explicitly like below or omit it to use the default `config.yaml` in the root dir.

```bash
./advocate-runner <test/compare> --config config.yaml ...
```

---

## `profiles.yaml` 
`profiles.yaml` defines individual profiles to run/compare the tests with. Currently there are three basic ones `default`, `quick`, `deep`, but are meant to be extended and individually configured.

A `--profile` encodes:
- the ADVOCATE arguments and scenario selection
- the fuzzing modes to run (if fuzzing testing is done)
- the timeouts set
- whether to keep advoacte artefacts

In the UI menus profiles can be selected by name and in non-interactive mode they have to be passed with `--profile <name>`.

```
analysisProfiles:
  mixed-default:
    scen: "m"
    timeoutRec: 20
    timeoutRep: 20
    keepTrace: true
    stats: true
    time: true

fuzzProfiles:
  default:
    maxRuns: 50
    timeoutFuz: 10
    timeoutRec: 20
    timeoutRep: 20
    keepTrace: true
    stats: true
    time: true
```
---