package main

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"
)

type FrontMatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Date        string   `yaml:"date"`
	Categories  []string `yaml:"categories"`
	Tags        []string `yaml:"tags"`
	Series      string   `yaml:"series_title"`
	SeriesPart  int      `yaml:"series_part"`
}
type Page struct {
	Title, Description, Canonical, Date string
	HTML                                template.HTML
	HideTitle                           bool
}

var frontMatter = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n(.*)$`)
var liquidRaw = regexp.MustCompile(`(?s)\{%\s*(?:raw|endraw)\s*%\}`)
var markdown = goldmark.New()
var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>{{.Title}} · Kayis Rahman</title><meta name="description" content="{{.Description}}"><link rel="canonical" href="{{.Canonical}}"><link rel="stylesheet" href="/prototype.css"></head><body>
<header class="site-header"><a class="wordmark" href="/">KR<span>.</span></a><nav aria-label="Primary"><a href="/">Home</a><a href="/projects/">Projects</a><a href="/llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html">Writing</a><a href="https://github.com/kayisrahman">GitHub ↗</a></nav></header>
<main>{{if not .HideTitle}}{{if .Date}}<p class="eyebrow">FIELD NOTES <span>·</span> {{.Date}}</p>{{end}}<h1>{{.Title}}</h1><p class="lede">{{.Description}}</p>{{end}}<article>{{.HTML}}</article></main><footer><span>Kayis Rahman</span><a href="/">Back to index ↑</a></footer></body></html>`))

func renderMarkdown(source string) (template.HTML, error) {
	var out bytes.Buffer
	if err := markdown.Convert([]byte(liquidRaw.ReplaceAllString(source, "")), &out); err != nil {
		return "", err
	}
	return template.HTML(out.String()), nil
}
func readMarkdown(path string) (FrontMatter, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FrontMatter{}, "", err
	}
	parts := frontMatter.FindSubmatch(data)
	if len(parts) != 3 {
		return FrontMatter{}, "", fmt.Errorf("%s has no YAML front matter", path)
	}
	var meta FrontMatter
	if err := yaml.Unmarshal(parts[1], &meta); err != nil {
		return meta, "", fmt.Errorf("%s front matter: %w", path, err)
	}
	return meta, string(parts[2]), nil
}

// resolvePath follows existing symlinks one component at a time without cleaning away
// later .. segments first (which would change their meaning after a symlink).
func resolvePath(path string) (string, error) {
	current, remainder := string(filepath.Separator), path
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		current, err = filepath.EvalSymlinks(cwd)
		if err != nil {
			return "", err
		}
	} else {
		remainder = strings.TrimPrefix(path, string(filepath.Separator))
	}

	for _, component := range strings.Split(remainder, string(filepath.Separator)) {
		switch component {
		case "", ".":
			continue
		case "..":
			current = filepath.Dir(current)
		default:
			candidate := filepath.Join(current, component)
			info, err := os.Lstat(candidate)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return "", err
			}
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				current, err = filepath.EvalSymlinks(candidate)
				if err != nil {
					return "", err
				}
			} else {
				current = candidate
			}
		}
	}
	return filepath.Clean(current), nil
}

func pathContains(parent, candidate string) bool {
	rel, err := filepath.Rel(parent, candidate)
	return err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)))
}

func safeOutputPath(sourceRoot, outputPath string) (string, error) {
	root, err := resolvePath(sourceRoot)
	if err != nil {
		return "", fmt.Errorf("resolve source root: %w", err)
	}
	output, err := resolvePath(outputPath)
	if err != nil {
		return "", fmt.Errorf("resolve output path: %w", err)
	}
	if pathContains(root, output) || pathContains(output, root) {
		return "", fmt.Errorf("output path %q overlaps source root %q", output, root)
	}
	return output, nil
}

