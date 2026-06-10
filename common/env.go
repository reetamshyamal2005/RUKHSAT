package common

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	loadLocalEnv()
}

// loadLocalEnv hydrates missing process env vars from a local .env-style file.
// Existing env vars always win so hosted environments like Vercel are unaffected.
func loadLocalEnv() {
	for _, path := range envCandidatePaths(".env.local") {
		if loadEnvFile(path) {
			return
		}
	}
}

func envCandidatePaths(filename string) []string {
	seen := map[string]struct{}{}
	var paths []string

	addPath := func(path string) {
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if _, exists := seen[clean]; exists {
			return
		}
		seen[clean] = struct{}{}
		paths = append(paths, clean)
	}

	if cwd, err := os.Getwd(); err == nil {
		for dir := cwd; ; dir = filepath.Dir(dir) {
			addPath(filepath.Join(dir, filename))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}

	if exePath, err := os.Executable(); err == nil {
		for dir := filepath.Dir(exePath); ; dir = filepath.Dir(dir) {
			addPath(filepath.Join(dir, filename))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}

	return paths
}

func loadEnvFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		_ = os.Setenv(key, value)
	}

	return true
}
