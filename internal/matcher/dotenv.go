package matcher

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadDotEnv tries to load a .env file from the working directory or the executable directory.
// It also checks parent directories up to the project root (go.mod, wails.json, or .git).
// Existing env vars are not overwritten.
func LoadDotEnv() {
	paths := candidateEnvPaths()
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			_ = loadDotEnvFile(p)
			return
		}
	}
}

func candidateEnvPaths() []string {
	paths := []string{}
	seen := map[string]bool{}
	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		paths = append(paths, p)
	}
	if wd, err := os.Getwd(); err == nil {
		for _, p := range envPathsUp(wd) {
			add(p)
		}
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, p := range envPathsUp(exeDir) {
			add(p)
		}
	}
	return paths
}

func envPathsUp(start string) []string {
	paths := []string{}
	dir := start
	for {
		paths = append(paths, filepath.Join(dir, ".env"))
		if hasProjectMarker(dir) {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return paths
}

func hasProjectMarker(dir string) bool {
	markers := []string{"go.mod", "wails.json", ".git"}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
			return true
		}
	}
	return false
}

func loadDotEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := parseEnvLine(line)
		if !ok {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

func parseEnvLine(line string) (string, string, bool) {
	idx := strings.Index(line, "=")
	if idx <= 0 {
		return "", "", false
	}
	key := strings.TrimSpace(line[:idx])
	val := strings.TrimSpace(line[idx+1:])
	if key == "" {
		return "", "", false
	}
	val = strings.Trim(val, "\"'")
	return key, val, true
}
