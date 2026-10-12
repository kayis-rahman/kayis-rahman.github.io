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
		"index.html":          {"Kayis", "Synapse", "The Config That Finally Worked", "canonical", "Building automated trading systems and LLM inference infrastructure on the side."},
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

func TestPrototypeRejectsOutputPathsOverlappingSource(t *testing.T) {
	tmp := t.TempDir()
	binary := filepath.Join(tmp, "prototype-generator")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build generator: %v\n%s", err, output)
	}

	for i, outputPathKind := range []string{"root", "child", "parent-relative", "parent-traversal", "symlink-root", "symlink-child", "symlink-parent"} {
		base := filepath.Join(tmp, strings.ReplaceAll(outputPathKind, "/", "-"))
		root := filepath.Join(base, "source")
		workdir := filepath.Join(base, "runner")
		if err := os.MkdirAll(root, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(workdir, 0755); err != nil {
			t.Fatal(err)
		}
		marker := filepath.Join(root, "sentinel.txt")
		if err := os.WriteFile(marker, []byte("must remain"), 0644); err != nil {
			t.Fatal(err)
		}
		alias := filepath.Join(base, "source-alias")
		if err := os.Symlink(root, alias); err != nil {
			t.Fatal(err)
		}

		var output string
		switch outputPathKind {
		case "root":
			output = root
		case "child":
			output = filepath.Join(root, "generated")
		case "parent-relative":
			output = ".."
		case "parent-traversal":
			output = root + string(filepath.Separator) + ".."
		case "symlink-root":
			output = alias
		case "symlink-child":
			output = filepath.Join(alias, "generated")
		case "symlink-parent":
			output = alias + string(filepath.Separator) + ".."
		}
		cmd := exec.Command(binary, root)
		cmd.Dir = workdir
		cmd.Env = append(os.Environ(), "GO_PROTOTYPE_OUT="+output)
		if result, err := cmd.CombinedOutput(); err == nil {
			t.Errorf("case %d (%s): expected overlapping output path to fail, got %s", i, outputPathKind, result)
		}
		if contents, err := os.ReadFile(marker); err != nil || string(contents) != "must remain" {
			t.Errorf("case %d (%s): source sentinel changed or disappeared: contents=%q err=%v", i, outputPathKind, contents, err)
		}
	}
}
