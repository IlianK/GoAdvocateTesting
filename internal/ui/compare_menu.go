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

func CompareMenu(p CompareMenuParams) CompareSelection {
	PrintHeader(fmt.Sprintf("Compare dataset: %s", p.TestDir))

	kinds := p.Index.Kinds()
	if len(kinds) == 0 {
		PrintError("No runs discovered.")
		return CompareSelection{}
	}

	kind := chooseOneFromList("Select kind:", kinds)

	// Actions depend on kind:
	actions := []string{
		"Cross-test (compare across tests)",
		"Per-test (compare modes within one test)",
	}
	actionMap := []CompareAction{
		CompareActionCrossTest,
		CompareActionPerTest,
	}

	// Pivot only makes sense for fuzzing
	if kind == "fuzzing" {
		actions = append(actions, "Pivot (compare fuzzing modes across tests for one metric)")
		actionMap = append(actionMap, CompareActionPivot)
	}

	ch := chooseOne("Select compare action:", actions)
	action := actionMap[ch-1]

	switch action {
	case CompareActionPerTest:
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

	case CompareActionCrossTest:
		profile := chooseOneFromList("Select profile:", p.Index.Profiles(kind))
		label := chooseLabel(p.Index.Labels(kind, profile))

		mode := ""
		if kind == "fuzzing" {
			modes := append([]string{"(ALL modes)"}, p.Index.Modes(profile, label)...)
			chm := chooseOneFromList("Restrict to one mode?", modes)
			if chm != "(ALL modes)" {
				mode = chm
			}
		}

		return CompareSelection{
			Action:  CompareActionCrossTest,
			Kind:    kind,
			Profile: profile,
			Label:   label,
			Mode:    mode,
		}

	case CompareActionPivot:
		// fuzzing only
		profile := chooseOneFromList("Select fuzzing profile:", p.Index.Profiles("fuzzing"))
		label := chooseLabel(p.Index.Labels("fuzzing", profile))

		metricKeys := uniqueSorted(p.MetricKeys)
		if len(metricKeys) == 0 {
			PrintError("No metrics available for pivot (check metrics_select.yaml).")
			return CompareSelection{}
		}
		metric := chooseOneFromList("Select metric to pivot:", metricKeys)

		return CompareSelection{
			Action:  CompareActionPivot,
			Kind:    "fuzzing",
			Profile: profile,
			Label:   label,
			Metric:  metric,
		}

	default:
		return CompareSelection{Action: CompareActionCrossTest, Kind: kind}
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
