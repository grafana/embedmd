package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIntegration(t *testing.T) {
	cmd := exec.Command("../embedmd/embedmd", "docs.md")
	cmd.Dir = "sample"
	got, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("could not process file (%v): %s", err, got)
	}
	wants, err := ioutil.ReadFile(filepath.Join("sample", "result.md"))
	if err != nil {
		t.Fatalf("could not read result: %v", err)
	}
	assert.Equal(t, string(wants), string(got))
}
