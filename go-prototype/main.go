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
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"gopkg.in/yaml.v3"
)

type FrontMatter struct {
	Title        string   `yaml:"title"`
	Description  string   `yaml:"description"`
	Date         string   `yaml:"date"`
	Categories   []string `yaml:"categories"`
	Tags         []string `yaml:"tags"`
	Series       string   `yaml:"series_title"`
	LegacySeries string   `yaml:"series"`
	Published    *bool    `yaml:"published"`
	SeriesPart   int      `yaml:"series_part"`
}
type Page struct {
	Title, Description, Canonical, Date string
	OGType, OGImage                     string
	Variant                             string
	HTML                                template.HTML
	HideTitle                           bool
}
type Post struct {
	Meta FrontMatter
	HTML template.HTML
	Date time.Time
	URL  string
}

var frontMatter = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n(.*)$`)
var liquidRaw = regexp.MustCompile(`(?s)\{%\s*(?:raw|endraw)\s*%\}`)
var markdownTable = regexp.MustCompile(`(?s)<table(?:\s[^>]*)?>.*?</table>`)
var markdown = goldmark.New(goldmark.WithExtensions(extension.Table))
var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en" data-theme="light"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><meta name="description" content="{{.Description}}"><title>{{.Title}} · Kayis Rahman</title><link rel="canonical" href="{{.Canonical}}"><meta property="og:type" content="{{.OGType}}"><meta property="og:title" content="{{.Title}} · Kayis Rahman"><meta property="og:description" content="{{.Description}}"><meta property="og:url" content="{{.Canonical}}"><meta property="og:image" content="{{.OGImage}}"><meta name="twitter:card" content="summary_large_image"><script>(function(){var root=document.documentElement,key='theme',stored=null,media=null;try{stored=window.localStorage.getItem(key)}catch(e){}try{media=window.matchMedia?window.matchMedia('(prefers-color-scheme: dark)'):null}catch(e){}var valid=stored==='light'||stored==='dark',state={theme:valid?stored:(media&&media.matches?'dark':'light'),hasStoredChoice:valid};root.setAttribute('data-theme',state.theme);window.__kayisThemeState=state;if(media){var onChange=function(event){if(state.hasStoredChoice)return;state.theme=event.matches?'dark':'light';if(state.applyTheme)state.applyTheme(state.theme);else root.setAttribute('data-theme',state.theme)};if(media.addEventListener)media.addEventListener('change',onChange);else if(media.addListener)media.addListener(onChange)}})();</script><link rel="stylesheet" href="/prototype.css"></head><body data-variant="{{.Variant}}">
<a class="skip-link" href="#main-content">Skip to main content</a><header class="site-header"><a class="wordmark" href="/" aria-label="Kayis Rahman home">KR<span>.</span></a><nav aria-label="Primary"><a href="/about/">About</a><a href="/posts/">Writing</a><a href="/projects/">Projects</a><a href="/resume/">Resume</a><a href="/contact">Contact</a><a href="/search/">Search</a><a href="/feed.xml">RSS</a></nav><button class="theme-toggle" type="button" data-theme-toggle aria-label="Dark theme" aria-pressed="false"><span aria-hidden="true">◐</span> Dark mode</button></header>
<main id="main-content">{{if not .HideTitle}}{{if .Date}}<p class="eyebrow">FIELD NOTES <span>·</span> {{.Date}}</p>{{end}}<h1>{{.Title}}</h1>{{if .Description}}<p class="lede">{{.Description}}</p>{{end}}{{end}}<article>{{.HTML}}</article></main><footer><span>Kayis Rahman · Cork, Ireland</span><nav aria-label="Social and site links"><a href="mailto:kaisbk1@gmail.com">Email</a><a href="https://github.com/kayis-rahman" rel="noopener noreferrer">GitHub ↗</a><a href="https://www.linkedin.com/in/kayisrahman/" rel="noopener noreferrer">LinkedIn ↗</a></nav></footer><script>(function(){var root=document.documentElement,button=document.querySelector('[data-theme-toggle]'),state=window.__kayisThemeState||{theme:root.getAttribute('data-theme')||'light',hasStoredChoice:false};if(!button)return;function update(){button.setAttribute('aria-pressed',String(root.getAttribute('data-theme')==='dark'))}function apply(theme){root.setAttribute('data-theme',theme);state.theme=theme;update()}state.applyTheme=apply;apply(root.getAttribute('data-theme')||state.theme||'light');button.addEventListener('click',function(){var next=root.getAttribute('data-theme')==='dark'?'light':'dark';state.hasStoredChoice=true;apply(next);try{window.localStorage.setItem('theme',next)}catch(e){}})})();</script></body></html>`))

