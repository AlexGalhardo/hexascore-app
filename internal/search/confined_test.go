package search_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// FR-07-1: search SQL lives in internal/search; migrations define the objects and are the only other owner.
var searchSQL = regexp.MustCompile(`search_tsv|search_suggest|search_dirty|\b(title|name|label)_norm\b|\bhs_norm\b|<%`)

func TestSearchSQLConfined(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel := filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator)))

		if d.IsDir() {
			switch rel {
			case ".git", "node_modules", "internal/search", "database/migrations", "vendor", ".orchestration":
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(rel, ".go") {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if m := searchSQL.Find(src); m != nil {
			t.Errorf("%s references search SQL (%q) outside internal/search", rel, m)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
