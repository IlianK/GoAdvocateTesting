package discovery

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"GoAdvocateTesting/internal/app"
)

type Run struct {
	Test    string // test key (may include rel path)
	Kind    string // analysis|fuzzing
	Mode    string // fuzzing only
	Profile string
	Label   string // runLabel (may be empty)
	RunID   string

	Path string

	ExitCode  int
	StartedAt time.Time
	EndedAt   time.Time

	Meta app.RunMeta
}

type Filter struct {
	Kind    string // "", "analysis", "fuzzing"
	Profile string
	Label   string // "" means no label filtering
	Test    string
	Mode    string // "" means any
}

func testKeyFromMeta(meta app.RunMeta) string {
	if strings.TrimSpace(meta.TestRel) == "" {
		return meta.TestName
	}
	// normalize with slashes for stable keys/CSV display
	return filepath.ToSlash(filepath.Join(meta.TestRel, meta.TestName))
}

func DiscoverRuns(testDir string, resultsRoot string) ([]Run, error) {
	resultsPath := filepath.Join(testDir, resultsRoot)
	st, err := os.Stat(resultsPath)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("results path is not a directory: %s", resultsPath)
	}

	var out []Run

	err = filepath.WalkDir(resultsPath, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "meta.json" {
			return nil
		}

		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		var meta app.RunMeta
		if err := json.Unmarshal(b, &meta); err != nil {
			return nil
		}

		runDir := filepath.Dir(p)

		out = append(out, Run{
			Test:      testKeyFromMeta(meta),
			Kind:      meta.Kind,
			Mode:      meta.Mode,
			Profile:   meta.Profile,
			Label:     meta.RunLabel,
			RunID:     meta.RunID,
			Path:      runDir,
			ExitCode:  meta.ExitCode,
			StartedAt: meta.StartedAt,
			EndedAt:   meta.EndedAt,
			Meta:      meta,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Test != b.Test {
			return a.Test < b.Test
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Profile != b.Profile {
			return a.Profile < b.Profile
		}
		if a.Mode != b.Mode {
			return a.Mode < b.Mode
		}
		return a.StartedAt.Before(b.StartedAt)
	})

	return out, nil
}

func FilterRuns(runs []Run, f Filter) []Run {
	var out []Run
	for _, r := range runs {
		if f.Kind != "" && r.Kind != f.Kind {
			continue
		}
		if f.Profile != "" && r.Profile != f.Profile {
			continue
		}
		if f.Label != "" && r.Label != f.Label {
			continue
		}
		if f.Test != "" && r.Test != f.Test {
			continue
		}
		if f.Mode != "" && r.Mode != f.Mode {
			continue
		}
		out = append(out, r)
	}
	return out
}

// LatestPerGroup selects newest run by StartedAt within each group
// Keys: "test","kind","profile","mode","label"
func LatestPerGroup(runs []Run, keys ...string) []Run {
	type key string
	m := map[key]Run{}

	build := func(r Run) key {
		var parts []string
		for _, k := range keys {
			switch strings.ToLower(k) {
			case "test":
				parts = append(parts, "test="+r.Test)
			case "kind":
				parts = append(parts, "kind="+r.Kind)
			case "profile":
				parts = append(parts, "profile="+r.Profile)
			case "mode":
				parts = append(parts, "mode="+r.Mode)
			case "label":
				parts = append(parts, "label="+r.Label)
			}
		}
		return key(strings.Join(parts, "|"))
	}

	for _, r := range runs {
		k := build(r)
		prev, ok := m[k]
		if !ok || prev.StartedAt.Before(r.StartedAt) {
			m[k] = r
		}
	}

	out := make([]Run, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}

	sort.Slice(out, func(i, j int) bool {
		ti := out[i].StartedAt
		tj := out[j].StartedAt
		if ti.Equal(tj) {
			return out[i].Test < out[j].Test
		}
		return ti.Before(tj)
	})

	return out
}

// ---- RunIndex (for interactive compare menus) ----
type RunIndex struct {
	Runs []Run
}

func NewRunIndex(runs []Run) RunIndex {
	return RunIndex{Runs: runs}
}

func (idx RunIndex) Kinds() []string {
	m := map[string]struct{}{}
	for _, r := range idx.Runs {
		if r.Kind != "" {
			m[r.Kind] = struct{}{}
		}
	}
	return sortedKeys(m)
}

func (idx RunIndex) Profiles(kind string) []string {
	m := map[string]struct{}{}
	for _, r := range idx.Runs {
		if kind != "" && r.Kind != kind {
			continue
		}
		if r.Profile != "" {
			m[r.Profile] = struct{}{}
		}
	}
	return sortedKeys(m)
}

func (idx RunIndex) Labels(kind, profile string) []string {
	m := map[string]struct{}{}
	for _, r := range idx.Runs {
		if kind != "" && r.Kind != kind {
			continue
		}
		if profile != "" && r.Profile != profile {
			continue
		}
		if strings.TrimSpace(r.Label) != "" {
			m[r.Label] = struct{}{}
		}
	}
	return sortedKeys(m)
}

// Tests returns test names with runs matching kind/profile
// If label == "" (meaning "latest"), do not restrict by label
func (idx RunIndex) Tests(kind, profile, label string) []string {
	m := map[string]struct{}{}
	for _, r := range idx.Runs {
		if kind != "" && r.Kind != kind {
			continue
		}
		if profile != "" && r.Profile != profile {
			continue
		}
		if label != "" && r.Label != label {
			continue
		}
		if r.Test != "" {
			m[r.Test] = struct{}{}
		}
	}
	return sortedKeys(m)
}

// Modes returns fuzz modes present for a fuzz profile
// If label == "" do not restrict by label
func (idx RunIndex) Modes(profile, label string) []string {
	m := map[string]struct{}{}
	for _, r := range idx.Runs {
		if r.Kind != "fuzzing" {
			continue
		}
		if profile != "" && r.Profile != profile {
			continue
		}
		if label != "" && r.Label != label {
			continue
		}
		if r.Mode != "" {
			m[r.Mode] = struct{}{}
		}
	}
	return sortedKeys(m)
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