func writePage(outPath string, page Page) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return pageTemplate.Execute(f, page)
}
func main() {
	root := ".."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	var err error
	root, err = resolvePath(root)
	if err != nil {
		fatal(fmt.Errorf("resolve source root: %w", err))
	}
	out := os.Getenv("GO_PROTOTYPE_OUT")
	if out == "" {
		out = filepath.Join(os.TempDir(), "kayis-go-prototype-site")
	}
	out, err = safeOutputPath(root, out)
	if err != nil {
		fatal(err)
	}
	if err := os.RemoveAll(out); err != nil {
		fatal(err)
	}
	copyAsset := func(from, to string) {
		data, err := os.ReadFile(filepath.Join(root, from))
		if err != nil {
			fatal(err)
		}
		target := filepath.Join(out, to)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			fatal(err)
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			fatal(err)
		}
	}
	copyAsset("img/home.jpg", "img/home.jpg")
	if err := os.WriteFile(filepath.Join(out, "prototype.css"), []byte(css), 0644); err != nil {
		fatal(err)
	}

	homeMeta, _, err := readMarkdown(filepath.Join(root, "index.markdown"))
	if err != nil {
		fatal(err)
	}
	postMeta, postText, err := readMarkdown(filepath.Join(root, "_posts/2026-06-09-the-config-that-finally-worked.md"))
	if err != nil {
		fatal(err)
	}
	postHTML, err := renderMarkdown(postText)
	if err != nil {
		fatal(err)
	}
	postDate, err := time.Parse("2006-01-02", postMeta.Date)
	if err != nil {
		fatal(err)
	}
	postRoute := "/llm/infrastructure/self-hosting/" + postDate.Format("2006/01/02") + "/the-config-that-finally-worked.html"
	if err := writePage(filepath.Join(out, filepath.FromSlash(strings.TrimPrefix(postRoute, "/"))), Page{Title: postMeta.Title, Description: postMeta.Description, Canonical: "https://kayis-rahman.github.io" + postRoute, Date: postDate.Format("02 Jan 2006"), HTML: postHTML}); err != nil {
		fatal(err)
	}

	projects, err := os.ReadFile(filepath.Join(root, "projects.md"))
	if err != nil {
		fatal(err)
	}
	projectSource := string(projects)
	i := strings.Index(projectSource, "**Synapse**")
	if i < 0 {
		fatal(errors.New("projects.md has no **Synapse** section"))
	}
	projectSource = projectSource[i:]
	if j := strings.Index(projectSource, "\n---\n"); j >= 0 {
		projectSource = projectSource[:j]
	}
	projectHTML, err := renderMarkdown(projectSource)
	if err != nil {
		fatal(err)
	}
	if err := writePage(filepath.Join(out, "projects", "index.html"), Page{Title: "Projects · Synapse", Description: "Selected project from the existing projects.md source.", Canonical: "https://kayis-rahman.github.io/projects/", HTML: projectHTML}); err != nil {
		fatal(err)
	}

	homeBody := `<section class="hero"><p class="eyebrow">STAFF SECURITY ENGINEER <span>·</span> CORK, IRELAND</p><h1>Kayis<br><em>Rahman.</em></h1><p class="lede">` + html.EscapeString(homeMeta.Description) + `</p><div class="actions"><a class="button" href="/projects/">Explore my work <span>↗</span></a><a class="text-link" href="/llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html">Read the latest <span>↗</span></a></div></section><section class="feature"><div><p class="eyebrow">SELECTED PROJECT <span>·</span> IN PROGRESS</p><h2>Synapse</h2><p>A self-hosted AI proxy routing inference between remote GPUs and lightweight local models.</p><a href="/projects/">View project details ↗</a></div><img src="/img/home.jpg" alt=""></section><section class="latest"><p class="eyebrow">LATEST FIELD NOTE</p><h2><a href="/llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html">The Config That Finally Worked</a></h2><p>Every vLLM flag, patch, and template fix that made the stack work.</p></section>`
	if err := writePage(filepath.Join(out, "index.html"), Page{Title: "Kayis Rahman — Security Engineering & Systems", Description: homeMeta.Description, Canonical: "https://kayis-rahman.github.io/", HTML: template.HTML(homeBody), HideTitle: true}); err != nil {
		fatal(err)
	}
	fmt.Printf("Built home, /projects/, and %s into %s\n", postRoute, out)
}
func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }

const css = `:root{color-scheme:light;--ink:#111;--paper:#fff;--muted:#676767;--blue:#0099ff;--line:#dedede}*{box-sizing:border-box}body{margin:0;background:var(--paper);color:var(--ink);font:16px/1.7 Inter,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}.site-header,footer{max-width:1120px;margin:auto;padding:28px 36px;display:flex;align-items:center;justify-content:space-between;border-bottom:1px solid var(--line)}.wordmark{font-weight:900;font-size:24px;letter-spacing:-2px;color:var(--ink);text-decoration:none}.wordmark span{color:var(--blue)}nav{display:flex;gap:28px}a{color:inherit;text-decoration:none}a:hover{color:var(--blue)}main{max-width:1120px;margin:auto;padding:100px 36px 110px}h1{font-size:clamp(54px,9vw,112px);line-height:.98;letter-spacing:-.075em;margin:14px 0 26px;font-weight:850}h1 em{font-style:normal;color:var(--blue)}h2{font-size:42px;letter-spacing:-.05em;line-height:1.1;margin:8px 0 16px}.eyebrow{font-size:11px;letter-spacing:.16em;font-weight:750;color:var(--muted)}.eyebrow span{color:var(--blue)}.lede{font-size:20px;line-height:1.55;max-width:660px;color:#555}.actions{display:flex;align-items:center;gap:28px;margin-top:36px}.button{display:inline-flex;gap:22px;align-items:center;background:var(--blue);color:white;padding:14px 19px;font-size:14px;font-weight:700}.button:hover{color:white;opacity:.85}.text-link,article a,.feature a{color:var(--blue);font-weight:650}.feature{display:grid;grid-template-columns:1fr 1fr;align-items:center;gap:52px;border-top:1px solid var(--line);padding:46px 0;margin-top:110px}.feature img{width:100%;height:260px;object-fit:cover}.feature p:not(.eyebrow),.latest p{color:var(--muted);max-width:550px}.latest{border-top:1px solid var(--line);padding:40px 0}.latest h2{font-size:34px}.latest h2 a:hover{color:var(--blue)}article{max-width:760px;margin-top:52px;font-size:18px;line-height:1.8}article h2{font-size:32px;margin-top:48px}article h3{margin-top:36px}article pre{overflow:auto;padding:22px;background:#f4f5f6;border:1px solid var(--line);font-size:13px;line-height:1.55}article code{font-family:ui-monospace,SFMono-Regular,Menlo,monospace}article blockquote{border-left:3px solid var(--blue);padding-left:20px;color:var(--muted)}footer{border-top:1px solid var(--line);border-bottom:0;color:var(--muted);font-size:13px}@media(max-width:650px){.site-header,footer{padding:20px}nav{gap:12px;font-size:12px}main{padding:70px 22px}.feature{grid-template-columns:1fr;gap:18px;margin-top:75px}.feature img{height:190px}.actions{gap:18px;flex-wrap:wrap}h2{font-size:34px}}`
