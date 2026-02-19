package app

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	// Defaults
	if cfg.Results.Root == "" {
		cfg.Results.Root = "results"
	}

	// Validate
	if cfg.Runtime.AdvocateBin == "" {
		return nil, fmt.Errorf("config: runtime.advocate_bin is required")
	}
	if len(cfg.Modes) == 0 {
		return nil, fmt.Errorf("config: modes must not be empty")
	}

	return &cfg, nil
}

func LoadProfiles(path string) (*Profiles, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var p Profiles
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, err
	}

	if len(p.AnalysisProfiles) == 0 {
		return nil, fmt.Errorf("profiles: analysisProfiles must not be empty")
	}
	if len(p.FuzzProfiles) == 0 {
		return nil, fmt.Errorf("profiles: fuzzProfiles must not be empty")
	}

	return &p, nil
}
