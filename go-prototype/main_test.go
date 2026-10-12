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
		"index.html":          {"Kayis", "Synapse", "The Config That Finally Worked", "canonical", "Building automated trading systems and LLM inference infrastructure on the side.", "data-variant-choice=\"signal\"", "data-variant-choice=\"fieldnotes\"", "data-variant-choice=\"systems\""},
		"projects/index.html": {"Synapse", "Raspberry Pi 5", "Swing Trading System", "Pomodoro Timer iOS", "PKI Certificate Lifecycle Tool", "Nifty Screener", "canonical", "Visit my GitHub profile for project updates"},
		"posts/index.html":    {"Ideas to Production", "Why I Stopped Paying Per Token", "The Model Hunt", "The GPU Graveyard", "The Config That Finally Worked"},
		"career/engineering/2026/05/17/ideas-to-production.html":                                  {"Ten years", "MediSoft", "Qualcomm", "canonical"},
		"llm/infrastructure/self-hosting/2026/05/19/why-i-stopped-paying-per-token.html":          {"The Cost Maths", "Part 1", "canonical"},
		"llm/infrastructure/self-hosting/2026/05/26/the-model-hunt-five-attempts-one-winner.html": {"Attempt 1", "Part 2", "canonical"},
		"llm/infrastructure/self-hosting/2026/06/02/the-gpu-graveyard-cuinit-999.html":            {"cuInit: 999", "Part 3", "canonical"},
		"llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html":          {"CUDA 13", "vLLM", "canonical", "llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html", "SELF-HOSTING LLM INFERENCE", "Previous: The GPU Graveyard", "Is It Actually Worth It?"},
		"contact.html":       {"mailto:kaisbk1@gmail.com", "github.com/kayis-rahman", "linkedin.com/in/kayisrahman/", "canonical"},
		"contact/index.html": {"mailto:kaisbk1@gmail.com", "canonical", "https://kayis-rahman.github.io/contact"},
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
		if strings.Contains(string(data), `href="#"`) {
			t.Errorf("%s contains a dead fragment link", file)
		}
		for _, property := range []string{"og:title", "og:description", "og:url"} {
			if !strings.Contains(string(data), `property="`+property+`"`) {
				t.Errorf("%s missing %s metadata", file, property)
			}
		}
		if file == "index.html" && !strings.Contains(string(data), `alt=""`) {
			t.Errorf("home's decorative coding image should be ignored by assistive technology")
		}
		if file == "index.html" && strings.Count(string(data), "<h1>") != 1 {
			t.Errorf("home page should expose exactly one primary heading")
		}
	}
	css, err := os.ReadFile(filepath.Join(tmp, "prototype-site", "prototype.css"))
	if err != nil || !strings.Contains(string(css), "prefers-reduced-motion") {
		t.Errorf("generated styles are missing reduced-motion support: %v", err)
	}
	err = filepath.WalkDir(filepath.Join(tmp, "prototype-site"), func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && filepath.Ext(path) == ".js" {
			t.Errorf("unexpected separate JavaScript payload: %s", path)
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

func TestPrototypeFailsWhenSelectedProjectIsMissing(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "source")
	output := filepath.Join(tmp, "prototype-site")
	files := map[string]string{
		"index.markdown":  "---\ndescription: Prototype\n---\n",
		"about.markdown":  "---\ntitle: About\n---\nAbout\n",
		"resume.markdown": "---\ntitle: Resume\n---\nResume\n",
		"_posts/2026-06-09-the-config-that-finally-worked.md": "---\ntitle: Example\ndescription: Example\ndate: 2026-06-09\n---\nPost body\n",
		"projects.md": "---\ntitle: Projects\n---\nProjects intro\n\n---\n\n**Another project**\nAll projects must not be rendered as a fallback.\n",
	}
	for name, contents := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	asset := filepath.Join(root, "img", "home.jpg")
	if err := os.MkdirAll(filepath.Dir(asset), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(asset, []byte("image"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", ".", root)
	cmd.Env = append(os.Environ(), "GO_PROTOTYPE_OUT="+output)
	result, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected missing selected project to fail, got %s", result)
	}
	if !strings.Contains(string(result), "projects.md has no ## Synapse section") {
		t.Fatalf("expected explicit selected-project error, got %s", result)
	}
	if _, err := os.Stat(filepath.Join(output, "projects", "index.html")); !os.IsNotExist(err) {
		t.Fatalf("missing selected project unexpectedly produced a projects page: %v", err)
	}
}
