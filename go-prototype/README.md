---
layout: page
title: Go static-site generator
description: How the Go overlay builds portfolio pages alongside Jekyll's supporting routes.
permalink: /go-prototype/
---

The GitHub Pages workflow builds the full Jekyll site, then overlays Go-generated home, projects, about, resume, writing index, and each published, dated Markdown post whose date is not in the future. Both generators read the existing source Markdown; the Go generator does not move or rewrite it. The overlay preserves the current canonical paths, including the nested category/date routes for all five posts. Jekyll owns `/contact` and continues to supply auxiliary routes such as `/categories/`, `/tags/`, `/search/`, `/subscribe/`, RSS, sitemap, and other supporting assets.

The generator validates its output destination after resolving symlinks and rejects any path that overlaps or contains the source repository before removing output. To roll back the overlay, remove the `Overlay Go-generated pages` workflow step; Jekyll remains a complete fallback.

## Build and preview

From the repository root:

```sh
cd go-prototype
GO_PROTOTYPE_OUT=/tmp/kayis-go-prototype-site go run . ..
python3 -m http.server 4173 --directory /tmp/kayis-go-prototype-site
```

Open `http://localhost:4173/`. Run tests with `go test ./...` from `go-prototype`.

## Generated routes

- `/`, `/about/`, `/resume/`, `/projects/`, `/posts/`
- `/career/engineering/2026/05/17/ideas-to-production.html`
- `/llm/infrastructure/self-hosting/2026/05/19/why-i-stopped-paying-per-token.html`
- `/llm/infrastructure/self-hosting/2026/05/26/the-model-hunt-five-attempts-one-winner.html`
- `/llm/infrastructure/self-hosting/2026/06/02/the-gpu-graveyard-cuinit-999.html`
- `/llm/infrastructure/self-hosting/2026/06/09/the-config-that-finally-worked.html`

Post routes derive from source categories, dates, and filenames; unpublished and future-dated posts are excluded from generated routes, listings, latest-post selection, and series navigation. Posts retain their full Markdown bodies, canonical metadata, original inline links, and series navigation; Part 5 is still marked coming soon because it is not present in the source tree. The projects page now renders every existing project rather than only Synapse.

The homepage offers three keyboard-operable style variants (Signal, Field notes, Systems). Variant selection is client-side and leaves all content and links unchanged. Motion is limited to restrained color transitions; reduced-motion preferences are honored. No separate JavaScript bundle or external images are needed; the existing `img/home.jpg` asset is copied into the output.

All Go pages share the accessible light/dark toggle and saved `theme` preference used by Jekyll pages. First visit follows the system color preference; stored choices are used when browser storage is available. If storage is blocked, the system preference and in-page toggle still work. The theme is applied in the document head before the stylesheet to avoid a flash of the wrong theme.

## Auxiliary Jekyll routes

Categories, tags, client-side search, subscription forms, RSS, sitemap, 404 handling, and static metadata continue to be emitted by the full Jekyll build and are intentionally not reimplemented in Go in this pass. The workflow overlays Go-owned routes only; it does not remove these base-site routes or assets.
