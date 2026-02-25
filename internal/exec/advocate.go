package exec

import (
	"path/filepath"
	"strconv"

	"GoAdvocateTesting/internal/app"
)

// BuildAnalysisCommand builds:
// advocate analysis -path <dir> -timeoutRec X -timeoutRep Y -keepTrace -stats -scen m
func BuildAnalysisCommand(cfg *app.Config, testDir string, testName, profileName string, p app.AnalysisProfile) (bin string, argv []string) {
	bin = cfg.Runtime.AdvocateBin

	argv = []string{
		"analysis",
		"-path", filepath.Clean(testDir),
		"-exec", testName,
		"-prog", testName,

		"-timeoutRec", strconv.Itoa(p.TimeoutRec),
		"-timeoutRep", strconv.Itoa(p.TimeoutRep),
	}

	if p.KeepTrace {
		argv = append(argv, "-keepTrace")
	}
	if p.Stats {
		argv = append(argv, "-stats")
	}
	if p.Scen != "" {
		argv = append(argv, "-scen", p.Scen)
	}
	if p.Time {
		argv = append(argv, "-time")
	}

	return bin, argv
}

// BuildFuzzingCommand builds:
// advocate fuzzing -path <dir> -exec <TestName> -mode <Mode> -prog <TestName> ...
func BuildFuzzingCommand(cfg *app.Config, testDir, testName, mode string, profileName string, p app.FuzzProfile) (bin string, argv []string) {
	bin = cfg.Runtime.AdvocateBin

	argv = []string{
		"fuzzing",
		"-path", filepath.Clean(testDir),
		"-exec", testName,
		"-mode", mode,
		"-prog", testName,

		"-maxFuzzingRuns", strconv.Itoa(p.MaxRuns),
		"-timeoutFuz", strconv.Itoa(p.TimeoutFuz),
		"-timeoutRec", strconv.Itoa(p.TimeoutRec),
		"-timeoutRep", strconv.Itoa(p.TimeoutRep),
	}

	if p.KeepTrace {
		argv = append(argv, "-keepTrace")
	}
	if p.Stats {
		argv = append(argv, "-stats")
	}
	if p.Time {
		argv = append(argv, "-time")
	}

	return bin, argv
}
