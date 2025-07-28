//go:build unit
// +build unit

package config

import (
	"os"
	"testing"
)

const testYAML = `
repositories:
  - https://github.com/example/testrepo1.git
  - https://github.com/example/testrepo2.git
`

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/depsync.yaml"
	if err := os.WriteFile(file, []byte(testYAML), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(file)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Repositories) != 2 {
		t.Errorf("expected 2 repositories, got %d", len(cfg.Repositories))
	}
	if cfg.Repositories[0] != "https://github.com/example/testrepo1.git" || cfg.Repositories[1] != "https://github.com/example/testrepo2.git" {
		t.Errorf("unexpected repository URLs: %+v", cfg.Repositories)
	}
}
