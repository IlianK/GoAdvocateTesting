package metrics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"GoAdvocateTesting/internal/discovery"
)

type MetricSet struct {
	Numbers map[string]float64
	Strings map[string]string
}

// Reads metrics for a run using metrics_def.yaml + metrics_select.yaml
// missing files/columns just yield missing/zero values
func Extract(run discovery.Run) (MetricSet, error) {
	cfg, err := getRuntimeConfig()
	if err != nil {
		// If config can't be loaded, still return identity
		ms := MetricSet{Numbers: map[string]float64{}, Strings: map[string]string{}}
		fillIdentity(&ms, run)
		return ms, err
	}
	return extractWithCfg(run, cfg)
}

func extractWithCfg(run discovery.Run, cfg *RuntimeConfig) (MetricSet, error) {
	ms := MetricSet{
		Numbers: map[string]float64{},
		Strings: map[string]string{},
	}

	fillIdentity(&ms, run)

	// Cache parsed rows per source
	rowCache := map[string]map[string]string{}

	getRow := func(source string) map[string]string {
		if r, ok := rowCache[source]; ok {
			return r
		}
		r := loadSourceRow(run.Path, cfg.Def, source)
		rowCache[source] = r
		return r
	}

	// --------------------
	// 1) Direct sources (from metrics_select.yaml)
	// --------------------
	for sourceName, spec := range cfg.Selection.Sources {
		row := getRow(sourceName)
		for _, col := range spec.Number {
			col = strings.TrimSpace(col)
			if col == "" {
				continue
			}
			v := getFloat(row, col)
			if v != 0 {
				ms.Numbers[col] = v
			}
		}
	}

	// --------------------
	// 2) Derived metrics (only those listed in selection.derived)
	// --------------------
	for _, name := range cfg.OrderedDerived {
		rule, ok := cfg.Def.Derived[name]
		if !ok {
			continue
		}
		if err := applyDerived(&ms, run.Path, cfg.Def, getRow, name, rule); err != nil {
			return ms, err
		}
	}

	return ms, nil
}

func fillIdentity(ms *MetricSet, run discovery.Run) {
	ms.Strings["test_name"] = run.Test
	ms.Strings["kind"] = run.Kind
	ms.Strings["profile"] = run.Profile
	ms.Strings["run_id"] = run.RunID
	if run.Label != "" {
		ms.Strings["label"] = run.Label
	}
	if run.Mode != "" {
		ms.Strings["mode"] = run.Mode
	}
	ms.Numbers["exit_code"] = float64(run.ExitCode)
	if !run.StartedAt.IsZero() && !run.EndedAt.IsZero() {
		d := run.EndedAt.Sub(run.StartedAt)
		ms.Numbers["duration_ms"] = float64(d.Milliseconds())
	}
}

// --------------------
// Source loading
// --------------------

func loadSourceRow(runPath string, def MetricsDef, source string) map[string]string {
	switch source {
	case "statsAll", "statsAnalysis", "statsFuzz", "statsTrace":
		statsDir := filepath.Join(runPath, def.Dirs.Stats)
		pat := ""
		switch source {
		case "statsAll":
			pat = def.Files.StatsAll
		case "statsAnalysis":
			pat = def.Files.StatsAnalysis
		case "statsFuzz":
			pat = def.Files.StatsFuzz
		case "statsTrace":
			pat = def.Files.StatsTrace
		}
		return readFirstMatchingCSVRow(statsDir, pat)

	case "timesTotal":
		return map[string]string{}

	case "timesDetail":
		return map[string]string{}
	default:
		return map[string]string{}
	}
}

func readFirstMatchingCSVRow(dir, pattern string) map[string]string {
	out := map[string]string{}
	if pattern == "" {
		return out
	}
	matches, _ := filepath.Glob(filepath.Join(dir, pattern))
	for _, p := range matches {
		if row := parseCSVRowToMap(p); len(row) > 0 {
			return row
		}
	}
	return out
}

// parseCSVRowToMap reads header + first non-empty data row and returns map[header]=value.
func parseCSVRowToMap(path string) map[string]string {
	out := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		return out
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return out
	}

	var row []string
	for {
		rec, e := r.Read()
		if e != nil {
			return out
		}
		if len(rec) == 0 || strings.TrimSpace(strings.Join(rec, "")) == "" {
			continue
		}
		row = rec
		break
	}

	for i := 0; i < len(header) && i < len(row); i++ {
		h := strings.TrimSpace(header[i])
		v := strings.TrimSpace(row[i])
		if h == "" {
			continue
		}
		out[h] = v
	}
	return out
}

