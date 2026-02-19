# Configuration

`advocate-runner` reads two config files:
- `config.yaml`: runner defaults (results root, comparisons root, runtime/mode info, etc.)
- `profiles.yaml`: profiles used for running and comparing

---

## `config.yaml`
`config.yaml` is read by the CLI to locate defaults such as:
- where results are written (default: `results/`)
- where comparisons are written (typically `<datasetDir>/comparisons/`)
- any other repo-specific defaults like available fuzzing modes and runtime information

Use it explicitly:

```bash
./advocate-runner --config config.yaml <command> ...
```

or per subcommand (as implemented in this project):
```
./advocate-runner compare --config config.yaml ...
./advocate-runner compare all --config config.yaml ...
```

---

## `profiles.yaml` 
`profiles.yaml` defines profiles to run/compare the tests with. 

Currently there are three basic ones (extendable): `default`, `quick`, `deep`.

A “profile” typically encodes:
- ADVOCATE arguments / scenario selection
- fuzzing modes to run (if fuzzing)
- timeouts / flags / other run variants

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