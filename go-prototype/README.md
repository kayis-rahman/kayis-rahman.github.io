# Go static-generator prototype (reversible)

This is an isolated, HTML-first prototype. Existing Jekyll, GitHub Pages workflow, and public content/routes are unchanged. It reads the current `index.markdown`-adjacent site identity, `projects.md`, and a representative post directly from the existing source tree; it does not copy or rewrite the Markdown. Generated files default to the operating system's temp directory as `kayis-go-prototype-site` (override with `GO_PROTOTYPE_OUT`), outside the repository so they cannot contaminate Jekyll's build.

## Build and preview

From the repository root:

```sh
cd go-prototype
go mod download
GO_PROTOTYPE_OUT=/tmp/kayis-go-prototype-site go run . ..
python3 -m http.server 4173 --directory /tmp/kayis-go-prototype-site
```

Then open `http://localhost:4173/`. Routes generated: `/`, `/projects/`, and `/llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html`. The project page intentionally contains the existing Synapse section only. The home page introduces the engineering focus and links to the project and representative post. This prototype is not a replacement for Jekyll and intentionally does not port its complete search, taxonomy, RSS, newsletter, or auxiliary-page surface.

Run checks with `go test ./...` from `go-prototype`; it builds in a temporary directory and asserts the rendered routes, content, canonical origin, and zero JavaScript files.

## Comparison against Jekyll

Build Jekyll independently with `bundle exec jekyll build --destination /tmp/kayis-jekyll-site` at the repository root. Inspect Jekyll's corresponding HTML at `/`, `/projects/`, and `/llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html` before comparing. Both emit static HTML for GitHub Pages and retain `https://kayis-rahman.github.io` as canonical origin and current route conventions.

- **Built HTML:** prototype page bytes and total HTML bytes can be measured with `wc -c`; compare semantic page content, metadata, and links rather than raw formatting alone.
- **JavaScript transfer:** this prototype ships no JavaScript. Jekyll's exact transfer depends on production minification/theme output; use the generated asset references and file sizes (gzip included if served with compression) for a fair same-scope comparison.
- **Authoring:** edit the original Jekyll Markdown files as before; the prototype reads the post frontmatter and body and selects the existing Synapse prose. No content migration or custom content schema is introduced.
- **Build complexity:** Go generator has two dependencies (Goldmark and yaml.v3), direct `go run`/`go test` commands and static file output; Jekyll remains the production build and requires its existing Ruby/Bundler/theme/plugin toolchain. The generator does not claim feature parity and would need deliberate work for all current site features.

Prototype evidence (render checks, route checks, HTML sizes, JavaScript transfer, and local screenshot) belongs outside the repository under the task evidence directory. Do not change deployment or cut over public routes based on this experiment alone.