func getFloat(m map[string]string, key string) float64 {
	raw, ok := m[key]
	if !ok {
		return 0
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0
	}
	return f
}

func parseFloatLoose(s string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return f
}

// --------------------
// Derived operations
// --------------------

func applyDerived(ms *MetricSet, runPath string, def MetricsDef, getRow func(source string) map[string]string, name string, rule DerivedRule) error {
	switch strings.ToLower(strings.TrimSpace(rule.Op)) {
	case "codes_present":
		// Produce string of codes where columns matching regex are >0
		if rule.Source == "" || rule.Pattern == "" {
			return nil
		}
		row := getRow(rule.Source)
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("derived %s: bad regex: %w", name, err)
		}
		codes := map[string]struct{}{}
		for k, v := range row {
			m := re.FindStringSubmatch(k)
			if len(m) != 2 {
				continue
			}
			val := parseFloatLoose(v)
			if comparePresent(val, rule.PresentIf) {
				codes[m[1]] = struct{}{}
			}
		}
		list := make([]string, 0, len(codes))
		for c := range codes {
			list = append(list, c)
		}
		sort.Strings(list)
		sep := rule.Join
		if sep == "" {
			sep = ";"
		}
		ms.Strings[name] = strings.Join(list, sep)
		return nil

	case "count_codes":
		// Count number of codes in a previously computed string metric
		src := rule.FromMetric
		if src == "" {
			return nil
		}
		s := strings.TrimSpace(ms.Strings[src])
		if s == "" {
			ms.Numbers[name] = 0
			return nil
		}
		parts := strings.Split(s, ";")
		n := 0
		for _, p := range parts {
			if strings.TrimSpace(p) != "" {
				n++
			}
		}
		ms.Numbers[name] = float64(n)
		return nil

	case "count_files":
		// Count file matches. Patterns may include placeholders
		if len(rule.Patterns) == 0 {
			return nil
		}
		total := 0
		for _, pat := range rule.Patterns {
			pat = expandPlaceholders(pat, runPath, def)
			matches, _ := filepath.Glob(pat)
			total += len(matches)
		}
		ms.Numbers[name] = float64(total)
		return nil

	case "sum_columns":
		// Sum columns from a source row
		if rule.Source == "" || len(rule.Columns) == 0 {
			return nil
		}
		row := getRow(rule.Source)
		sum := 0.0
		for _, c := range rule.Columns {
			sum += getFloat(row, c)
		}
		ms.Numbers[name] = sum
		return nil

	case "alias":
		// Copy one column from a source under a new name
		if rule.Source == "" || rule.Column == "" {
			return nil
		}
		row := getRow(rule.Source)
		v := getFloat(row, rule.Column)
		if v != 0 {
			ms.Numbers[name] = v
		}
		return nil

	case "regex_sum":
		// Sum all columns matching regex in a source
		if rule.Source == "" || rule.Pattern == "" {
			return nil
		}
		row := getRow(rule.Source)
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("derived %s: bad regex: %w", name, err)
		}
		sum := 0.0
		for k, v := range row {
			if re.MatchString(k) {
				sum += parseFloatLoose(v)
			}
		}
		ms.Numbers[name] = sum
		return nil

	case "times_total":
		// Read times_total CSV (run root or run/times) and apply optional where clause
		col := rule.Column
		if col == "" {
			col = "Time"
		}
		whereCol := "TestName"
		whereEq := "Total"
		if rule.Where != nil {
			if rule.Where.Column != "" {
				whereCol = rule.Where.Column
			}
			if rule.Where.Equals != "" {
				whereEq = rule.Where.Equals
			}
		}

		v := readTimesTotal(runPath, def, whereCol, whereEq, col)
		if v != 0 {
			ms.Numbers[name] = v
		}
		return nil

	case "times_detail":
		// Read times_detail CSV (run root or run/times) and take first data row's column
		if rule.Column == "" {
			return nil
		}
		v := readTimesDetail(runPath, def, rule.Column)
		if v != 0 {
			ms.Numbers[name] = v
		}
		return nil

	default:
		// Unknown op: ignore
		return nil
	}
}

func comparePresent(val float64, expr string) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return val > 0
	}
	switch expr {
	case ">0":
		return val > 0
	case ">=1":
		return val >= 1
	case "!=0":
		return val != 0
	default:
		// fallback
		return val > 0
	}
}

func expandPlaceholders(pat string, runPath string, def MetricsDef) string {
	out := pat
	out = strings.ReplaceAll(out, "{run}", runPath)
	out = strings.ReplaceAll(out, "{bugReports}", def.Files.BugReports)
	out = strings.ReplaceAll(out, "{bugsDir}", def.Dirs.Bugs)
	out = strings.ReplaceAll(out, "//", "/")
	return out
}

