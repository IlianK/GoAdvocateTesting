package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"GoAdvocateTesting/internal/app"
	"GoAdvocateTesting/internal/discovery"
	"GoAdvocateTesting/internal/exec"
	"GoAdvocateTesting/internal/storage"
	"GoAdvocateTesting/internal/ui"
)

func cmdTest(args []string) {
	args = normalizeArgsForFlags(args)

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "Path to config.yaml")
	profilesPath := fs.String("profiles", "profiles.yaml", "Path to profiles.yaml")
	label := fs.String("label", "", "Run label to group comparable results (e.g. baseline, nightly, commit-abc123). If set, folder becomes run-<label>.")
	recursive := fs.Bool("recursive", false, "Recursively discover tests below the given path (directories containing *_test.go)")
	testsFile := fs.String("tests-file", "", "Optional txt file with one selected test per line (format: relative/package/path/TestName)")
	nonInteractive := fs.Bool("non-interactive", false, "Disable interactive menu (not implemented in this version)")
	keepRaw := fs.Bool("keep-raw", false, "Do not delete advocateResult after moving (debug)")
	_ = fs.Parse(args)

	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: advocate-runner test <path> [--recursive] [--tests-file filtered_tests.txt] [--config config.yaml] [--profiles profiles.yaml] [--label baseline]")
		os.Exit(2)
	}
	rootPath := filepath.Clean(rest[0])

	if *nonInteractive {
		fmt.Fprintln(os.Stderr, "--non-interactive is not implemented in this version. Run without it.")
		os.Exit(2)
	}

	cfg, err := app.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	profiles, err := app.LoadProfiles(*profilesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load profiles: %v\n", err)
		os.Exit(1)
	}

	var tests []app.TestCase
	if *recursive {
		tests, err = discovery.DiscoverTestsRecursive(rootPath)
	} else {
		tests, err = discovery.DiscoverTests(rootPath)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to discover tests: %v\n", err)
		os.Exit(1)
	}
	if len(tests) == 0 {
		fmt.Fprintf(os.Stderr, "No tests found under %s\n", rootPath)
		os.Exit(1)
	}

	ui.PrintHeader(fmt.Sprintf("Discovered %d tests%s under: %s",
		len(tests),
		func() string {
			if *recursive {
				return " (recursive)"
			}
			return ""
		}(),
		rootPath,
	))

	// Optional pre-filter by tests-file
	if *testsFile != "" {
		selectedIDs, err := readTestsFile(*testsFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read --tests-file: %v\n", err)
			os.Exit(1)
		}
		if len(selectedIDs) == 0 {
			fmt.Fprintf(os.Stderr, "--tests-file is empty: %s\n", *testsFile)
			os.Exit(1)
		}

		tests, err = filterDiscoveredTests(tests, selectedIDs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to apply --tests-file: %v\n", err)
			os.Exit(1)
		}
		if len(tests) == 0 {
			fmt.Fprintf(os.Stderr, "No matching tests from %s were found under %s\n", *testsFile, rootPath)
			os.Exit(1)
		}

		ui.PrintHeader(fmt.Sprintf("Selected %d tests from: %s", len(tests), *testsFile))
		for _, tc := range tests {
			fmt.Printf("- %s\n", testFilterID(tc))
		}

		if !confirmYesNo("Proceed with these selected tests? [y/n]: ") {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	var selection ui.Selection
	if *testsFile != "" {
		selection = ui.RunMenuFixedTests(ui.RunMenuParams{
			Modes:           cfg.Modes,
			Tests:           tests,
			AnalysisProfile: profiles.AnalysisProfiles,
			FuzzProfiles:    profiles.FuzzProfiles,
		})
	} else {
		selection = ui.RunMenu(ui.RunMenuParams{
			Modes:           cfg.Modes,
			Tests:           tests,
			AnalysisProfile: profiles.AnalysisProfiles,
			FuzzProfiles:    profiles.FuzzProfiles,
		})
	}

	r := exec.NewRunner(cfg)

	testDirOf := func(tc app.TestCase) string {
		if strings.Contains(tc.File, string(os.PathSeparator)) {
			return filepath.Dir(filepath.Join(rootPath, tc.File))
		}
		return rootPath
	}

	runAnalysisOne := func(tc app.TestCase) {
		td := testDirOf(tc)
		ui.PrintHeader(fmt.Sprintf("Running analysis: %s [dir=%s profile=%s label=%s]", ui.DisplayTest(tc), td, selection.Profile, *label))

		prof, ok := profiles.AnalysisProfiles[selection.Profile]
		if !ok {
			ui.PrintError(fmt.Sprintf("Unknown analysis profile: %s", selection.Profile))
			return
		}

		runInfo, err := r.RunAnalysis(td, tc.Name, selection.Profile, prof, *label)
		if err != nil {
			ui.PrintError(fmt.Sprintf("Run failed: %v", err))
		}

		dest, err2 := storage.MoveAdvocateAnalysisForTest(storage.MoveParams{
			TestDir:     td,
			ResultsRoot: cfg.Results.Root,
			TestName:    tc.Name,
			Kind:        "analysis",
			Profile:     selection.Profile,
			RunID:       runInfo.RunID,
			Label:       *label,
			KeepRaw:     *keepRaw || cfg.Results.KeepRawAdvocateResult,
		})
		if err2 != nil {
			ui.PrintError(fmt.Sprintf("Failed moving analysis results: %v", err2))
			return
		}

		if err := storage.WriteMetaJSON(dest, runInfo.Meta); err != nil {
			ui.PrintError(fmt.Sprintf("Failed writing meta.json: %v", err))
			return
		}

		ui.PrintOK(fmt.Sprintf("Saved to: %s", dest))
	}

	runFuzzingOne := func(mode string, tc app.TestCase) {
		td := testDirOf(tc)
		ui.PrintHeader(fmt.Sprintf("Running fuzzing: %s [dir=%s mode=%s profile=%s label=%s]", ui.DisplayTest(tc), td, mode, selection.Profile, *label))

		prof, ok := profiles.FuzzProfiles[selection.Profile]
		if !ok {
			ui.PrintError(fmt.Sprintf("Unknown fuzz profile: %s", selection.Profile))
			return
		}

		runInfo, err := r.RunFuzzing(td, tc.Name, mode, selection.Profile, prof, *label)
		if err != nil {
			ui.PrintError(fmt.Sprintf("Run failed: %v", err))
		}

		dest, err := storage.MoveAdvocateResult(storage.MoveParams{
			TestDir:     td,
			ResultsRoot: cfg.Results.Root,
			TestName:    tc.Name,
			Kind:        "fuzzing",
			Mode:        mode,
			Profile:     selection.Profile,
			RunID:       runInfo.RunID,
			Label:       *label,
			KeepRaw:     *keepRaw || cfg.Results.KeepRawAdvocateResult,
		})
		if err != nil {
			ui.PrintError(fmt.Sprintf("Failed moving results: %v", err))
			return
		}

		if err := storage.WriteMetaJSON(dest, runInfo.Meta); err != nil {
			ui.PrintError(fmt.Sprintf("Failed writing meta.json: %v", err))
			return
		}

		ui.PrintOK(fmt.Sprintf("Saved to: %s", dest))
	}

	if selection.Kind == ui.KindAnalysis {
		switch selection.Scope {
		case ui.ScopeAllTests:
			for _, tc := range tests {
				runAnalysisOne(tc)
			}
		case ui.ScopeOneTest:
			runAnalysisOne(selection.Test)
		default:
			panic("unknown analysis scope")
		}
	} else if selection.Kind == ui.KindFuzzing {
		switch selection.ModeType {
		case ui.AllOnAll:
			for _, tc := range tests {
				for _, m := range cfg.Modes {
					runFuzzingOne(m, tc)
				}
			}
		case ui.AllOnOne:
			for _, m := range cfg.Modes {
				runFuzzingOne(m, selection.Test)
			}
		case ui.OneOnAll:
			for _, tc := range tests {
				runFuzzingOne(selection.Mode, tc)
			}
		case ui.OneOnOne:
			runFuzzingOne(selection.Mode, selection.Test)
		default:
			panic("unknown fuzz selection")
		}
	} else {
		panic("unknown kind")
	}

	ui.PrintOK("Done.")
}

func readTestsFile(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	var out []string

	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = filepath.ToSlash(line)
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		out = append(out, line)
	}

	sort.Strings(out)
	return out, nil
}

