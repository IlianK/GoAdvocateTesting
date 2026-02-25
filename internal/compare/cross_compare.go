package compare

import (
	"fmt"
	"path/filepath"

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
		rows = append(rows, RowForCompareCSV(ms, true))
	}

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

	header, err := metrics.ActiveCSVHeader(true, true)
	if err != nil {
		return "", err
	}

	csvPath := filepath.Join(outDir, "cross_test.csv")
	if err := WriteCSVOrdered(csvPath, rows, header); err != nil {
		return "", err
	}

	inputs := metrics.BuildInputsFile(p.Kind, p.Profile, labelFolder, "", runs, sets)
	if err := metrics.WriteJSON(filepath.Join(outDir, "inputs.json"), inputs); err != nil {
		return "", err
	}

	return outDir, nil
}
