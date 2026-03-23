package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"GoAdvocateTesting/internal/app"
	"GoAdvocateTesting/internal/compare"
	"GoAdvocateTesting/internal/discovery"
	"GoAdvocateTesting/internal/metrics"
	"GoAdvocateTesting/internal/ui"
)

func cmdCompare(args []string) { // no normalization here, subcommands have their own flags
	if len(args) == 0 {
		printCompareUsage()
		os.Exit(2)
	}

	// Find first non-flag token => candidate subcommand or dataset path.
	subcmdIdx := -1
	for i, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		subcmdIdx = i
		break
	}

	// Dispatch known subcommands even if flags are placed before them with normalize
	if subcmdIdx >= 0 {
		switch args[subcmdIdx] {
		case "all":
			rest := append(args[:subcmdIdx], args[subcmdIdx+1:]...)
			cmdCompareAll(normalizeArgsForFlags(rest))
			return
		case "per-test":
			rest := append(args[:subcmdIdx], args[subcmdIdx+1:]...)
			cmdComparePerTest(normalizeArgsForFlags(rest))
			return
		case "cross-test":
			rest := append(args[:subcmdIdx], args[subcmdIdx+1:]...)
			cmdCompareCrossTest(normalizeArgsForFlags(rest))
			return
		case "pivot":
			rest := append(args[:subcmdIdx], args[subcmdIdx+1:]...)
			cmdComparePivot(normalizeArgsForFlags(rest))
			return
		}
	}

	// Otherwise treat as dataset compare entrypoint (interactive only).
	cmdCompareInteractive(normalizeArgsForFlags(args))
}

func printCompareUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  advocate-runner compare <datasetDir|datasetDir/results> --interactive [--config config.yaml] [--metrics-select metrics_select.yaml] [--metrics-def internal/metrics/metrics_def.yaml]")
	fmt.Fprintln(os.Stderr, "  advocate-runner compare all <datasetDir|datasetDir/results> [--all-labels] [--config config.yaml] [--metrics-select metrics_select.yaml] [--metrics-def internal/metrics/metrics_def.yaml]")
	fmt.Fprintln(os.Stderr, "  advocate-runner compare per-test <datasetDir>/results/<TestName> --kind fuzzing --profile default [--label baseline]")
	fmt.Fprintln(os.Stderr, "  advocate-runner compare cross-test <datasetDir|datasetDir/results> --kind analysis --profile mixed-default [--label baseline] [--mode <fuzzMode>]")
	fmt.Fprintln(os.Stderr, "  advocate-runner compare pivot <datasetDir|datasetDir/results> --profile default [--label baseline] --metric <MetricKey>")
}

func cmdCompareInteractive(args []string) {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "Path to config.yaml")
	metricsSelectPath := fs.String("metrics-select", "", "Path to metrics_select.yaml (default: alongside config.yaml)")
	metricsDefPath := fs.String("metrics-def", "internal/metrics/metrics_def.yaml", "Path to metrics_def.yaml")
	interactive := fs.Bool("interactive", false, "Interactive compare menu (required to open menu)")
	resultsRootOverride := fs.String("results-root", "", "Override results root folder (default from config)")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		printCompareUsage()
		os.Exit(2)
	}
	if !*interactive {
		fmt.Fprintln(os.Stderr, "Non-interactive compare requires a subcommand: all | per-test | cross-test | pivot")
		fmt.Fprintln(os.Stderr, "Tip: add --interactive to open the menu.")
		os.Exit(2)
	}

	cfg, err := app.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Default metrics_select.yaml path: same directory as config.yaml
	msel := *metricsSelectPath
	if strings.TrimSpace(msel) == "" {
		msel = filepath.Join(filepath.Dir(*configPath), "metrics_select.yaml")
	}
	if err := metrics.Configure(*metricsDefPath, msel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load metrics YAMLs: %v\n", err)
		os.Exit(1)
	}

	datasetDir, resultsRoot, err := resolveDatasetAndResultsRoot(rest[0], cfg, *resultsRootOverride)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve dataset/results: %v\n", err)
		os.Exit(1)
	}

	runs, err := discovery.DiscoverRuns(datasetDir, resultsRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to discover runs: %v\n", err)
		os.Exit(1)
	}
	if len(runs) == 0 {
		fmt.Fprintf(os.Stderr, "No runs found under %s/%s\n", datasetDir, resultsRoot)
		os.Exit(1)
	}

	index := discovery.NewRunIndex(runs)

	// Metrics for pivot menu: ordered keys from metrics_select.yaml/metrics_def.yaml
	metricKeys, err := metrics.ActiveCSVHeader(false, false)
	if err != nil || len(metricKeys) == 0 {
		// fallback to sampling extracted keys
		metricKeys = unionMetricKeysFromSample(runs, 200)
	}

	sel := ui.CompareMenu(ui.CompareMenuParams{
		TestDir:    datasetDir,
		Index:      index,
		MetricKeys: metricKeys,
	})

	switch sel.Action {
	case ui.CompareActionPerTest:
		outDir, err := compare.ComparePerTest(compare.PerTestParams{
			TestDir:     datasetDir,
			ResultsRoot: resultsRoot,
			TestName:    sel.TestName,
			Kind:        sel.Kind,
			Profile:     sel.Profile,
			Label:       sel.Label,
		})
		if err != nil {
			ui.PrintError(err.Error())
			os.Exit(1)
		}
		ui.PrintOK("Wrote: " + outDir)

	case ui.CompareActionCrossTest:
		outDir, err := compare.CompareCrossTest(compare.CrossTestParams{
			TestDir:     datasetDir,
			ResultsRoot: resultsRoot,
			Kind:        sel.Kind,
			Profile:     sel.Profile,
			Label:       sel.Label,
			Mode:        sel.Mode,
		})
		if err != nil {
			ui.PrintError(err.Error())
			os.Exit(1)
		}
		ui.PrintOK("Wrote: " + outDir)

	case ui.CompareActionPivot:
		outDir, err := compare.Pivot(compare.PivotParams{
			TestDir:     datasetDir,
			ResultsRoot: resultsRoot,
			Profile:     sel.Profile,
			Label:       sel.Label,
			Metric:      sel.Metric,
		})
		if err != nil {
			ui.PrintError(err.Error())
			os.Exit(1)
		}
		ui.PrintOK("Wrote: " + outDir)

	default:
		ui.PrintError("Unknown compare action")
		os.Exit(1)
	}
}

