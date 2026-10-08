// This file applies the opted-in HTML transform recursively to generated pages.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func processDirectory(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk %s: %w", path, walkErr)
		}
		if entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 || !strings.EqualFold(filepath.Ext(path), ".html") {
			return nil
		}
		original, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		transformed, changed, err := transformHTML(original)
		if err != nil {
			return fmt.Errorf("transform %s: %w", path, err)
		}
		if changed {
			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf("inspect %s: %w", path, err)
			}
			if err := os.WriteFile(path, transformed, info.Mode().Perm()); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
		}
		return nil
	})
}
