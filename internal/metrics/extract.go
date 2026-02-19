package metrics

import (
	"encoding/csv"
	"encoding/json"
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

func Extract(run discovery.Run) (MetricSet, error) {
	ms := MetricSet{
		Numbers: map[string]float64{},
		Strings: map[string]string{},
	}

	// ---- identity ----
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

	// ---- Total_Bugs (bugs_*.md in run root; fallback to <run>/bugs/) ----
	totalBugs := 0
	if n, _ := countMatchingFiles(run.Path, func(name string) bool {
		return strings.HasPrefix(name, "bugs_") && strings.HasSuffix(name, ".md")
	}); n > 0 {
		totalBugs = n
	} else {
		if n2, _ := countMatchingFiles(filepath.Join(run.Path, "bugs"), func(name string) bool {
			return strings.HasPrefix(name, "bugs_") && strings.HasSuffix(name, ".md")
		}); n2 > 0 {
			totalBugs = n2
		}
	}
	ms.Numbers["Total_Bugs"] = float64(totalBugs)

	// ---- stats (statsAll, statsAnalysis, statsFuzz) ----
	var (
		uniqueByCode   = map[string]float64{}
		replayWritten  float64
		replaySuccess  float64
		panicsTotal    float64
		leaksTotal     float64
		replayConfirms float64
		totalRuns      float64
	)

	statsDir := filepath.Join(run.Path, "stats")
	entries, err := os.ReadDir(statsDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			full := filepath.Join(statsDir, name)

			switch {
			case strings.HasPrefix(name, "statsAnalysis_"):
				row := parseCSVRowToMap(full)
				panicsTotal = getFloat(row, "NrPanicsTotal")
				leaksTotal = getFloat(row, "NrLeaksTotal")
				replayConfirms = getFloat(row, "NrPanicsVerifiedViaReplayTotal") + getFloat(row, "NrLeaksResolvedViaReplayTotal")

			case strings.HasPrefix(name, "statsFuzz_"):
				row := parseCSVRowToMap(full)
				totalRuns = getFloat(row, "NrMut")

			case strings.HasPrefix(name, "statsAll_"):
				u, rw, rs := parseStatsAllUniqueAndReplay(full)
				for k, v := range u {
					uniqueByCode[k] += v
				}
				replayWritten += rw
				replaySuccess += rs
			}
		}
	}

	ms.Numbers["Panics"] = panicsTotal
	ms.Numbers["Leaks"] = leaksTotal
	ms.Numbers["Confirmed_Replays"] = replayConfirms
	ms.Numbers["Total_Runs"] = totalRuns
	ms.Numbers["Replays_Written"] = replayWritten
	ms.Numbers["Replays_Successful"] = replaySuccess

	// ---- Unique_Bugs + Bug_Types (from statsAll_* first) ----
	var (
		uniqueSum float64
		types     []string
	)
	for code, v := range uniqueByCode {
		if v > 0 {
			types = append(types, code)
			uniqueSum += v
		}
	}
	sort.Strings(types)
	ms.Numbers["Unique_Bugs"] = uniqueSum
	ms.Strings["Bug_Types"] = strings.Join(types, ";")

	// ---- Bug_Types fallback: output/results_machine.log ----
	// (This is the authoritative detected-bug listing; one line per bug type)
	if strings.TrimSpace(ms.Strings["Bug_Types"]) == "" {
		fb := readBugTypesFromResultsMachine(run.Path)
		if len(fb) > 0 {
			ms.Strings["Bug_Types"] = strings.Join(fb, ";")
			// if unique count is missing/zero but we have types, use count(types) as fallback
			if ms.Numbers["Unique_Bugs"] == 0 {
				ms.Numbers["Unique_Bugs"] = float64(len(fb))
			}
		}
	}

	// ---- Mode (as user-facing column) ----
	// For analysis runs this can be empty.
	ms.Strings["Mode"] = run.Mode

	// ---- Total_Time_s (times_total_*.csv either in run root or run/times/) ----
	ms.Numbers["Total_Time_s"] = readTotalTimeSeconds(run.Path)

	// ---- Rec_s / Ana_s / Rep_s (prefer dedicated files; fallback to times_detail_*.csv) ----
	rec, ana, rep := readPhaseSeconds(run.Path)
	if rec > 0 {
		ms.Numbers["Rec_s"] = rec
	}
	if ana > 0 {
		ms.Numbers["Ana_s"] = ana
	}
	if rep > 0 {
		ms.Numbers["Rep_s"] = rep
	}

	return ms, nil
}

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

