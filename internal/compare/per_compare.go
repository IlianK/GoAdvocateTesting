package compare

import (
	"fmt"
	"path/filepath"

	"GoAdvocateTesting/internal/discovery"
	"GoAdvocateTesting/internal/metrics"
)

type PerTestParams struct {
	TestDir     string
	ResultsRoot string

	TestName string

	Kind    string // typically "fuzzing" (but allow "analysis")
	Profile string
	Label   string // if empty -> latest per group

	OutRoot string // <testDir>/comparisons
}

func ComparePerTest(p PerTestParams) (outDir string, err error) {
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
		Test:    p.TestName,
	})

	if p.Label == "" {
		if p.Kind == "analysis" {
			runs = discovery.LatestPerGroup(runs, "test", "kind", "profile")
		} else {
			runs = discovery.LatestPerGroup(runs, "test", "kind", "profile", "mode")
		}
	}

	if len(runs) == 0 {
		return "", fmt.Errorf("no runs found for per-test compare (test=%s kind=%s profile=%s label=%s)", p.TestName, p.Kind, p.Profile, p.Label)
	}

	sets := make([]metrics.MetricSet, 0, len(runs))
	rows := make([]Row, 0, len(runs))
	for _, r := range runs {
		ms, _ := metrics.Extract(r)
		sets = append(sets, ms)
		rows = append(rows, RowForCompareCSV(ms, false))
	}

	labelFolder := p.Label
	if labelFolder == "" {
		labelFolder = "latest"
	}

	outDir = filepath.Join(
		p.OutRoot,
		"per-test",
		p.TestName,
		"kind-"+p.Kind,
		"profile-"+p.Profile,
		"label-"+labelFolder,
	)

	header, err := metrics.ActiveCSVHeader(false, true)
	if err != nil {
		return "", err
	}

	csvPath := filepath.Join(outDir, "compare.csv")
	if err := WriteCSVOrdered(csvPath, rows, header); err != nil {
		return "", err
	}

	inputs := metrics.BuildInputsFile(p.Kind, p.Profile, labelFolder, "", runs, sets)
	if err := metrics.WriteJSON(filepath.Join(outDir, "inputs.json"), inputs); err != nil {
		return "", err
	}

	return outDir, nil
}
