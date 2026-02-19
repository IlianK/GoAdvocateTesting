package exec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"GoAdvocateTesting/internal/app"
)

type Runner struct {
	cfg *app.Config
}

func NewRunner(cfg *app.Config) *Runner {
	return &Runner{cfg: cfg}
}

func (r *Runner) RunAnalysis(testDir, testName, profileName string, prof app.AnalysisProfile, label string) (*app.RunInfo, error) {
	runID := time.Now().UTC().Format("2006-01-02T15-04-05Z")

	bin, argv := BuildAnalysisCommand(r.cfg, testDir, profileName, prof)

	cmd := exec.Command(bin, argv...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now().UTC()
	err := cmd.Run()
	end := time.Now().UTC()

	exitCode := 0
	if err != nil {
		exitCode = exitCodeFromError(err)
	}

	meta := app.RunMeta{
		Tool:      "advocate",
		Kind:      "analysis",
		TestDir:   testDir,
		TestName:  testName,
		Profile:   profileName,
		RunID:     runID,
		RunLabel:  label,
		Argv:      append([]string{bin}, argv...),
		StartedAt: start,
		EndedAt:   end,
		ExitCode:  exitCode,
	}

	if err != nil {
		return &app.RunInfo{RunID: runID, Meta: meta}, fmt.Errorf("advocate analysis failed (exit=%d): %w", exitCode, err)
	}
	return &app.RunInfo{RunID: runID, Meta: meta}, nil
}

func (r *Runner) RunFuzzing(testDir, testName, mode, profileName string, prof app.FuzzProfile, label string) (*app.RunInfo, error) {
	runID := time.Now().UTC().Format("2006-01-02T15-04-05Z")

	bin, argv := BuildFuzzingCommand(r.cfg, testDir, testName, mode, profileName, prof)

	cmd := exec.Command(bin, argv...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now().UTC()
	err := cmd.Run()
	end := time.Now().UTC()

	exitCode := 0
	if err != nil {
		exitCode = exitCodeFromError(err)
	}

	meta := app.RunMeta{
		Tool:      "advocate",
		Kind:      "fuzzing",
		TestDir:   testDir,
		TestName:  testName,
		Mode:      mode,
		Profile:   profileName,
		RunID:     runID,
		RunLabel:  label,
		Argv:      append([]string{bin}, argv...),
		StartedAt: start,
		EndedAt:   end,
		ExitCode:  exitCode,
	}

	if err != nil {
		return &app.RunInfo{RunID: runID, Meta: meta}, fmt.Errorf("advocate fuzzing failed (exit=%d): %w", exitCode, err)
	}
	return &app.RunInfo{RunID: runID, Meta: meta}, nil
}

func exitCodeFromError(err error) int {
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		return 1
	}
	if ee.ProcessState != nil {
		return ee.ProcessState.ExitCode()
	}
	return 1
}
