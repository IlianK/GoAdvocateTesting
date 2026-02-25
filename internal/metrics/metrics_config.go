package metrics

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	globalMu   sync.Mutex
	globalOnce sync.Once
	globalCfg  *RuntimeConfig
	globalErr  error
)

// Sets paths for metrics_def.yaml and metrics_select.yaml and loads them
// Call once from CLI startup (compare command)
func Configure(defPath, selectPath string) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	globalOnce = sync.Once{}
	globalCfg = nil
	globalErr = nil

	globalOnce.Do(func() {
		cfg, err := loadRuntimeConfig(defPath, selectPath)
		if err != nil {
			globalErr = err
			return
		}
		globalCfg = cfg
	})

	return globalErr
}

// Returns loaded config, loading defaults if Configure() was not called
func getRuntimeConfig() (*RuntimeConfig, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	globalOnce.Do(func() {
		cfg, err := loadRuntimeConfig("internal/metrics/metrics_def.yaml", "metrics_select.yaml")
		if err != nil {
			globalErr = err
			return
		}
		globalCfg = cfg
	})
	return globalCfg, globalErr
}

// ------------------------
// YAML structs
// ------------------------

type MetricsDefFile struct {
	Metrics MetricsDef `yaml:"metrics"`
}

type MetricsSelFile struct {
	Metrics MetricsSelectRoot `yaml:"metrics"`
}

type MetricsDef struct {
	Dirs struct {
		Stats  string `yaml:"stats"`
		Output string `yaml:"output"`
		Times  string `yaml:"times"`
		Bugs   string `yaml:"bugs"`
	} `yaml:"dirs"`

	Files struct {
		StatsAll      string `yaml:"statsAll"`
		StatsAnalysis string `yaml:"statsAnalysis"`
		StatsFuzz     string `yaml:"statsFuzz"`
		StatsTrace    string `yaml:"statsTrace"`
		TimesTotal    string `yaml:"timesTotal"`
		TimesDetail   string `yaml:"timesDetail"`
		BugReports    string `yaml:"bugReports"`
	} `yaml:"files"`

	Derived map[string]DerivedRule `yaml:"derived"`
}

type MetricsSelectRoot struct {
	Default  Selection `yaml:"default"`
	Specific Selection `yaml:"specific"`
}

type Selection struct {
	Sources map[string]SourceSpec `yaml:"sources"`
	Derived []string              `yaml:"derived"`
}

type SourceSpec struct {
	Number []string          `yaml:"number"`
	String map[string]string `yaml:"string,omitempty"`
}

type DerivedRule struct {
	Op         string   `yaml:"op"`
	Source     string   `yaml:"source,omitempty"`
	Pattern    string   `yaml:"pattern,omitempty"`
	PresentIf  string   `yaml:"present_if,omitempty"`
	Join       string   `yaml:"join,omitempty"`
	FromMetric string   `yaml:"from_metric,omitempty"`
	Columns    []string `yaml:"columns,omitempty"`
	Column     string   `yaml:"column,omitempty"`
	Where      *Where   `yaml:"where,omitempty"`
	Patterns   []string `yaml:"patterns,omitempty"`
}

type Where struct {
	Column string `yaml:"column"`
	Equals string `yaml:"equals"`
}

// ------------------------
// Runtime plan derived from def+select
// ------------------------

type RuntimeConfig struct {
	Def       MetricsDef
	Selection Selection

	DerivedRequested map[string]struct{}
	OrderedKeys      []string
	OrderedDerived   []string
}

func loadRuntimeConfig(defPath, selectPath string) (*RuntimeConfig, error) {
	defPath = filepath.Clean(defPath)
	selectPath = filepath.Clean(selectPath)

	def, err := loadDef(defPath)
	if err != nil {
		return nil, fmt.Errorf("load metrics_def.yaml (%s): %w", defPath, err)
	}
	selRoot, err := loadSelect(selectPath)
	if err != nil {
		return nil, fmt.Errorf("load metrics_select.yaml (%s): %w", selectPath, err)
	}

	// Strict validation: def file must be complete
	if err := validateDef(def); err != nil {
		return nil, err
	}

	// Choose selection: specific overrides default if it contains anything
	choice := selRoot.Default
	if selectionNonEmpty(selRoot.Specific) {
		choice = selRoot.Specific
	}

	req := map[string]struct{}{}
	for _, d := range choice.Derived {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		req[d] = struct{}{}
	}

	ordered := make([]string, 0, len(choice.Derived))
	for _, d := range choice.Derived {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		ordered = append(ordered, d)
		req[d] = struct{}{}
	}

	return &RuntimeConfig{
		Def:              def,
		Selection:        choice,
		DerivedRequested: req,
		OrderedKeys:      buildOrderedKeys(choice),
		OrderedDerived:   ordered,
	}, nil
}

