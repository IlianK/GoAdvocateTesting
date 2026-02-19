package compare

import (
	"fmt"
	"path/filepath"
	"sort"

	"GoAdvocateTesting/internal/discovery"
	"GoAdvocateTesting/internal/metrics"
)

type PivotParams struct {
	TestDir     string
	ResultsRoot string

	Profile string
	Label   string // if empty -> latest per (test,mode)

	Metric string // supports numeric OR string (e.g. "Bug_Types", "Unique_Bugs", "Total_Time_s")

	OutRoot string // <testDir>/comparisons
}

func Pivot(p PivotParams) (outDir string, err error) {
	if p.OutRoot == "" {
		p.OutRoot = filepath.Join(p.TestDir, "comparisons")
	}
	if p.Metric == "" {
		return "", fmt.Errorf("pivot requires --metric")
	}

	runs, err := discovery.DiscoverRuns(p.TestDir, p.ResultsRoot)
	if err != nil {
		return "", err
	}

	runs = discovery.FilterRuns(runs, discovery.Filter{
		Kind:    "fuzzing",
		Profile: p.Profile,
		Label:   p.Label,
	})

	if p.Label == "" {
		runs = discovery.LatestPerGroup(runs, "test", "kind", "profile", "mode")
	}

	if len(runs) == 0 {
		return "", fmt.Errorf("no runs found for pivot (profile=%s label=%s)", p.Profile, p.Label)
	}

	tests := map[string]struct{}{}
	modes := map[string]struct{}{}

	// grid: test -> mode -> (string or number)
	type cell struct {
		isNum bool
		num   float64
		str   string
		ok    bool
	}
	grid := map[string]map[string]cell{}

	sets := make([]metrics.MetricSet, 0, len(runs))
	for _, r := range runs {
		ms, _ := metrics.Extract(r)
		sets = append(sets, ms)

		tests[r.Test] = struct{}{}
		modes[r.Mode] = struct{}{}

		if _, ok := grid[r.Test]; !ok {
			grid[r.Test] = map[string]cell{}
		}

		if v, ok := ms.Numbers[p.Metric]; ok {
			grid[r.Test][r.Mode] = cell{isNum: true, num: v, ok: true}
		} else if s, ok := ms.Strings[p.Metric]; ok {
			grid[r.Test][r.Mode] = cell{isNum: false, str: s, ok: true}
		} else {
			grid[r.Test][r.Mode] = cell{ok: false}
		}
	}

	testList := make([]string, 0, len(tests))
	for t := range tests {
		testList = append(testList, t)
	}
	sort.Strings(testList)

	modeList := make([]string, 0, len(modes))
	for m := range modes {
		modeList = append(modeList, m)
	}
	sort.Strings(modeList)

	// Build rows: header is Test + modes...
	header := make([]string, 0, 1+len(modeList))
	header = append(header, "Test")
	header = append(header, modeList...)

	rows := make([]Row, 0, len(testList))
	for _, t := range testList {
		row := Row{
			Fixed:   map[string]string{"Test": t},
			Numbers: map[string]float64{},
			Strings: map[string]string{},
		}
		for _, m := range modeList {
			c := grid[t][m]
			if !c.ok {
				continue
			}
			if c.isNum {
				row.Numbers[m] = c.num
			} else {
				row.Strings[m] = c.str
			}
		}
		rows = append(rows, row)
	}

	labelFolder := p.Label
	if labelFolder == "" {
		labelFolder = "latest"
	}

	outDir = filepath.Join(
		p.OutRoot,
		"pivot",
		"kind-fuzzing",
		"profile-"+p.Profile,
		"label-"+labelFolder,
	)

	csvPath := filepath.Join(outDir, fmt.Sprintf("pivot_%s.csv", sanitizeMetricName(p.Metric)))
	if err := WriteCSVOrdered(csvPath, rows, header); err != nil {
		return "", err
	}

	inputs := metrics.BuildInputsFile("fuzzing", p.Profile, labelFolder, p.Metric, runs, sets)
	if err := metrics.WriteJSON(filepath.Join(outDir, "inputs_"+sanitizeMetricName(p.Metric)+".json"), inputs); err != nil {
		return "", err
	}

	return outDir, nil
}

func sanitizeMetricName(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r)
		case r >= '0' && r <= '9':
			out = append(out, r)
		case r == '.' || r == '_' || r == '-':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}
