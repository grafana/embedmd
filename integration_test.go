package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIntegration(t *testing.T) {
	const base = "testData"
	dir, err := os.ReadDir(base)
	assert.NoError(t, err)
	for _, d := range dir {
		name := d.Name()
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command("../../embedmd/embedmd",
				"-m",
				"lgtm=https://raw.githubusercontent.com/grafana/docker-otel-lgtm/73272e8995e9c5460d543d0b909317d5877c3855",
				"docs.md")
			cmd.Dir = filepath.Join(base, name)
			got, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("could not process file (%v): %s", err, got)
			}
			wants, err := os.ReadFile(filepath.Join(base, name, "result.md"))
			if err != nil {
				t.Fatalf("could not read result: %v", err)
			}
			assert.Equal(t, string(wants), string(got))
		})
	}
}