func cmdCompareAll(args []string) {
	// Batch generation: no interactive flag
	fs := flag.NewFlagSet("compare all", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "Path to config.yaml")
	metricsSelectPath := fs.String("metrics-select", "", "Path to metrics_select.yaml (default: alongside config.yaml)")
	metricsDefPath := fs.String("metrics-def", "internal/metrics/metrics_def.yaml", "Path to metrics_def.yaml")
	resultsRootOverride := fs.String("results-root", "", "Override results root folder (default from config)")
	allLabels := fs.Bool("all-labels", false, "Also generate comparisons for each explicit label (in addition to latest)")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: advocate-runner compare all <datasetDir|datasetDir/results> [--all-labels] [--config config.yaml]")
		os.Exit(2)
	}

	cfg, err := app.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Configure metrics yaml
	msel := *metricsSelectPath
	if strings.TrimSpace(msel) == "" {
		msel = filepath.Join(filepath.Dir(*configPath), "metrics_select.yaml")
	}
	if err := metrics.Configure(*metricsDefPath, msel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load metrics YAMLs: %v\n", err)
		os.Exit(1)
	}

	datasetDir, resultsRoot, err := resolveDatasetAndResultsRoot(rest[0], cfg, *resultsRootOverride)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve dataset/results: %v\n", err)
		os.Exit(1)
	}

	runs, err := discovery.DiscoverRuns(datasetDir, resultsRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to discover runs: %v\n", err)
		os.Exit(1)
	}
	if len(runs) == 0 {
		fmt.Fprintf(os.Stderr, "No runs found under %s/%s\n", datasetDir, resultsRoot)
		os.Exit(1)
	}

	idx := discovery.NewRunIndex(runs)

	for _, kind := range idx.Kinds() {
		profiles := idx.Profiles(kind)
		sort.Strings(profiles)

		for _, profile := range profiles {
			labels := []string{""} // "" means latest
			if *allLabels {
				labels = append(labels, idx.Labels(kind, profile)...)
			}

			for _, label := range labels {
				// Cross-test
				if _, err := compare.CompareCrossTest(compare.CrossTestParams{
					TestDir:     datasetDir,
					ResultsRoot: resultsRoot,
					Kind:        kind,
					Profile:     profile,
					Label:       label,
					Mode:        "",
				}); err != nil {
					ui.PrintError(fmt.Sprintf("[cross-test] kind=%s profile=%s label=%s: %v", kind, profile, labelOrLatest(label), err))
				} else {
					ui.PrintOK(fmt.Sprintf("[cross-test] kind=%s profile=%s label=%s", kind, profile, labelOrLatest(label)))
				}

				// Per-test for all tests
				tests := idx.Tests(kind, profile, label)
				sort.Strings(tests)
				for _, testName := range tests {
					if _, err := compare.ComparePerTest(compare.PerTestParams{
						TestDir:     datasetDir,
						ResultsRoot: resultsRoot,
						TestName:    testName,
						Kind:        kind,
						Profile:     profile,
						Label:       label,
					}); err != nil {
						ui.PrintError(fmt.Sprintf("[per-test] %s kind=%s profile=%s label=%s: %v", testName, kind, profile, labelOrLatest(label), err))
					}
				}
				ui.PrintOK(fmt.Sprintf("[per-test] done kind=%s profile=%s label=%s (%d tests)", kind, profile, labelOrLatest(label), len(tests)))
			}
		}
	}

	ui.PrintOK("Batch compare finished.")
}