func loadDef(path string) (MetricsDef, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return MetricsDef{}, err
	}
	var f MetricsDefFile
	if err := yaml.Unmarshal(b, &f); err != nil {
		return MetricsDef{}, err
	}
	return f.Metrics, nil
}

func loadSelect(path string) (MetricsSelectRoot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return MetricsSelectRoot{}, err
	}
	var f MetricsSelFile
	if err := yaml.Unmarshal(b, &f); err != nil {
		return MetricsSelectRoot{}, err
	}
	return f.Metrics, nil
}

func selectionNonEmpty(s Selection) bool {
	if len(s.Derived) > 0 {
		return true
	}
	for _, spec := range s.Sources {
		if len(spec.Number) > 0 {
			return true
		}
	}
	return false
}

func validateDef(d MetricsDef) error {
	// dirs
	if strings.TrimSpace(d.Dirs.Stats) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.dirs.stats")
	}
	if strings.TrimSpace(d.Dirs.Output) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.dirs.output")
	}
	if strings.TrimSpace(d.Dirs.Times) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.dirs.times")
	}
	if strings.TrimSpace(d.Dirs.Bugs) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.dirs.bugs")
	}

	// files
	if strings.TrimSpace(d.Files.StatsAll) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.files.statsAll")
	}
	if strings.TrimSpace(d.Files.StatsAnalysis) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.files.statsAnalysis")
	}
	if strings.TrimSpace(d.Files.StatsFuzz) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.files.statsFuzz")
	}
	if strings.TrimSpace(d.Files.StatsTrace) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.files.statsTrace")
	}
	if strings.TrimSpace(d.Files.TimesTotal) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.files.timesTotal")
	}
	if strings.TrimSpace(d.Files.TimesDetail) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.files.timesDetail")
	}
	if strings.TrimSpace(d.Files.BugReports) == "" {
		return fmt.Errorf("metrics_def.yaml missing metrics.files.bugReports")
	}
	return nil
}

// Returns CSV header keys in the order defined by the active selection
func ActiveCSVHeader(includeTest, includeMode bool) ([]string, error) {
	cfg, err := getRuntimeConfig()
	if err != nil {
		return nil, err
	}

	header := make([]string, 0, 2+len(cfg.OrderedKeys))
	if includeTest {
		header = append(header, "Test")
	}
	if includeMode {
		header = append(header, "Mode")
	}
	header = append(header, cfg.OrderedKeys...)
	return header, nil
}

// Returns direct-source metrics first (in source priority), then derived as in YAML list order
func buildOrderedKeys(sel Selection) []string {
	keys := make([]string, 0)

	priority := []string{"statsAnalysis", "statsFuzz", "statsAll", "statsTrace", "timesTotal", "timesDetail"}
	seen := map[string]struct{}{}

	appendSource := func(src string) {
		spec, ok := sel.Sources[src]
		if !ok {
			return
		}
		for _, k := range spec.Number {
			k = strings.TrimSpace(k)
			if k == "" {
				continue
			}
			keys = append(keys, k)
		}
		seen[src] = struct{}{}
	}

	for _, src := range priority {
		appendSource(src)
	}

	// any extra sources appended stable
	extra := make([]string, 0)
	for src := range sel.Sources {
		if _, ok := seen[src]; ok {
			continue
		}
		extra = append(extra, src)
	}
	sort.Strings(extra)
	for _, src := range extra {
		appendSource(src)
	}

	// derived metrics preserve exact list order
	for _, d := range sel.Derived {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		keys = append(keys, d)
	}

	return keys
}
