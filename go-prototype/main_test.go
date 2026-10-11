package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrototypeBuildEmitsPagesAtExistingRoutes(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	cmd := exec.Command("go", "-C", filepath.Join(root, "go-prototype"), "run", ".", root)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(), "GO_PROTOTYPE_OUT="+filepath.Join(tmp, "prototype-site"))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generator failed: %v\n%s", err, output)
	}
	checks := map[string][]string{
		"index.html":          {"Kayis", "Synapse", "The Config That Finally Worked", "canonical"},
		"projects/index.html": {"Synapse", "Raspberry Pi 5", "canonical"},
		"llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html": {"CUDA 13", "vLLM", "canonical", "llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html"},
	}
	for file, needles := range checks {
		data, err := os.ReadFile(filepath.Join(tmp, "prototype-site", file))
		if err != nil {
			t.Fatalf("missing route %s: %v", file, err)
		}
		for _, needle := range needles {
			if !strings.Contains(string(data), needle) {
				t.Errorf("%s missing expected rendered content %q", file, needle)
			}
		}
		if file == "index.html" && strings.Count(string(data), "<h1>") != 1 {
			t.Errorf("home page should expose exactly one primary heading")
		}
	}
	err = filepath.WalkDir(filepath.Join(tmp, "prototype-site"), func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && filepath.Ext(path) == ".js" {
			t.Errorf("unexpected JavaScript payload: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
