# Go static-generator prototype (reversible)

This is an HTML-first Go generator for three selected pages. The GitHub Pages workflow still builds the complete site with Jekyll, then overlays the Go-generated `/`, `/projects/`, and representative post pages before minification and deployment. This keeps all other Jekyll routes and assets intact while these selected routes use the Go output. The generator reads the home description from `index.markdown`, the selected project from `projects.md`, and the representative post directly from the existing source tree; it does not copy or rewrite the Markdown. Generated files default to the operating system's temp directory as `kayis-go-prototype-site` (override with `GO_PROTOTYPE_OUT`). Before removing any output, the generator resolves path components and symlinks and rejects outputs inside, equal to, or containing the source repository. To roll back the cutover, remove the `Overlay Go-generated pages` step from `.github/workflows/jekyll-gh-pages.yml`; the Jekyll build remains the complete fallback site.

## Build and preview

From the repository root:

```sh
cd go-prototype
go mod download
GO_PROTOTYPE_OUT=/tmp/kayis-go-prototype-site go run . ..
python3 -m http.server 4173 --directory /tmp/kayis-go-prototype-site
```

Then open `http://localhost:4173/`. Routes generated: `/`, `/projects/`, and `/llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html`. The project page intentionally contains the existing Synapse section only. The home page introduces the engineering focus and links to the project and representative post. The Go generator is not a replacement for Jekyll and intentionally does not port its complete search, taxonomy, RSS, newsletter, or auxiliary-page surface; those features remain available on the Jekyll-generated routes and assets that are not overlaid.

Run checks with `go test ./...` from `go-prototype`; they assert rendered routes, preserved home Markdown description, canonical origin, zero JavaScript files, and that unsafe destination paths (including traversal and symlink forms) fail without removing a source sentinel.

## Comparison against Jekyll

Measured on the existing source tree for `/`, `/projects/`, and `/llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html`. Both emit static HTML for GitHub Pages and retain `https://kayis-rahman.github.io` as canonical origin and current route conventions.

- **Built HTML bytes** (Go prototype / Jekyll): home, 1,967 / 17,613; projects, 1,529 / 11,836; representative post, 19,314 / 46,794. Total: 22,810 / 76,243 bytes. This compares the three generated pages only and excludes other Jekyll routes and assets.
- **JavaScript transfer:** the prototype has no JavaScript files (0 bytes). Jekyll's local theme and site scripts are 1,966 + 9,892 = 11,858 bytes uncompressed; external jQuery and Bootstrap are also referenced, but their transfer sizes were not measured.
- **Authoring:** both use the existing Markdown sources. The prototype reads the home description and post frontmatter/body and selects the existing Synapse prose; no content migration or custom content schema is introduced.
- **Build complexity:** the Go generator has two direct module dependencies (Goldmark and yaml.v3), with direct `go run`/`go test` commands and static file output. The Gemfile has seven direct gem declarations: six in addition to `github-pages`. Jekyll remains the production build and requires its existing Ruby/Bundler/theme/plugin toolchain. The generator does not claim feature parity and would need deliberate work for all current site features.

The deployment overlay is intentionally limited to the prototype's three generated pages. Future expansion to other routes requires deliberate parity work and route-by-route validation.