func renderMarkdown(source string) (template.HTML, error) {
	var out bytes.Buffer
	if err := markdown.Convert([]byte(liquidRaw.ReplaceAllString(source, "")), &out); err != nil {
		return "", err
	}
	rendered := markdownTable.ReplaceAllStringFunc(out.String(), func(table string) string {
		return `<div class="table-scroll" role="region" aria-label="Scrollable table" tabindex="0">` + table + `</div>`
	})
	return template.HTML(rendered), nil
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
func writeRoute(out, route string, page Page) error {
	if page.Variant == "" {
		page.Variant = "systems"
	}
	name := strings.TrimPrefix(route, "/")
	if strings.HasSuffix(route, "/") {
		name += "index.html"
	}
	root, err := resolvePath(out)
	if err != nil {
		return fmt.Errorf("resolve output root: %w", err)
	}
	target, err := resolvePath(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		return fmt.Errorf("resolve route %q: %w", route, err)
	}
	if !pathContains(root, target) || target == root {
		return fmt.Errorf("route %q escapes output root %q", route, root)
	}
	return writePage(target, page)
}
func parsePostDate(dateText string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02 15:04:05 -0700", time.RFC3339, "2006-01-02"} {
		if date, err := time.Parse(layout, dateText); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid post date %q", dateText)
}
func postRoute(meta FrontMatter, filename string) (string, time.Time, error) {
	date, err := parsePostDate(meta.Date)
	if err != nil {
		return "", time.Time{}, err
	}
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	base = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`).ReplaceAllString(base, "")
	parts := append([]string{}, meta.Categories...)
	parts = append(parts, date.Format("2006/01/02"), base+".html")
	return "/" + strings.Join(parts, "/"), date, nil
}
func loadPosts(root string) ([]Post, error) {
	files, err := filepath.Glob(filepath.Join(root, "_posts", "*.md"))
	if err != nil {
		return nil, err
	}
	other, err := filepath.Glob(filepath.Join(root, "_posts", "*.markdown"))
	if err != nil {
		return nil, err
	}
	files = append(files, other...)
	posts := make([]Post, 0, len(files))
	for _, file := range files {
		meta, body, err := readMarkdown(file)
		if err != nil {
			return nil, err
		}
		if meta.Date == "" || meta.Title == "" || (meta.Published != nil && !*meta.Published) {
			continue
		}
		route, date, err := postRoute(meta, file)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		if date.After(time.Now()) {
			continue
		}
		htmlBody, err := renderMarkdown(body)
		if err != nil {
			return nil, err
		}
		posts = append(posts, Post{Meta: meta, HTML: htmlBody, Date: date, URL: route})
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date.After(posts[j].Date) })
	return posts, nil
}
func postArticle(posts []Post, i int) template.HTML {
	body := string(posts[i].HTML)
	current := posts[i].Meta
	if current.SeriesPart == 0 || current.LegacySeries == "" {
		return template.HTML(body)
	}
	series := `<nav class="series-nav" aria-label="Series navigation"><p class="eyebrow">` + html.EscapeString(strings.ToUpper(current.LegacySeries)) + ` · PART ` + fmt.Sprint(current.SeriesPart) + ` OF 5</p>`
	for _, post := range posts {
		part := post.Meta.SeriesPart
		if post.Meta.LegacySeries != current.LegacySeries || part == 0 || part == current.SeriesPart {
			continue
		}
		if part == current.SeriesPart-1 {
			label := strings.Split(post.Meta.Title, " — ")[0]
			series += `<a href="` + html.EscapeString(post.URL) + `">← Previous: ` + html.EscapeString(label) + `</a>`
		}
		if part == current.SeriesPart+1 {
			label := strings.Split(post.Meta.Title, " — ")[0]
			series += `<a href="` + html.EscapeString(post.URL) + `">Next: ` + html.EscapeString(label) + ` →</a>`
		}
	}
	if current.LegacySeries == "Self-Hosting LLM Inference on Rented GPUs" && current.SeriesPart == 4 {
		series += `<p>Part 5 — Is It Actually Worth It? <span class="coming-soon">Coming soon</span></p>`
	}
	series += `</nav>`
	return template.HTML(body + series)
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
	if err := os.MkdirAll(out, 0755); err != nil {
		fatal(err)
	}
	copyAsset := func(from, to string) {
		data, e := os.ReadFile(filepath.Join(root, from))
		if e != nil {
			fatal(e)
		}
		target := filepath.Join(out, to)
		if e = os.MkdirAll(filepath.Dir(target), 0755); e != nil {
			fatal(e)
		}
		if e = os.WriteFile(target, data, 0644); e != nil {
			fatal(e)
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
	posts, err := loadPosts(root)
	if err != nil {
		fatal(err)
	}
	if len(posts) == 0 {
		fatal(errors.New("no dated writing posts found"))
	}
	for i, post := range posts {
		if err := writeRoute(out, post.URL, Page{Title: post.Meta.Title, Description: post.Meta.Description, Canonical: "https://kayis-rahman.github.io" + post.URL, Date: post.Date.Format("02 Jan 2006"), OGType: "article", OGImage: "https://kayis-rahman.github.io/img/home.jpg", HTML: postArticle(posts, i)}); err != nil {
			fatal(err)
		}
	}

	_, projectText, err := readMarkdown(filepath.Join(root, "projects.md"))
	if err != nil {
		fatal(err)
	}
	if !strings.Contains(projectText, "## Synapse") {
		fatal(errors.New("projects.md has no ## Synapse section"))
	}
	projectHTML, err := renderMarkdown(projectText)
	if err != nil {
		fatal(err)
	}
	if err := writeRoute(out, "/projects/", Page{Title: "Projects", Description: "A selection of independent engineering projects in AI infrastructure, security, and thoughtful software.", Canonical: "https://kayis-rahman.github.io/projects/", OGType: "website", OGImage: "https://kayis-rahman.github.io/img/home.jpg", HTML: projectHTML}); err != nil {
		fatal(err)
	}

	writeSourcePage := func(file, route string) {
		meta, text, e := readMarkdown(filepath.Join(root, file))
		if e != nil {
			fatal(e)
		}
		body, e := renderMarkdown(text)
		if e != nil {
			fatal(e)
		}
		if e = writeRoute(out, route, Page{Title: meta.Title, Description: meta.Description, Canonical: "https://kayis-rahman.github.io" + route, OGType: "website", OGImage: "https://kayis-rahman.github.io/img/home.jpg", HTML: body}); e != nil {
			fatal(e)
		}
	}
	writeSourcePage("about.markdown", "/about/")
	writeSourcePage("resume.markdown", "/resume/")

	var listing strings.Builder
	for _, post := range posts {
		listing.WriteString(`<section class="post-card"><p class="eyebrow">` + post.Date.Format("02 Jan 2006") + ` · ` + strings.Join(post.Meta.Categories, " / ") + `</p><h2><a href="` + html.EscapeString(post.URL) + `">` + html.EscapeString(post.Meta.Title) + `</a></h2><p>` + html.EscapeString(post.Meta.Description) + `</p></section>`)
	}
	if err := writeRoute(out, "/posts/", Page{Title: "Writing", Description: "Writing on security engineering, PKI, Java, and self-hosted LLM inference.", Canonical: "https://kayis-rahman.github.io/posts/", OGType: "website", OGImage: "https://kayis-rahman.github.io/img/home.jpg", HTML: template.HTML(listing.String())}); err != nil {
		fatal(err)
	}

	latest := posts[0]
	homeBody := `<section class="hero"><p class="eyebrow">STAFF SECURITY ENGINEER <span>·</span> CORK, IRELAND</p><h1>Kayis<br><em>Rahman.</em></h1><p class="lede">` + html.EscapeString(homeMeta.Description) + `</p><div class="actions"><a class="button" href="/projects/">Explore my work <span>↗</span></a><a class="text-link" href="` + html.EscapeString(latest.URL) + `">Read the latest <span>↗</span></a></div><p class="focus-note">PKI · HSM · secure provisioning · systems that hold</p></section><section class="feature"><div><p class="eyebrow">SELECTED PROJECT · IN PROGRESS</p><h2>Synapse</h2><p>A self-hosted AI proxy routing inference between remote GPUs and lightweight local models.</p><a href="/projects/">View project details ↗</a></div><img src="/img/home.jpg" alt="" width="1200" height="700"></section><section class="latest"><p class="eyebrow">LATEST FIELD NOTE</p><h2><a href="` + html.EscapeString(latest.URL) + `">` + html.EscapeString(latest.Meta.Title) + `</a></h2><p>` + html.EscapeString(latest.Meta.Description) + `</p><a href="/posts/">Browse all writing ↗</a></section><section class="variant-picker" aria-labelledby="variant-title"><p class="eyebrow">LANDING DIRECTIONS · PROTOTYPE</p><h2 id="variant-title">Choose a visual direction</h2><p>Each direction uses the same content and routes; the presentation changes.</p><div class="variant-controls" role="group" aria-label="Landing page direction"><button type="button" data-variant-choice="signal" aria-pressed="false">01 · Signal</button><button type="button" data-variant-choice="fieldnotes" aria-pressed="false">02 · Field notes</button><button type="button" data-variant-choice="systems" aria-pressed="true">03 · Systems</button></div><p class="variant-description" aria-live="polite">Systems — a structured console aesthetic with project-first navigation.</p></section><section class="contact-strip"><p>Building careful systems. Curious about the hard edges.</p><a href="/contact">Start a conversation ↗</a></section><script>(function(){var body=document.body,buttons=document.querySelectorAll('[data-variant-choice]'),copy={signal:'Signal — high-contrast editorial, a clear first action, and one vivid technical accent.',fieldnotes:'Field notes — warm paper, compact typography, and writing-led discovery.',systems:'Systems — a structured console aesthetic with project-first navigation.'};buttons.forEach(function(button){button.addEventListener('click',function(){var value=button.getAttribute('data-variant-choice');body.setAttribute('data-variant',value);buttons.forEach(function(item){item.setAttribute('aria-pressed',String(item===button));});document.querySelector('.variant-description').textContent=copy[value];});});})();</script>`
	if err := writeRoute(out, "/", Page{Title: "Kayis Rahman — Security Engineering & Systems", Description: homeMeta.Description, Canonical: "https://kayis-rahman.github.io/", OGType: "website", OGImage: "https://kayis-rahman.github.io/img/home.jpg", Variant: "systems", HTML: template.HTML(homeBody), HideTitle: true}); err != nil {
		fatal(err)
	}
	fmt.Printf("Built home, about, resume, projects, posts index, and %d writing posts into %s\n", len(posts), out)
}
func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }

const css = `:root{color-scheme:light;--ink:#17191c;--paper:#fff;--muted:#5c6670;--blue:#005a9c;--line:#d4dbe0;--surface:#f2f4f6;--ease-out:cubic-bezier(.23,1,.32,1)}*{box-sizing:border-box}html{scroll-behavior:smooth}body{margin:0;background:var(--paper);color:var(--ink);font:16px/1.7 Georgia,"Times New Roman",serif;transition:background-color .2s ease,color .2s ease}.skip-link{position:absolute;left:12px;top:-60px;background:var(--ink);color:var(--paper);padding:10px;z-index:5}.skip-link:focus{top:12px}.site-header,footer{max-width:1160px;margin:auto;padding:25px 38px;display:flex;align-items:center;justify-content:space-between;border-bottom:1px solid var(--line)}.wordmark{font:900 24px/1 system-ui,sans-serif;letter-spacing:-2px;color:var(--ink);text-decoration:none}.wordmark span,h1 em{color:var(--blue)}nav{display:flex;gap:24px}.theme-toggle{border:1px solid var(--line);background:var(--surface);color:var(--ink);padding:9px 11px;font:700 12px/1 system-ui,sans-serif;white-space:nowrap;cursor:pointer}.theme-toggle span{color:var(--blue);margin-right:6px}.theme-toggle:hover{border-color:var(--blue)}html[data-theme=dark]{color-scheme:dark;--ink:#f2f5f7;--paper:#11161a;--muted:#bcc7cf;--blue:#87c8ff;--line:#3c4750;--surface:#1b2329}a{color:inherit;text-decoration:none}a:hover{color:var(--blue)}a:focus-visible,button:focus-visible{outline:3px solid var(--blue);outline-offset:4px}.site-header nav,footer nav{font:600 13px/1.3 system-ui,sans-serif}.site-header nav a,footer nav a{padding:8px 0}main{max-width:1160px;margin:auto;padding:86px 38px 96px}h1,h2{font-family:system-ui,-apple-system,sans-serif}h1{font-size:clamp(58px,10vw,118px);line-height:.91;letter-spacing:-.085em;margin:15px 0 24px;font-weight:850}h1 em{font-style:normal}.eyebrow{font:700 10px/1.5 system-ui,sans-serif;letter-spacing:.17em;color:var(--muted);text-transform:uppercase}.eyebrow span{color:var(--blue)}.lede{font-size:clamp(19px,2vw,23px);line-height:1.55;max-width:670px;color:var(--muted)}.actions{display:flex;align-items:center;gap:27px;margin-top:32px;flex-wrap:wrap}.button{display:inline-flex;gap:22px;align-items:center;background:var(--blue);color:var(--paper);padding:14px 19px;font:700 14px system-ui,sans-serif}.button:hover{color:var(--paper);opacity:.84}.text-link,article a,.feature a,.latest>a,.contact-strip a{color:var(--blue);font:650 14px system-ui,sans-serif;text-decoration:underline;text-underline-offset:.2em}.focus-note{font:700 10px system-ui,sans-serif;text-transform:uppercase;letter-spacing:.13em;margin-top:50px;color:var(--muted)}.feature{display:grid;grid-template-columns:1fr 1.05fr;align-items:center;gap:54px;border-top:1px solid var(--line);padding:50px 0;margin-top:88px}.feature img{width:100%;height:280px;object-fit:cover}.feature p:not(.eyebrow),.latest p{color:var(--muted);max-width:560px}.feature h2,.latest h2,.variant-picker h2{font-size:clamp(32px,5vw,48px);letter-spacing:-.06em;line-height:1.1;margin:10px 0 16px}.latest,.variant-picker,.contact-strip{border-top:1px solid var(--line);padding:38px 0}.latest h2{max-width:750px}.post-card{padding:26px 0;border-bottom:1px solid var(--line)}.post-card h2{font-size:clamp(24px,4vw,36px);line-height:1.15;letter-spacing:-.04em}.post-card p:last-child{color:var(--muted)}article{max-width:790px;margin-top:44px;font-size:18px;line-height:1.82}article h2{font-size:32px;margin-top:48px}article h3{margin-top:36px}.table-scroll{max-width:100%;overflow-x:auto;overscroll-behavior-inline:contain;border:1px solid var(--line)}.table-scroll:focus-visible{outline:3px solid var(--blue);outline-offset:3px}.table-scroll table{min-width:max-content;border-collapse:collapse}.table-scroll th,.table-scroll td{padding:8px 12px;border:1px solid var(--line);text-align:left}.table-scroll th{background:var(--surface)}article pre{overflow:auto;max-width:100%;padding:22px;background:var(--surface);border:1px solid var(--line);font-size:13px;line-height:1.55}article code{font-family:ui-monospace,SFMono-Regular,Menlo,monospace}article p code,article li code{overflow-wrap:anywhere}article blockquote{border-left:3px solid var(--blue);padding-left:20px;color:var(--muted)}article img{max-width:100%;height:auto}.series-nav{border-top:1px solid var(--line);margin-top:48px;padding-top:24px;display:grid;gap:14px}.series-nav p{margin:6px 0}.coming-soon{display:inline-block;margin-left:8px;color:var(--muted);font:700 11px system-ui,sans-serif}.variant-picker{margin-top:28px}.variant-controls{display:flex;flex-wrap:wrap;gap:8px;margin:24px 0}.variant-controls button{border:1px solid var(--line);background:var(--paper);color:var(--ink);padding:12px 15px;font:700 12px system-ui,sans-serif;cursor:pointer}.variant-controls button[aria-pressed=true]{border-color:var(--blue);background:var(--blue);color:var(--paper)}.variant-description{color:var(--muted)}.contact-strip{display:flex;align-items:center;justify-content:space-between;gap:20px}.contact-strip p{font-size:20px}footer{border-top:1px solid var(--line);border-bottom:0;color:var(--muted);font:13px system-ui,sans-serif}.post-card a:hover{text-decoration:underline}.site-header nav a:hover,footer nav a:hover{color:var(--blue)}body[data-variant=fieldnotes]{--paper:#f4efe5;--surface:#ebe4d8;--ink:#241f18;--muted:#686056;--blue:#a34d24;--line:#d2c8b9;font-family:Georgia,"Times New Roman",serif}html[data-theme=dark] body[data-variant=fieldnotes]{--paper:#211c17;--surface:#302920;--ink:#f4eadb;--muted:#c3b49f;--blue:#ffb286;--line:#514638}body[data-variant=fieldnotes] .hero{max-width:820px}body[data-variant=fieldnotes] h1{font-family:Georgia,"Times New Roman",serif;font-weight:500;letter-spacing:-.07em}body[data-variant=fieldnotes] .latest h2{letter-spacing:normal}body[data-variant=systems]{font-family:ui-monospace,SFMono-Regular,Menlo,monospace}html[data-theme=dark] body[data-variant=systems]{--paper:#10191b;--surface:#172426;--ink:#e8f1ed;--muted:#a6b7b1;--blue:#72dbb3;--line:#324345}body[data-variant=systems] h1,body[data-variant=systems] h2{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;letter-spacing:normal}body[data-variant=systems] .feature img{filter:saturate(.6) contrast(1.1)}@media(max-width:700px){.site-header,footer{padding:19px 20px}.site-header{gap:14px;flex-wrap:wrap}.site-header .wordmark{order:0}.site-header .theme-toggle{order:1;margin-left:auto}.site-header nav{order:2;flex:1 1 100%;gap:10px;flex-wrap:wrap;justify-content:flex-start;font-size:11px}.site-header nav a{white-space:nowrap}main{padding:60px 22px 72px}.feature{grid-template-columns:1fr;gap:20px;margin-top:64px}.feature img{height:210px}.contact-strip{align-items:flex-start;flex-direction:column;gap:5px}}@media(max-width:420px){.site-header{padding:17px 15px}.site-header nav{gap:6px;font-size:10px}.wordmark{font-size:21px}main{padding:50px 18px 64px}h1{font-size:clamp(52px,15vw,68px)}.actions{gap:17px}.feature{margin-top:52px}article{font-size:17px}article pre{padding:14px}.variant-controls button{flex:1 1 100%}footer{align-items:flex-start;flex-direction:column;gap:14px}}@media(prefers-reduced-motion:reduce){html{scroll-behavior:auto}*,*::before,*::after{transition-duration:.01ms!important;animation-duration:.01ms!important;animation-iteration-count:1!important}}`
