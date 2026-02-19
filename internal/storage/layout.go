package storage

import (
	"fmt"
	"path/filepath"
)

type MoveParams struct {
	// TestDir is the directory where ADVOCATE was executed and where advocateResult exists.
	TestDir string

	// DatasetDir is the directory where results/ should be written (the CLI root path).
	// If empty, we fall back to TestDir for backward compatibility.
	DatasetDir string

	// TestRel is the relative path from DatasetDir to TestDir (e.g. "cockroach/10214").
	// Used to keep results interpretable and avoid test-name collisions.
	TestRel string

	ResultsRoot string
	TestName    string

	Kind    string // "analysis" | "fuzzing"
	Mode    string // fuzzing only
	Profile string

	RunID   string // unique run id (timestamp)
	Label   string // stable compare label (optional)
	KeepRaw bool
}

func DestinationDir(p MoveParams) (string, error) {
	if p.ResultsRoot == "" || p.TestName == "" || p.Kind == "" {
		return "", fmt.Errorf("invalid destination params: missing required fields")
	}
	if p.Profile == "" {
		p.Profile = "default"
	}

	// Backward compatibility: if DatasetDir is empty, behave like old code.
	baseRoot := p.DatasetDir
	if baseRoot == "" {
		if p.TestDir == "" {
			return "", fmt.Errorf("invalid destination params: missing datasetDir/testDir")
		}
		baseRoot = p.TestDir
	}

	runFolder := runFolderName(p)

	base := filepath.Join(baseRoot, p.ResultsRoot)

	// If TestRel is present, nest under it (e.g. results/cockroach/10214/...)
	if p.TestRel != "" && p.TestRel != "." {
		base = filepath.Join(base, p.TestRel)
	}

	// Always keep test name as a final folder before kind/profile/run.
	base = filepath.Join(base, p.TestName)

	switch p.Kind {
	case "analysis":
		return filepath.Join(base, "analysis", p.Profile, runFolder), nil
	case "fuzzing":
		if p.Mode == "" {
			return "", fmt.Errorf("missing mode for fuzzing destination")
		}
		return filepath.Join(base, "fuzzing", p.Mode, p.Profile, runFolder), nil
	default:
		return "", fmt.Errorf("unknown kind: %s", p.Kind)
	}
}

func runFolderName(p MoveParams) string {
	if p.Label != "" {
		return "run-" + p.Label
	}
	// fallback to unique id
	if p.RunID != "" {
		return "run-" + p.RunID
	}
	return "run-unknown"
}
