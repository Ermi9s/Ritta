package env

import (
	"os"
	"path/filepath"
	"strings"
)

func Scan(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkipDirectory(path, root) {
				return filepath.SkipDir
			}

			return nil
		}

		name := info.Name()

		if strings.HasPrefix(name, ".env") {
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

func shouldSkipDirectory(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	if relative == "." {
		return false
	}

	first := strings.Split(relative, string(os.PathSeparator))[0]

	switch first {
	case ".git", "node_modules", "vendor", "target", "dist", "build":
		return true
	}

	return false
}