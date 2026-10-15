package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var jsonLDScriptPattern = regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)
var localFragmentPattern = regexp.MustCompile(`href="#([^"]+)"`)
var elementIDPattern = regexp.MustCompile(`\bid="([^"]+)"`)

func parseJSONLDScripts(t *testing.T, page []byte) []map[string]any {
	t.Helper()
	matches := jsonLDScriptPattern.FindAllSubmatch(page, -1)
	schemas := make([]map[string]any, 0, len(matches))
	for _, match := range matches {
		var schema map[string]any
		if err := json.Unmarshal(match[1], &schema); err != nil {
			t.Fatalf("invalid JSON-LD script: %v", err)
		}
		schemas = append(schemas, schema)
	}
	return schemas
}

func assertFragmentLinksResolve(t *testing.T, page []byte) {
	t.Helper()
	ids := make(map[string]struct{})
	for _, match := range elementIDPattern.FindAllSubmatch(page, -1) {
		ids[string(match[1])] = struct{}{}
	}
	for _, match := range localFragmentPattern.FindAllSubmatch(page, -1) {
		if _, exists := ids[string(match[1])]; !exists {
			t.Errorf("fragment link #%s has no target", match[1])
		}
	}
}

func schemaOfType(schemas []map[string]any, schemaType string) map[string]any {
	for _, schema := range schemas {
		if schema["@type"] == schemaType {
			return schema
		}
	}
	return nil
}

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
		"index.html":          {"Kayis", "Synapse", "The Config That Finally Worked", "canonical", "Building automated trading systems and LLM inference infrastructure on the side.", "<html lang=\"en\" data-theme=\"light\"", `href="/search/">Search</a>`, `href="/feed.xml">RSS</a>`, "<body data-variant=\"systems\"", "data-variant-choice=\"signal\" aria-pressed=\"false\"", "data-variant-choice=\"fieldnotes\" aria-pressed=\"false\"", "data-variant-choice=\"systems\" aria-pressed=\"true\"", "Systems — a structured console aesthetic"},
		"projects/index.html": {"Synapse", "Raspberry Pi 5", "Swing Trading System", "Pomodoro Timer iOS", "PKI Certificate Lifecycle Tool", "Nifty Screener", "canonical", "Visit my GitHub profile for project updates"},
		"posts/index.html":    {"Ideas to Production", "Why I Stopped Paying Per Token", "The Model Hunt", "The GPU Graveyard", "The Config That Finally Worked"},
		"career/engineering/2026/05/17/ideas-to-production.html":                                  {"Ten years", "MediSoft", "Qualcomm", "canonical", `<p class="post-subtitle">A decade of building things, across five companies and two continents.</p>`, `href="/tags/#career"`, `href="/tags/#engineering"`, `Share on Twitter/X`, `Share on LinkedIn`, `About Kayis Rahman`, `https://buttondown.email/api/emails/embed-subscribe/kayis`, `href="/subscribe/"`},
		"llm/infrastructure/self-hosting/2026/05/19/why-i-stopped-paying-per-token.html":          {"The Cost Maths", "Part 1", "canonical"},
		"llm/infrastructure/self-hosting/2026/05/26/the-model-hunt-five-attempts-one-winner.html": {"Attempt 1", "Part 2", "canonical", `href="/llm/infrastructure/self-hosting/2026/05/19/why-i-stopped-paying-per-token.html" title="Why I Stopped Paying Per Token — Part 1 of 5">← Previous Post`, `href="/llm/infrastructure/self-hosting/2026/06/02/the-gpu-graveyard-cuinit-999.html" title="The GPU Graveyard: cuInit 999, Broken Hosts, and How to Verify an Instance — Part 3 of 5">Next Post →`, "Related Posts", `href="/llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html"`, `href="/llm/infrastructure/self-hosting/2026/06/02/the-gpu-graveyard-cuinit-999.html"`, `href="/llm/infrastructure/self-hosting/2026/05/19/why-i-stopped-paying-per-token.html"`},
		"llm/infrastructure/self-hosting/2026/06/02/the-gpu-graveyard-cuinit-999.html":            {"cuInit: 999", "Part 3", "canonical", `<div id="toc"></div><div class="post-body">`, `<h2 id="the-error-that-tells-you-nothing">The Error That Tells You Nothing</h2>`, `href="#the-error-that-tells-you-nothing"`},
		"llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html":          {"CUDA 13", "vLLM", "canonical", "llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html", "SELF-HOSTING LLM INFERENCE", "Previous: The GPU Graveyard", "Is It Actually Worth It?"},
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
		for _, destination := range []string{`href="/search/">Search</a>`, `href="/feed.xml">RSS</a>`} {
			if !strings.Contains(string(data), destination) {
				t.Errorf("%s is missing shared navigation destination %s", file, destination)
			}
		}
		if !strings.Contains(string(data), `<body data-variant="systems">`) {
			t.Errorf("%s does not use the site-wide Systems visual direction", file)
		}
		if !strings.Contains(string(data), `<button class="theme-toggle" type="button" data-theme-toggle aria-label="Dark theme" aria-pressed="false">`) {
			t.Errorf("%s is missing the accessible theme control", file)
		}
		for _, property := range []string{"og:title", "og:description", "og:url"} {
			if !strings.Contains(string(data), `property="`+property+`"`) {
				t.Errorf("%s missing %s metadata", file, property)
			}
		}
		for _, name := range []string{"twitter:title", "twitter:description", "twitter:image"} {
			if !strings.Contains(string(data), `name="`+name+`"`) {
				t.Errorf("%s missing %s metadata", file, name)
			}
		}
		for _, media := range []string{`media="(prefers-color-scheme: light)"`, `media="(prefers-color-scheme: dark)"`} {
			if !strings.Contains(string(data), media) {
				t.Errorf("%s missing browser theme color for %s", file, media)
			}
		}
		if file == "index.html" && !strings.Contains(string(data), `alt=""`) {
			t.Errorf("home's decorative coding image should be ignored by assistive technology")
		}
		if file == "index.html" && strings.Count(string(data), "<h1>") != 1 {
			t.Errorf("home page should expose exactly one primary heading")
		}
		if strings.Contains(file, "/2026/") {
			assertFragmentLinksResolve(t, data)
			if !strings.Contains(string(data), `<div id="toc"></div><div class="post-body">`) {
				t.Errorf("%s is missing the shared table-of-contents and post-body targets", file)
			}
		}
		if file == "llm/infrastructure/self-hosting/2026/05/26/the-model-hunt-five-attempts-one-winner.html" {
			if strings.Count(string(data), `class="related-post-card"`) != 3 {
				t.Errorf("related posts should include the three nearest same-tag posts")
			}
			relatedStart := strings.Index(string(data), `<div class="related-posts`)
			if relatedStart < 0 || strings.Contains(string(data)[relatedStart:], `href="/llm/infrastructure/self-hosting/2026/05/26/the-model-hunt-five-attempts-one-winner.html"`) {
				t.Errorf("related posts should exclude the current post")
			}
		}
	}
	for file, header := range map[string]string{
		"resume/index.html": "<th>Language</th>",
		"llm/infrastructure/self-hosting/2026/05/26/the-model-hunt-five-attempts-one-winner.html": "<th>Setup</th>",
	} {
		data, readErr := os.ReadFile(filepath.Join(tmp, "prototype-site", file))
		if readErr != nil {
			t.Fatalf("read generated table page %s: %v", file, readErr)
		}
		if !strings.Contains(string(data), `<div class="table-scroll" role="region" aria-label="Scrollable table" tabindex="0"><table>`) || !strings.Contains(string(data), header) {
			t.Errorf("%s does not expose its rendered table as a keyboard-scrollable region", file)
		}
	}
	homePage, err := os.ReadFile(filepath.Join(tmp, "prototype-site", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	homeSchemas := parseJSONLDScripts(t, homePage)
	person := schemaOfType(homeSchemas, "Person")
	if len(homeSchemas) != 1 || person == nil || person["url"] != siteURL || person["image"] != siteURL+"/assets/img/about-me.png" {
		t.Errorf("home JSON-LD does not preserve its Person schema: %#v", homeSchemas)
	}
	for file, image := range map[string]string{
		"index.html":          "/assets/img/home.jpg",
		"about/index.html":    "/assets/img/about-me.png",
		"projects/index.html": "/assets/img/projects.jpg",
		"resume/index.html":   "/assets/img/resume.jpg",
	} {
		data, readErr := os.ReadFile(filepath.Join(tmp, "prototype-site", file))
		if readErr != nil {
			t.Fatalf("read generated source page %s: %v", file, readErr)
		}
		wantImage := siteURL + image
		if !strings.Contains(string(data), `<meta property="og:image" content="`+wantImage+`">`) || !strings.Contains(string(data), `<meta name="twitter:image" content="`+wantImage+`">`) {
			t.Errorf("%s does not use its front-matter image for social metadata", file)
		}
	}
	for file, expected := range map[string]struct {
		image       string
		schemaImage string
	}{
		"career/engineering/2026/05/17/ideas-to-production.html":                                  {"/assets/img/posts/ideas-to-production.jpg", "/assets/img/posts/ideas-to-production.jpg"},
		"llm/infrastructure/self-hosting/2026/05/19/why-i-stopped-paying-per-token.html":          {"/assets/img/posts/gpu.jpg", "/assets/img/posts/gpu.jpg"},
		"llm/infrastructure/self-hosting/2026/05/26/the-model-hunt-five-attempts-one-winner.html": {"/assets/img/posts/model-hunt.jpg", "/assets/img/posts/model-hunt.jpg"},
		"llm/infrastructure/self-hosting/2026/06/02/the-gpu-graveyard-cuinit-999.html":            {"/assets/img/posts/gpu.jpg", "/assets/img/posts/gpu.jpg"},
		"llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html":          {"/img/home.jpg", ""},
	} {
		data, readErr := os.ReadFile(filepath.Join(tmp, "prototype-site", file))
		if readErr != nil {
			t.Fatalf("read generated post %s: %v", file, readErr)
		}
		wantImage := siteURL + expected.image
		if !strings.Contains(string(data), `<meta property="og:image" content="`+wantImage+`">`) || !strings.Contains(string(data), `<meta name="twitter:image" content="`+wantImage+`">`) {
			t.Errorf("%s does not use its front-matter image for social metadata", file)
		}
		schemas := parseJSONLDScripts(t, data)
		blogPosting := schemaOfType(schemas, "BlogPosting")
		breadcrumbList := schemaOfType(schemas, "BreadcrumbList")
		if len(schemas) != 2 || blogPosting == nil || breadcrumbList == nil {
			t.Errorf("%s must emit BlogPosting and BreadcrumbList schemas: %#v", file, schemas)
			continue
		}
		wantCanonical := siteURL + "/" + file
		if blogPosting["url"] != wantCanonical {
			t.Errorf("%s BlogPosting URL = %v, want %s", file, blogPosting["url"], wantCanonical)
		}
		author, authorOK := blogPosting["author"].(map[string]any)
		publisher, publisherOK := blogPosting["publisher"].(map[string]any)
		if !authorOK || author["@type"] != "Person" || author["name"] != "Kayis Rahman" || author["url"] != siteURL || !publisherOK || publisher["@type"] != "Organization" || publisher["name"] != "Kayis Rahman" || blogPosting["datePublished"] == nil || blogPosting["dateModified"] == nil {
			t.Errorf("%s BlogPosting lost existing author, publisher, or date fields: %#v", file, blogPosting)
		}
		if file == "llm/infrastructure/self-hosting/2026/05/26/the-model-hunt-five-attempts-one-winner.html" && (blogPosting["wordCount"] != float64(4500) || blogPosting["timeRequired"] != "PT12M") {
			t.Errorf("BlogPosting word count or reading time missing: %#v", blogPosting)
		}
		if expected.schemaImage == "" {
			if _, exists := blogPosting["image"]; exists {
				t.Errorf("%s BlogPosting unexpectedly has an image: %v", file, blogPosting["image"])
			}
		} else if blogPosting["image"] != siteURL+expected.schemaImage {
			t.Errorf("%s BlogPosting image = %v, want %s", file, blogPosting["image"], siteURL+expected.schemaImage)
		}
		items, ok := breadcrumbList["itemListElement"].([]any)
		if !ok || len(items) < 3 {
			t.Errorf("%s BreadcrumbList does not contain its ancestors: %#v", file, breadcrumbList)
			continue
		}
		homeItem, homeOK := items[0].(map[string]any)
		blogItem, blogOK := items[1].(map[string]any)
		categoryItem, categoryOK := items[2].(map[string]any)
		category := "llm"
		if strings.HasPrefix(file, "career/") {
			category = "career"
		}
		if !homeOK || !blogOK || !categoryOK || homeItem["name"] != "Home" || blogItem["name"] != "Blog" || categoryItem["item"] != siteURL+"/posts/#"+category {
			t.Errorf("%s BreadcrumbList does not preserve its Home, Blog, and category ancestors: %#v", file, breadcrumbList)
		}
	}
	for _, route := range []string{"contact.html", filepath.Join("contact", "index.html")} {
		if _, err := os.Stat(filepath.Join(tmp, "prototype-site", route)); !os.IsNotExist(err) {
			t.Errorf("Go generator unexpectedly emitted Jekyll-owned contact route %s: %v", route, err)
		}
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

func TestFrontMatterImageUsesImageBackgroundThenFallback(t *testing.T) {
	for _, testCase := range []struct {
		meta FrontMatter
		want string
	}{
		{FrontMatter{Image: "/posts/image.jpg", Background: "/posts/background.jpg"}, siteURL + "/posts/image.jpg"},
		{FrontMatter{Background: "/posts/background.jpg"}, siteURL + "/posts/background.jpg"},
		{FrontMatter{}, siteURL + "/img/home.jpg"},
	} {
		if got := frontMatterImageURL(testCase.meta); got != testCase.want {
			t.Errorf("frontMatterImageURL(%#v) = %q, want %q", testCase.meta, got, testCase.want)
		}
	}
}

func TestWriteRouteRejectsPostCategoryTraversal(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "output")
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	route, _, err := postRoute(FrontMatter{Date: "2026-01-01", Categories: []string{"../escape"}}, "2026-01-01-example.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeRoute(out, route, Page{Title: "Example"}); err == nil {
		t.Fatalf("writeRoute accepted traversal route %q", route)
	}
	if _, err := os.Stat(filepath.Join(root, "escape", "2026", "01", "01", "example.html")); !os.IsNotExist(err) {
		t.Fatalf("traversal route wrote outside output root: %v", err)
	}
}

func TestPostArticleKeepsSeriesNavigationWithinSeries(t *testing.T) {
	posts := []Post{
		{Meta: FrontMatter{Title: "Previous — Part 1", Series: "Series A", SeriesPart: 1}, URL: "/a/previous.html"},
		{Meta: FrontMatter{Title: "Current — Part 2", Series: "Series A", SeriesPart: 2}},
		{Meta: FrontMatter{Title: "Other Previous — Part 1", Series: "Series B", SeriesPart: 1}, URL: "/b/previous.html"},
		{Meta: FrontMatter{Title: "Next — Part 3", Series: "Series A", SeriesPart: 3}, URL: "/a/next.html"},
		{Meta: FrontMatter{Title: "Other Next — Part 3", Series: "Series B", SeriesPart: 3}, URL: "/b/next.html"},
	}
	article := string(postArticle(posts, 1))
	for _, expected := range []string{`href="/a/previous.html"`, `href="/a/next.html"`, "SERIES A · PART 2 OF 5"} {
		if !strings.Contains(article, expected) {
			t.Errorf("series article missing %q: %s", expected, article)
		}
	}
	seriesEnd := strings.Index(article, `</nav>`)
	if seriesEnd < 0 {
		t.Fatal("series navigation is not closed")
	}
	seriesArticle := article[:seriesEnd]
	for _, unexpected := range []string{`href="/b/previous.html"`, `href="/b/next.html"`} {
		if strings.Contains(seriesArticle, unexpected) {
			t.Errorf("series article includes cross-series destination %q: %s", unexpected, seriesArticle)
		}
	}
	unassigned := []Post{{Meta: FrontMatter{Title: "Unassigned", SeriesPart: 1}}}
	if article := string(postArticle(unassigned, 0)); strings.Contains(article, `class="series-nav"`) {
		t.Errorf("post without series identity has series navigation: %s", article)
	}
}

func TestLoadPostsExcludesUnpublishedAndFuturePosts(t *testing.T) {
	root := t.TempDir()
	postsDir := filepath.Join(root, "_posts")
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"2020-01-01-published.md":   "---\ntitle: Published\ndate: 2020-01-01\n---\nVisible\n",
		"2020-01-02-unpublished.md": "---\ntitle: Unpublished\ndate: 2020-01-02\npublished: false\n---\nHidden\n",
		"2020-01-03-future.md":      "---\ntitle: Future\ndate: " + time.Now().Add(24*time.Hour).Format("2006-01-02 15:04:05 -0700") + "\n---\nHidden\n",
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(postsDir, name), []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}

	posts, err := loadPosts(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0].Meta.Title != "Published" {
		t.Fatalf("loadPosts() = %#v, want only the eligible Published post", posts)
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