// parseStatsAllUniqueAndReplay extracts:
// - NoUniqueetected<CODE> (and tolerate NoUniqueetected typo variants)
// - NoTotaleplayWritten<CODE> and NoTotaleplaySuccessful<CODE> (tolerate typo variants)
func parseStatsAllUniqueAndReplay(path string) (unique map[string]float64, replayWritten float64, replaySuccess float64) {
	unique = map[string]float64{}

	row := parseCSVRowToMap(path)

	// ADVOCATE typos: "NoUniqueetected"
	reUniq := regexp.MustCompile(`^NoUnique(?:e)?tected([A-Z]\d\d|P\d\d|L\d\d)$`)

	// Replay columns: "NoTotaleplayWritten<CODE>" and "...Successful<CODE>"
	reRW := regexp.MustCompile(`^NoTotal(?:e)?playWritten([A-Z]\d\d|P\d\d|L\d\d)$`)
	reRS := regexp.MustCompile(`^NoTotal(?:e)?playSuccessful([A-Z]\d\d|P\d\d|L\d\d)$`)

	for k, v := range row {
		if m := reUniq.FindStringSubmatch(k); len(m) == 2 {
			unique[m[1]] = parseFloatLoose(v)
			continue
		}
		if m := reRW.FindStringSubmatch(k); len(m) == 2 {
			replayWritten += parseFloatLoose(v)
			continue
		}
		if m := reRS.FindStringSubmatch(k); len(m) == 2 {
			replaySuccess += parseFloatLoose(v)
			continue
		}
	}

	return unique, replayWritten, replaySuccess
}

func readBugTypesFromResultsMachine(runPath string) []string {
	p := filepath.Join(runPath, "output", "results_machine.log")
	b, err := os.ReadFile(p)
	if err != nil {
		return nil
	}

	set := map[string]struct{}{}
	lines := strings.Split(string(b), "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		// Example: "P06,tp,T:3:..."
		// Take the token before first ',' (fallback: before whitespace)
		code := ln
		if i := strings.Index(code, ","); i >= 0 {
			code = code[:i]
		} else if i := strings.IndexAny(code, " \t"); i >= 0 {
			code = code[:i]
		}
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		set[code] = struct{}{}
	}

	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func parseFloatLoose(s string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return f
}

func readTotalTimeSeconds(runPath string) float64 {
	cands := []string{
		filepath.Join(runPath, "times_total_*.csv"),
		filepath.Join(runPath, "times", "times_total_*.csv"),
	}
	for _, pat := range cands {
		matches, _ := filepath.Glob(pat)
		for _, p := range matches {
			if v := readTimeTotalCSV(p); v > 0 {
				return v
			}
		}
	}
	return 0
}

func readTimeTotalCSV(path string) float64 {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	_, err = r.Read() // header
	if err != nil {
		return 0
	}
	for {
		rec, e := r.Read()
		if e != nil {
			return 0
		}
		if len(rec) < 2 {
			continue
		}
		name := strings.TrimSpace(rec[0])
		val := strings.TrimSpace(rec[1])
		if name == "Total" {
			return parseFloatLoose(val)
		}
	}
}

func readPhaseSeconds(runPath string) (rec, ana, rep float64) {
	rec = readScalarFromAny(runPath, []string{"rec_s", "rec_s.txt", "rec_s.csv"})
	ana = readScalarFromAny(runPath, []string{"ana_s", "ana_s.txt", "ana_s.csv"})
	rep = readScalarFromAny(runPath, []string{"rep_s", "rep_s.txt", "rep_s.csv"})

	if rec == 0 && ana == 0 && rep == 0 {
		cands := []string{
			filepath.Join(runPath, "times_detail_*.csv"),
			filepath.Join(runPath, "times", "times_detail_*.csv"),
		}
		for _, pat := range cands {
			matches, _ := filepath.Glob(pat)
			for _, p := range matches {
				r2, a2, p2 := readTimeDetailCSV(p)
				if r2 > 0 || a2 > 0 || p2 > 0 {
					return r2, a2, p2
				}
			}
		}
	}

	return rec, ana, rep
}

func readScalarFromAny(dir string, names []string) float64 {
	for _, n := range names {
		p := filepath.Join(dir, n)
		if _, err := os.Stat(p); err == nil {
			if v := readScalarFile(p); v > 0 {
				return v
			}
		}
	}
	return 0
}

func readScalarFile(path string) float64 {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	s := strings.TrimSpace(string(b))
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		s = strings.TrimSpace(parts[len(parts)-1])
	}
	return parseFloatLoose(s)
}

func readTimeDetailCSV(path string) (rec, ana, rep float64) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, 0
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return 0, 0, 0
	}

	var row []string
	for {
		recRow, e := r.Read()
		if e != nil {
			return 0, 0, 0
		}
		if len(recRow) == 0 || strings.TrimSpace(strings.Join(recRow, "")) == "" {
			continue
		}
		row = recRow
		break
	}

	idx := func(name string) int {
		for i, h := range header {
			if strings.TrimSpace(h) == name {
				return i
			}
		}
		return -1
	}

	if i := idx("Recording"); i >= 0 && i < len(row) {
		rec = parseFloatLoose(row[i])
	}
	if i := idx("Analysis"); i >= 0 && i < len(row) {
		ana = parseFloatLoose(row[i])
	}
	if i := idx("Replay"); i >= 0 && i < len(row) {
		rep = parseFloatLoose(row[i])
	}
	return rec, ana, rep
}

func countMatchingFiles(dir string, match func(name string) bool) (int, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if match(e.Name()) {
			n++
		}
	}
	return n, nil
}

// ---- inputs.json helpers (unchanged) ----

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
