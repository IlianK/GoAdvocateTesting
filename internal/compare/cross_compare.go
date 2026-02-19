package compare

import (
	"fmt"
	"path/filepath"
	"sort"

	"GoAdvocateTesting/internal/discovery"
	"GoAdvocateTesting/internal/metrics"
)

type CrossTestParams struct {
	TestDir     string
	ResultsRoot string

	Kind    string // analysis|fuzzing
	Profile string
	Label   string // if empty -> latest per group

	Mode string // fuzzing cross-test optional restriction

	OutRoot string // <testDir>/comparisons
}

func CompareCrossTest(p CrossTestParams) (outDir string, err error) {
	if p.OutRoot == "" {
		p.OutRoot = filepath.Join(p.TestDir, "comparisons")
	}

	runs, err := discovery.DiscoverRuns(p.TestDir, p.ResultsRoot)
	if err != nil {
		return "", err
	}

	runs = discovery.FilterRuns(runs, discovery.Filter{
		Kind:    p.Kind,
		Profile: p.Profile,
		Label:   p.Label,
		Mode:    p.Mode,
	})

	if p.Label == "" {
		if p.Kind == "analysis" {
			runs = discovery.LatestPerGroup(runs, "test", "kind", "profile")
		} else {
			runs = discovery.LatestPerGroup(runs, "test", "kind", "profile", "mode")
		}
	}

	if len(runs) == 0 {
		return "", fmt.Errorf("no runs found for cross-test compare (kind=%s profile=%s label=%s mode=%s)", p.Kind, p.Profile, p.Label, p.Mode)
	}

	sets := make([]metrics.MetricSet, 0, len(runs))
	rows := make([]Row, 0, len(runs))
	for _, r := range runs {
		ms, _ := metrics.Extract(r)
		sets = append(sets, ms)
		rows = append(rows, rowForCompareCSV(ms, true))
	}

	// Sort by Test, then Mode
	sort.Slice(rows, func(i, j int) bool {
		ti, _ := rows[i].Get("Test")
		tj, _ := rows[j].Get("Test")
		if ti != tj {
			return ti < tj
		}
		mi, _ := rows[i].Get("Mode")
		mj, _ := rows[j].Get("Mode")
		return mi < mj
	})

	labelFolder := p.Label
	if labelFolder == "" {
		labelFolder = "latest"
	}

	outDir = filepath.Join(
		p.OutRoot,
		"cross-test",
		"kind-"+p.Kind,
		"profile-"+p.Profile,
		"label-"+labelFolder,
	)

	csvPath := filepath.Join(outDir, "cross_test.csv")
	if err := WriteCSVOrdered(csvPath, rows, compareHeaderCrossTest()); err != nil {
		return "", err
	}

	inputs := metrics.BuildInputsFile(p.Kind, p.Profile, labelFolder, "", runs, sets)
	if err := metrics.WriteJSON(filepath.Join(outDir, "inputs.json"), inputs); err != nil {
		return "", err
	}

	return outDir, nil
}

func compareHeaderCrossTest() []string {
	return []string{
		"Test",
		"Mode",
		"Unique_Bugs",
		"Bug_Types",
		"Total_Bugs",
		"Panics",
		"Leaks",
		"Confirmed_Replays",
		"Total_Runs",
		"Total_Time_s",
		"Rec_s",
		"Ana_s",
		"Rep_s",
		"Replays_Written",
		"Replays_Successful",
	}
}

// rowForCompareCSV maps MetricSet -> Row with the exact column names used by the compare CSVs.
func rowForCompareCSV(ms metrics.MetricSet, includeTest bool) Row {
	r := Row{
		Fixed:   map[string]string{},
		Numbers: map[string]float64{},
		Strings: map[string]string{},
	}

	// Identity
	if includeTest {
		if v, ok := ms.Strings["test_name"]; ok {
			r.Fixed["Test"] = v
		}
	}
	if v, ok := ms.Strings["Mode"]; ok {
		r.Fixed["Mode"] = v
	} else if v2, ok2 := ms.Strings["mode"]; ok2 {
		r.Fixed["Mode"] = v2
	} else {
		r.Fixed["Mode"] = ""
	}

	// Prefer explicitly named outputs from Extract()
	if v, ok := ms.Strings["Bug_Types"]; ok {
		r.Strings["Bug_Types"] = v
	} else if v2, ok2 := ms.Strings["bug_types"]; ok2 {
		r.Strings["Bug_Types"] = v2
	}

	for _, k := range []string{
		"Unique_Bugs",
		"Total_Bugs",
		"Panics",
		"Leaks",
		"Confirmed_Replays",
		"Total_Runs",
		"Total_Time_s",
		"Rec_s",
		"Ana_s",
		"Rep_s",
		"Replays_Written",
		"Replays_Successful",
	} {
		if v, ok := ms.Numbers[k]; ok {
			r.Numbers[k] = v
		} else if v2, ok2 := ms.Numbers[stringsToLegacyKey(k)]; ok2 {
			r.Numbers[k] = v2
		}
	}

	// Ensure missing numeric columns appear as blanks (WriteCSVOrdered handles it)
	return r
}

func stringsToLegacyKey(k string) string {
	// kept for backwards compatibility if you had older keys; currently no-op-ish.
	return k
}