// --------------------
// times parsing
// --------------------

func readTimesTotal(runPath string, def MetricsDef, whereCol, whereEq, valueCol string) float64 {
	cands := []string{
		filepath.Join(runPath, def.Files.TimesTotal),
		filepath.Join(runPath, def.Dirs.Times, def.Files.TimesTotal),
	}
	for _, pat := range cands {
		matches, _ := filepath.Glob(pat)
		for _, p := range matches {
			if v := readTimesTotalCSV(p, whereCol, whereEq, valueCol); v != 0 {
				return v
			}
		}
	}
	return 0
}

func readTimesTotalCSV(path, whereCol, whereEq, valueCol string) float64 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return 0
	}

	idx := func(name string) int {
		for i, h := range header {
			if strings.TrimSpace(h) == name {
				return i
			}
		}
		return -1
	}

	iWhere := idx(whereCol)
	iVal := idx(valueCol)
	if iWhere < 0 || iVal < 0 {
		return 0
	}

	for {
		rec, e := r.Read()
		if e != nil {
			return 0
		}
		if len(rec) <= iVal || len(rec) <= iWhere {
			continue
		}
		if strings.TrimSpace(rec[iWhere]) == whereEq {
			return parseFloatLoose(rec[iVal])
		}
	}
}

func readTimesDetail(runPath string, def MetricsDef, col string) float64 {
	cands := []string{
		filepath.Join(runPath, def.Files.TimesDetail),
		filepath.Join(runPath, def.Dirs.Times, def.Files.TimesDetail),
	}
	for _, pat := range cands {
		matches, _ := filepath.Glob(pat)
		for _, p := range matches {
			if v := readTimesDetailCSV(p, col); v != 0 {
				return v
			}
		}
	}
	return 0
}

func readTimesDetailCSV(path string, col string) float64 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return 0
	}

	idx := func(name string) int {
		for i, h := range header {
			if strings.TrimSpace(h) == name {
				return i
			}
		}
		return -1
	}

	iCol := idx(col)
	if iCol < 0 {
		return 0
	}

	for {
		rec, e := r.Read()
		if e != nil {
			return 0
		}
		if len(rec) <= iCol {
			continue
		}
		if strings.TrimSpace(strings.Join(rec, "")) == "" {
			continue
		}
		return parseFloatLoose(rec[iCol])
	}
}

// ---- inputs.json helpers ----

type InputRunRecord struct {
	Path    string          `json:"path"`
	Meta    json.RawMessage `json:"meta"`
	Metrics map[string]any  `json:"metrics"`
}

type InputsFile struct {
	GeneratedAt time.Time        `json:"generatedAt"`
	Kind        string           `json:"kind"`
	Profile     string           `json:"profile"`
	Label       string           `json:"label"`
	Metric      string           `json:"metric,omitempty"`
	Runs        []InputRunRecord `json:"runs"`
}

func BuildInputsFile(kind, profile, label, metric string, runs []discovery.Run, sets []MetricSet) InputsFile {
	out := InputsFile{
		GeneratedAt: time.Now().UTC(),
		Kind:        kind,
		Profile:     profile,
		Label:       label,
		Metric:      metric,
		Runs:        make([]InputRunRecord, 0, len(runs)),
	}

	for i := range runs {
		metaPath := filepath.Join(runs[i].Path, "meta.json")
		raw, _ := os.ReadFile(metaPath)

		rec := InputRunRecord{
			Path:    metaPath,
			Meta:    raw,
			Metrics: map[string]any{},
		}

		for k, v := range sets[i].Numbers {
			rec.Metrics[k] = v
		}
		for k, v := range sets[i].Strings {
			rec.Metrics[k] = v
		}

		out.Runs = append(out.Runs, rec)
	}
	return out
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func WriteJSON(path string, v any) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func MergeMetricKeys(sets []MetricSet) (numKeys []string, strKeys []string) {
	numSet := map[string]struct{}{}
	strSet := map[string]struct{}{}
	for _, ms := range sets {
		for k := range ms.Numbers {
			numSet[k] = struct{}{}
		}
		for k := range ms.Strings {
			strSet[k] = struct{}{}
		}
	}
	for k := range numSet {
		numKeys = append(numKeys, k)
	}
	for k := range strSet {
		strKeys = append(strKeys, k)
	}
	sort.Strings(numKeys)
	sort.Strings(strKeys)
	return numKeys, strKeys
}
