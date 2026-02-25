package discovery

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"GoAdvocateTesting/internal/app"
)

// DiscoverTests finds TestXxx in *_test.go files directly under testDir (non-recursive)
func DiscoverTests(testDir string) ([]app.TestCase, error) {
	return discoverTestsImpl(filepath.Clean(testDir), filepath.Clean(testDir), false)
}

// DiscoverTestsRecursive finds TestXxx in *_test.go files under root recursively
// returns TestCases where TestCase.File is the relative path from root to the *_test.go file.
// Example: "10214/cockroach10214_test.go"
func DiscoverTestsRecursive(root string) ([]app.TestCase, error) {
	return discoverTestsImpl(filepath.Clean(root), filepath.Clean(root), true)
}

func discoverTestsImpl(root string, current string, recursive bool) ([]app.TestCase, error) {
	var goFiles []string

	skipDir := func(name string) bool {
		switch name {
		case ".git", "vendor", "node_modules", "results", "comparisons":
			return true
		}
		return strings.HasPrefix(name, ".")
	}

	if !recursive {
		entries, err := os.ReadDir(current)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if strings.HasSuffix(e.Name(), "_test.go") {
				goFiles = append(goFiles, filepath.Join(current, e.Name()))
			}
		}
	} else {
		err := filepath.WalkDir(current, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				base := filepath.Base(path)
				if path != current && skipDir(base) {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(d.Name(), "_test.go") {
				goFiles = append(goFiles, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Strings(goFiles)

	fset := token.NewFileSet()
	found := map[string]app.TestCase{} // key: relFile + "::" + testName (prevents collisions)

	for _, fpath := range goFiles {
		f, err := parser.ParseFile(fset, fpath, nil, 0)
		if err != nil {
			continue
		}

		rel := filepath.Base(fpath)
		if recursive {
			if r, err := filepath.Rel(root, fpath); err == nil {
				rel = filepath.Clean(r)
			}
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name == nil {
				continue
			}
			name := fn.Name.Name
			if !strings.HasPrefix(name, "Test") {
				continue
			}
			// Basic signature check: one param (testing.T-like)
			if fn.Type == nil || fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
				continue
			}

			key := rel + "::" + name
			found[key] = app.TestCase{
				Name: name,
				File: rel,
			}
		}
	}

	out := make([]app.TestCase, 0, len(found))
	for _, tc := range found {
		out = append(out, tc)
	}

	// Stable sort: by file then name
	sort.Slice(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Name < out[j].Name
	})

	return out, nil
}
