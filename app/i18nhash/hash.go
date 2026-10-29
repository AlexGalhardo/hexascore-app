// Package i18nhash fingerprints the translation catalogues.
package i18nhash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
)

// Catalogs returns a short hash over every *.json file in fsys, in name order.
func Catalogs(fsys fs.FS) (string, error) {
	names, err := fs.Glob(fsys, "*.json")
	if err != nil {
		return "", fmt.Errorf("i18n hash: %w", err)
	}
	sort.Strings(names)

	h := sha256.New()
	for _, name := range names {
		b, err := fs.ReadFile(fsys, name)
		if err != nil {
			return "", fmt.Errorf("i18n hash: %s: %w", name, err)
		}
		_, _ = fmt.Fprintf(h, "%s\x00%d\x00", name, len(b))
		h.Write(b)
	}

	return hex.EncodeToString(h.Sum(nil))[:12], nil
}
