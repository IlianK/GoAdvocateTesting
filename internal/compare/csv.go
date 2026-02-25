package compare

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"GoAdvocateTesting/internal/metrics"
)

type Row struct {
	Fixed   map[string]string
	Numbers map[string]float64
	Strings map[string]string
}

// Writes CSV using provided header order (only for fixed keys)
// Remaining keys are appended in sorted order (fixed first, then numbers, then strings)
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
		for _, k := range header {
			if v, ok := r.Fixed[k]; ok {
				rec = append(rec, v)
				continue
			}
			if v, ok := r.Numbers[k]; ok {
				if v == float64(int64(v)) {
					rec = append(rec, strconv.FormatInt(int64(v), 10))
				} else {
					rec = append(rec, fmt.Sprintf("%g", v))
				}
				continue
			}
			if v, ok := r.Strings[k]; ok {
				rec = append(rec, v)
				continue
			}
			rec = append(rec, "")
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	return w.Error()
}

// Rows for cross/per compare with stable identity keys
func RowForCompareCSV(ms metrics.MetricSet, includeTest bool) Row {
	fixed := map[string]string{}
	if includeTest {
		if v, ok := ms.Strings["test_name"]; ok {
			fixed["Test"] = v
		}
	}
	if v, ok := ms.Strings["Mode"]; ok {
		fixed["Mode"] = v
	} else if v, ok := ms.Strings["mode"]; ok {
		fixed["Mode"] = v
	} else {
		fixed["Mode"] = ""
	}

	nums := map[string]float64{}
	for k, v := range ms.Numbers {
		nums[k] = v
	}

	strs := map[string]string{}
	for k, v := range ms.Strings {
		if k == "test_name" || k == "mode" || k == "Mode" || k == "kind" || k == "profile" || k == "label" || k == "run_id" {
			continue
		}
		strs[k] = v
	}

	return Row{Fixed: fixed, Numbers: nums, Strings: strs}
}
