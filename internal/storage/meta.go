package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"GoAdvocateTesting/internal/app"
)

func WriteMetaJSON(runDir string, meta app.RunMeta) error {
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(runDir, "meta.json"), b, 0o644)
}
