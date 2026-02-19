package compare

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type Row struct {
	Fixed   map[string]string  // identity columns
	Numbers map[string]float64 // metric columns
	Strings map[string]string  // string metrics
}

func (r Row) Get(key string) (string, bool) {
	if r.Fixed != nil {
		if v, ok := r.Fixed[key]; ok {
			return v, true
		}
	}
	if r.Strings != nil {
		if v, ok := r.Strings[key]; ok {
			return v, true
		}
	}
	if r.Numbers != nil {
		if v, ok := r.Numbers[key]; ok {
			return formatNumber(v), true
		}
	}
	return "", false
}

func formatNumber(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return fmt.Sprintf("%g", v)
}

// WriteCSV keeps the old behavior: fixed columns first (with fixedOrder preference),
// then numeric metrics sorted, then string metrics sorted.
func WriteCSV(path string, rows []Row, fixedOrder []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	fixedKeysSet := map[string]struct{}{}
	numKeysSet := map[string]struct{}{}
	strKeysSet := map[string]struct{}{}

	for _, r := range rows {
		for k := range r.Fixed {
			fixedKeysSet[k] = struct{}{}
		}
		for k := range r.Numbers {
			numKeysSet[k] = struct{}{}
		}
		for k := range r.Strings {
			strKeysSet[k] = struct{}{}
		}
	}

	fixed := make([]string, 0, len(fixedKeysSet))
	for k := range fixedKeysSet {
		fixed = append(fixed, k)
	}
	sort.Strings(fixed)

	headerFixed := make([]string, 0, len(fixed))
	used := map[string]bool{}
	for _, k := range fixedOrder {
		if _, ok := fixedKeysSet[k]; ok {
			headerFixed = append(headerFixed, k)
			used[k] = true
		}
	}
	for _, k := range fixed {
		if !used[k] {
			headerFixed = append(headerFixed, k)
		}
	}

	numKeys := make([]string, 0, len(numKeysSet))
	for k := range numKeysSet {
		numKeys = append(numKeys, k)
	}
	sort.Strings(numKeys)

	strKeys := make([]string, 0, len(strKeysSet))
	for k := range strKeysSet {
		strKeys = append(strKeys, k)
	}
	sort.Strings(strKeys)

	header := append([]string{}, headerFixed...)
	header = append(header, numKeys...)
	header = append(header, strKeys...)

	return WriteCSVOrdered(path, rows, header)
}

// WriteCSVOrdered writes exactly the given header order (mixing numeric + string freely).
func WriteCSVOrdered(path string, rows []Row, header []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(header); err != nil {
		return err
	}

	for _, r := range rows {
		rec := make([]string, 0, len(header))
		for _, col := range header {
			if v, ok := r.Get(col); ok {
				rec = append(rec, v)
			} else {
				rec = append(rec, "")
			}
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}

	return w.Error()
}
