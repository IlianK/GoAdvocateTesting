package ui

import (
	"fmt"

	"GoAdvocateTesting/internal/discovery"
)

type CompareAction string

const (
	CompareActionPerTest   CompareAction = "per-test"
	CompareActionCrossTest CompareAction = "cross-test"
	CompareActionPivot     CompareAction = "pivot"
)

type CompareMenuParams struct {
	TestDir    string
	Index      discovery.RunIndex
	MetricKeys []string
}

type CompareSelection struct {
	Action  CompareAction
	Kind    string // analysis|fuzzing
	Profile string
	Label   string // "" means latest

	TestName string // per-test
	Mode     string // cross-test optional
	Metric   string // pivot
}

// add near top of file
var pivotAllowedMetrics = map[string]struct{}{
	"Unique_Bugs":        {},
	"Bug_Types":          {},
	"Total_Bugs":         {},
	"Panics":             {},
	"Leaks":              {},
	"Confirmed_Replays":  {},
	"Total_Runs":         {},
	"Total_Time_s":       {},
	"Rec_s":              {},
	"Ana_s":              {},
	"Rep_s":              {},
	"Replays_Written":    {},
	"Replays_Successful": {},
}

func filterPivotMetrics(keys []string) []string {
	out := make([]string, 0, len(keys))
	seen := map[string]bool{}
	for _, k := range keys {
		if _, ok := pivotAllowedMetrics[k]; ok && !seen[k] {
			out = append(out, k)
			seen[k] = true
		}
	}
	out = uniqueSorted(out)
	return out
}

func CompareMenu(p CompareMenuParams) CompareSelection {
	PrintHeader(fmt.Sprintf("Compare dataset: %s", p.TestDir))

	action := chooseOne("Select compare action:", []string{
		"Per-test (compare fuzz modes within one test)",
		"Cross-test (compare across tests)",
		"Pivot (compare fuzzing modes across tests for one metric)",
	})

	switch action {
	case 1:
		kinds := p.Index.Kinds()
		if len(kinds) == 0 {
			PrintError("No runs discovered.")
			return CompareSelection{}
		}
		kind := chooseOneFromList("Select kind:", kinds)
		profile := chooseOneFromList("Select profile:", p.Index.Profiles(kind))
		label := chooseLabel(p.Index.Labels(kind, profile))
		testName := chooseOneFromList("Select test:", p.Index.Tests(kind, profile, label))

		return CompareSelection{
			Action:   CompareActionPerTest,
			Kind:     kind,
			Profile:  profile,
			Label:    label,
			TestName: testName,
		}

	case 2:
		kinds := p.Index.Kinds()
		if len(kinds) == 0 {
			PrintError("No runs discovered.")
			return CompareSelection{}
		}
		kind := chooseOneFromList("Select kind:", kinds)
		profile := chooseOneFromList("Select profile:", p.Index.Profiles(kind))
		label := chooseLabel(p.Index.Labels(kind, profile))

		mode := ""
		if kind == "fuzzing" {
			modes := append([]string{"(ALL modes)"}, p.Index.Modes(profile, label)...)
			ch := chooseOneFromList("Restrict to one mode?", modes)
			if ch != "(ALL modes)" {
				mode = ch
			}
		}

		return CompareSelection{
			Action:  CompareActionCrossTest,
			Kind:    kind,
			Profile: profile,
			Label:   label,
			Mode:    mode,
		}

	case 3:
		// pivot is fuzzing-only
		kind := "fuzzing"
		profile := chooseOneFromList("Select fuzzing profile:", p.Index.Profiles(kind))
		label := chooseLabel(p.Index.Labels(kind, profile))
		metricKeys := filterPivotMetrics(p.MetricKeys)
		metric := chooseOneFromList("Select metric to pivot:", metricKeys)

		return CompareSelection{
			Action:  CompareActionPivot,
			Kind:    kind,
			Profile: profile,
			Label:   label,
			Metric:  metric,
		}

	default:
		return CompareSelection{Action: CompareActionCrossTest, Kind: "analysis"}
	}
}

func chooseLabel(labels []string) string {
	labels = uniqueSorted(labels)
	labels = append([]string{"(latest)"}, labels...)
	ch := chooseOneFromList("Select label:", labels)
	if ch == "(latest)" {
		return ""
	}
	return ch
}
