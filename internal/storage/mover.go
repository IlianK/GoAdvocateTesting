package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Moves <testDir>/advocateResult into the destination for fuzzing and flattens
func MoveAdvocateResult(p MoveParams) (string, error) {
	src := filepath.Join(p.TestDir, "advocateResult")
	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("advocateResult not found at %s: %w", src, err)
	}

	dest, err := DestinationDir(p)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return "", err
	}

	// If advocateResult contains exactly one directory and it's a wrapper, flatten
	if len(entries) == 1 && entries[0].IsDir() {
		wrapper := filepath.Join(src, entries[0].Name())
		if looksLikeAdvocateRunDir(wrapper) {
			if err := moveDirChildren(wrapper, dest); err != nil {
				return "", err
			}
			if !p.KeepRaw {
				_ = os.RemoveAll(src)
			}
			return dest, nil
		}
	}

	// Otherwise, move advocateResult/* into dest
	for _, e := range entries {
		oldPath := filepath.Join(src, e.Name())
		newPath := filepath.Join(dest, e.Name())
		if err := os.Rename(oldPath, newPath); err != nil {
			return "", err
		}
	}

	if !p.KeepRaw {
		_ = os.RemoveAll(src)
	}
	return dest, nil
}

// ---------- HELPERS ----------
func MoveAdvocateAnalysisForTest(p MoveParams) (string, error) {
	if p.Kind != "analysis" {
		return "", fmt.Errorf("MoveAdvocateAnalysisForTest: Kind must be 'analysis'")
	}
	if p.TestName == "" {
		return "", fmt.Errorf("MoveAdvocateAnalysisForTest: TestName is required")
	}

	src := filepath.Join(p.TestDir, "advocateResult")
	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("advocateResult not found at %s: %w", src, err)
	}

	dest, err := DestinationDir(p)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return "", err
	}

	var matched string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		tname := extractTestNameFromAdvocateEntry(name)
		if tname == p.TestName {
			matched = filepath.Join(src, name)
			break
		}
	}

	if matched == "" {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if strings.Contains(e.Name(), p.TestName) {
				matched = filepath.Join(src, e.Name())
				break
			}
		}
	}

	if matched == "" {
		return "", fmt.Errorf("no analysis entry found in advocateResult for test %q", p.TestName)
	}

	if looksLikeAdvocateRunDir(matched) {
		if err := moveDirChildren(matched, dest); err != nil {
			return "", err
		}
	} else {
		if err := os.Rename(matched, filepath.Join(dest, filepath.Base(matched))); err != nil {
			return "", err
		}
	}

	if !p.KeepRaw {
		_ = os.RemoveAll(src)
	}
	return dest, nil
}

type AnalysisSplitParams struct {
	TestDir     string
	DatasetDir  string // where results should be written
	TestRel     string // relative test path within dataset
	ResultsRoot string
	Profile     string
	RunID       string
	Label       string
	KeepRaw     bool
}

// Returns map[TestName]RunDir
func MoveAdvocateAnalysisSplitAll(p AnalysisSplitParams) (map[string]string, error) {
	src := filepath.Join(p.TestDir, "advocateResult")
	if _, err := os.Stat(src); err != nil {
		return nil, fmt.Errorf("advocateResult not found at %s: %w", src, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return nil, err
	}

	moved := map[string]string{}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		entryName := e.Name()
		testName := extractTestNameFromAdvocateEntry(entryName)
		if testName == "" {
			continue
		}

		dest, err := DestinationDir(MoveParams{
			TestDir:     p.TestDir,
			DatasetDir:  p.DatasetDir,
			TestRel:     p.TestRel,
			ResultsRoot: p.ResultsRoot,
			TestName:    testName,
			Kind:        "analysis",
			Profile:     p.Profile,
			RunID:       p.RunID,
			Label:       p.Label,
		})
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return nil, err
		}

		wrapper := filepath.Join(src, entryName)
		if looksLikeAdvocateRunDir(wrapper) {
			if err := moveDirChildren(wrapper, dest); err != nil {
				return nil, err
			}
		} else {
			if err := os.Rename(wrapper, filepath.Join(dest, entryName)); err != nil {
				return nil, err
			}
		}

		moved[testName] = dest
	}

	if !p.KeepRaw {
		_ = os.RemoveAll(src)
	}
	return moved, nil
}

// ---------- INTERNAL UTILS ----------

func looksLikeAdvocateRunDir(dir string) bool {
	for _, sub := range []string{"output", "stats", "traces"} {
		if fi, err := os.Stat(filepath.Join(dir, sub)); err == nil && fi.IsDir() {
			return true
		}
	}
	return false
}

func moveDirChildren(srcDir, destDir string) error {
	ents, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, e := range ents {
		oldPath := filepath.Join(srcDir, e.Name())
		newPath := filepath.Join(destDir, e.Name())
		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	}
	return nil
}

func extractTestNameFromAdvocateEntry(entry string) string {
	idx := strings.LastIndex(entry, "-Test")
	if idx < 0 {
		return ""
	}
	return entry[idx+1:]
}
