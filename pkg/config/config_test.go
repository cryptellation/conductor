//go:build unit
// +build unit

package config

import (
	"os"
	"testing"
)

const testYAML = `
repositories:
  - name: testrepo1
    url: https://github.com/example/testrepo1.git
  - name: testrepo2
    url: https://github.com/example/testrepo2.git
`

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/conductor.yaml"
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
	if cfg.Repositories[0].Name != "testrepo1" || cfg.Repositories[1].Name != "testrepo2" {
		t.Errorf("unexpected repository names: %+v", cfg.Repositories)
	}
}
