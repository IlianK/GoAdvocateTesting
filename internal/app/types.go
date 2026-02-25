package app

import "time"

// ---------- Discovery ----------

type TestCase struct {
	Name string
	File string
}

// ---------- Config (config.yaml) ----------

type Config struct {
	Runtime struct {
		AdvocateBin string `yaml:"advocate_bin"`
		PatchedGo   string `yaml:"patched_go"`
		GoRuntime   string `yaml:"go_runtime"`
	} `yaml:"runtime"`

	Modes []string `yaml:"modes"`

	Results struct {
		Root                  string `yaml:"root"`
		KeepRawAdvocateResult bool   `yaml:"keep_raw_advocate_result"`
	} `yaml:"results"`
}

// ---------- Profiles (profiles.yaml) ----------

type Profiles struct {
	AnalysisProfiles map[string]AnalysisProfile `yaml:"analysisProfiles"`
	FuzzProfiles     map[string]FuzzProfile     `yaml:"fuzzProfiles"`
}

type AnalysisProfile struct {
	Scen       string `yaml:"scen"`
	TimeoutRec int    `yaml:"timeoutRec"`
	TimeoutRep int    `yaml:"timeoutRep"`
	KeepTrace  bool   `yaml:"keepTrace"`
	Stats      bool   `yaml:"stats"`
	Time       bool   `yaml:"time"`
}

type FuzzProfile struct {
	MaxRuns    int  `yaml:"maxRuns"`
	TimeoutFuz int  `yaml:"timeoutFuz"`
	TimeoutRec int  `yaml:"timeoutRec"`
	TimeoutRep int  `yaml:"timeoutRep"`
	KeepTrace  bool `yaml:"keepTrace"`
	Stats      bool `yaml:"stats"`
	Time       bool `yaml:"time"`
}

// ---------- Run Metadata ----------

type RunMeta struct {
	Tool     string `json:"tool"`
	Kind     string `json:"kind"` // fuzzing | analysis
	TestDir  string `json:"testDir"`
	TestRel  string `json:"testRel,omitempty"` // relative path from dataset root to TestDir (e.g. cockroach/10214)
	TestName string `json:"testName"`

	Mode     string `json:"mode,omitempty"`
	Profile  string `json:"profile,omitempty"`
	RunID    string `json:"runId"`
	RunLabel string `json:"runLabel,omitempty"`

	Argv      []string  `json:"argv"`
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`
	ExitCode  int       `json:"exitCode"`
}

type RunInfo struct {
	RunID string
	Meta  RunMeta
}