func filterDiscoveredTests(discovered []app.TestCase, wanted []string) ([]app.TestCase, error) {
	byID := make(map[string][]app.TestCase, len(discovered))
	for _, tc := range discovered {
		id := testFilterID(tc)
		byID[id] = append(byID[id], tc)
	}

	var selected []app.TestCase
	var missing []string
	seen := make(map[string]struct{})

	for _, id := range wanted {
		matches, ok := byID[id]
		if !ok || len(matches) == 0 {
			missing = append(missing, id)
			continue
		}

		for _, tc := range matches {
			uniq := tc.File + "::" + tc.Name
			if _, exists := seen[uniq]; exists {
				continue
			}
			seen[uniq] = struct{}{}
			selected = append(selected, tc)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf(
			"some tests from --tests-file were not found under the given root:\n- %s",
			strings.Join(missing, "\n- "),
		)
	}

	sort.Slice(selected, func(i, j int) bool {
		if selected[i].File != selected[j].File {
			return selected[i].File < selected[j].File
		}
		return selected[i].Name < selected[j].Name
	})

	return selected, nil
}

func testFilterID(tc app.TestCase) string {
	dir := filepath.Dir(tc.File)
	if dir == "." || dir == "" {
		return tc.Name
	}
	return filepath.ToSlash(filepath.Join(dir, tc.Name))
}

func confirmYesNo(prompt string) bool {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		line, _ := in.ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}
