package env

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)


var skipDirs = map[string]bool{
	".git":         true,
	".hg":          true,
	".svn":         true,
	"node_modules": true,
	"vendor":       true,
	"target":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
	".nuxt":        true,
	".turbo":       true,
	".cache":       true,
	".terraform":   true,
	".idea":        true,
	".vscode":      true,
	"__pycache__":  true,
	".venv":        true,
	"venv":         true,
	"env":          true, 
	".mypy_cache":  true,
	".pytest_cache": true,
	".tox":         true,
}


func Scan(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			return err
		}

		if d.IsDir() {
			if path != root && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if isEnvFile(d.Name()) {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, relative)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}


func isEnvFile(name string) bool {
	if name == ".env" {
		return true
	}
	return strings.HasPrefix(name, ".env.")
}