func labelOrLatest(label string) string {
	if strings.TrimSpace(label) == "" {
		return "latest"
	}
	return label
}

func cmdComparePerTest(args []string) {
	args = normalizeArgsForFlags(args)

	fs := flag.NewFlagSet("compare per-test", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "Path to config.yaml")
	metricsSelectPath := fs.String("metrics-select", "", "Path to metrics_select.yaml (default: alongside config.yaml)")
	metricsDefPath := fs.String("metrics-def", "internal/metrics/metrics_def.yaml", "Path to metrics_def.yaml")
	kind := fs.String("kind", "fuzzing", "analysis|fuzzing")
	profile := fs.String("profile", "", "Profile name (required)")
	label := fs.String("label", "", "Run label (optional; empty means latest)")
	resultsRootOverride := fs.String("results-root", "", "Override results root folder (default from config)")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: advocate-runner compare per-test <datasetDir>/results/<TestName> --kind fuzzing --profile default [--label baseline]")
		os.Exit(2)
	}
	if *profile == "" {
		fmt.Fprintln(os.Stderr, "--profile is required")
		os.Exit(2)
	}

	cfg, err := app.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	msel := *metricsSelectPath
	if strings.TrimSpace(msel) == "" {
		msel = filepath.Join(filepath.Dir(*configPath), "metrics_select.yaml")
	}
	if err := metrics.Configure(*metricsDefPath, msel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load metrics YAMLs: %v\n", err)
		os.Exit(1)
	}

	datasetDir, resultsRoot, testName, err := resolvePerTestTarget(rest[0], cfg, *resultsRootOverride)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bad target: %v\n", err)
		os.Exit(2)
	}

	outDir, err := compare.ComparePerTest(compare.PerTestParams{
		TestDir:     datasetDir,
		ResultsRoot: resultsRoot,
		TestName:    testName,
		Kind:        *kind,
		Profile:     *profile,
		Label:       *label,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compare failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(outDir)
}

func cmdCompareCrossTest(args []string) {
	args = normalizeArgsForFlags(args)

	fs := flag.NewFlagSet("compare cross-test", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "Path to config.yaml")
	metricsSelectPath := fs.String("metrics-select", "", "Path to metrics_select.yaml (default: alongside config.yaml)")
	metricsDefPath := fs.String("metrics-def", "internal/metrics/metrics_def.yaml", "Path to metrics_def.yaml")
	kind := fs.String("kind", "analysis", "analysis|fuzzing")
	profile := fs.String("profile", "", "Profile name (required)")
	label := fs.String("label", "", "Run label (optional; empty means latest)")
	mode := fs.String("mode", "", "For fuzzing only: restrict to one mode (optional)")
	resultsRootOverride := fs.String("results-root", "", "Override results root folder (default from config)")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: advocate-runner compare cross-test <datasetDir|datasetDir/results> --kind analysis --profile mixed-default [--label baseline]")
		os.Exit(2)
	}
	if *profile == "" {
		fmt.Fprintln(os.Stderr, "--profile is required")
		os.Exit(2)
	}

	cfg, err := app.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	msel := *metricsSelectPath
	if strings.TrimSpace(msel) == "" {
		msel = filepath.Join(filepath.Dir(*configPath), "metrics_select.yaml")
	}
	if err := metrics.Configure(*metricsDefPath, msel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load metrics YAMLs: %v\n", err)
		os.Exit(1)
	}

	datasetDir, resultsRoot, err := resolveDatasetAndResultsRoot(rest[0], cfg, *resultsRootOverride)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bad target: %v\n", err)
		os.Exit(2)
	}

	outDir, err := compare.CompareCrossTest(compare.CrossTestParams{
		TestDir:     datasetDir,
		ResultsRoot: resultsRoot,
		Kind:        *kind,
		Profile:     *profile,
		Label:       *label,
		Mode:        *mode,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compare failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(outDir)
}

func cmdComparePivot(args []string) {
	args = normalizeArgsForFlags(args)

	fs := flag.NewFlagSet("compare pivot", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "Path to config.yaml")
	metricsSelectPath := fs.String("metrics-select", "", "Path to metrics_select.yaml (default: alongside config.yaml)")
	metricsDefPath := fs.String("metrics-def", "internal/metrics/metrics_def.yaml", "Path to metrics_def.yaml")
	profile := fs.String("profile", "", "Fuzz profile name (required)")
	label := fs.String("label", "", "Run label (optional; empty means latest)")
	metric := fs.String("metric", "", "Metric key (required)")
	resultsRootOverride := fs.String("results-root", "", "Override results root folder (default from config)")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: advocate-runner compare pivot <datasetDir|datasetDir/results> --profile default [--label baseline] --metric <MetricKey>")
		os.Exit(2)
	}
	if *profile == "" || *metric == "" {
		fmt.Fprintln(os.Stderr, "--profile and --metric are required")
		os.Exit(2)
	}

	cfg, err := app.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	msel := *metricsSelectPath
	if strings.TrimSpace(msel) == "" {
		msel = filepath.Join(filepath.Dir(*configPath), "metrics_select.yaml")
	}
	if err := metrics.Configure(*metricsDefPath, msel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load metrics YAMLs: %v\n", err)
		os.Exit(1)
	}

	datasetDir, resultsRoot, err := resolveDatasetAndResultsRoot(rest[0], cfg, *resultsRootOverride)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bad target: %v\n", err)
		os.Exit(2)
	}

	outDir, err := compare.Pivot(compare.PivotParams{
		TestDir:     datasetDir,
		ResultsRoot: resultsRoot,
		Profile:     *profile,
		Label:       *label,
		Metric:      *metric,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Pivot failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(outDir)
}

// --------------------------
// Helpers: resolve targets
// --------------------------

func resolveDatasetAndResultsRoot(path string, cfg *app.Config, overrideResultsRoot string) (datasetDir string, resultsRoot string, err error) {
	p := filepath.Clean(path)

	rr := cfg.Results.Root
	if overrideResultsRoot != "" {
		rr = overrideResultsRoot
	}
	if rr == "" {
		rr = "results"
	}

	base := filepath.Base(p)
	parent := filepath.Dir(p)

	if base != rr && filepath.Base(parent) == rr {
		// "<datasetDir>/results/<TestName>"
		p = filepath.Dir(parent)
	} else if base == rr {
		// "<datasetDir>/results"
		p = parent
	}

	datasetDir = p
	resultsRoot = rr

	if st, e := os.Stat(filepath.Join(datasetDir, resultsRoot)); e != nil || !st.IsDir() {
		return "", "", fmt.Errorf("expected results folder at %s", filepath.Join(datasetDir, resultsRoot))
	}

	return datasetDir, resultsRoot, nil
}

func resolvePerTestTarget(path string, cfg *app.Config, overrideResultsRoot string) (datasetDir, resultsRoot, testName string, err error) {
	p := filepath.Clean(path)

	rr := cfg.Results.Root
	if overrideResultsRoot != "" {
		rr = overrideResultsRoot
	}
	if rr == "" {
		rr = "results"
	}

	testName = filepath.Base(p)
	if testName == "" || strings.EqualFold(testName, rr) {
		return "", "", "", fmt.Errorf("expected <datasetDir>/%s/<TestName>", rr)
	}

	resultsDir := filepath.Dir(p)
	if filepath.Base(resultsDir) != rr {
		return "", "", "", fmt.Errorf("expected parent folder named %s", rr)
	}

	datasetDir = filepath.Dir(resultsDir)
	resultsRoot = rr

	if st, e := os.Stat(filepath.Join(datasetDir, resultsRoot)); e != nil || !st.IsDir() {
		return "", "", "", fmt.Errorf("results folder not found at %s", filepath.Join(datasetDir, resultsRoot))
	}

	return datasetDir, resultsRoot, testName, nil
}

// Menu metric list helper fallback
func unionMetricKeysFromSample(runs []discovery.Run, limit int) []string {
	if limit <= 0 {
		limit = 100
	}
	if len(runs) < limit {
		limit = len(runs)
	}
	set := map[string]struct{}{}
	for i := 0; i < limit; i++ {
		ms, err := metrics.Extract(runs[i])
		if err != nil {
			continue
		}
		for k := range ms.Numbers {
			set[k] = struct{}{}
		}
		for k := range ms.Strings {
			set[k] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
