# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this repository.

## Project Overview
Personal portfolio/blog site deployed on GitHub Pages (`kayis-rahman.github.io`). Built with Jekyll 4.4 + `startbootstrap-clean-blog-jekyll` remote theme.

## Structure
```
_config.yml          # Site config: title, navigation, remote theme, URL
index.markdown       # Home page (layout: home)
about.markdown       # About page (layout: page, permalink: /about/)
resume.markdown      # Resume/CV (layout: page, permalink: /resume/)
contact.html         # Contact form (layout: page, permalink: /contact)
_posts/              # Blog posts (chronological filenames)
_includes/           # Custom partials: footer, head, navbar, scripts
_layouts/            # Custom layouts: default, home, page, post
_sass/styles.scss    # Imports Clean Blog SCSS from vendor
img/                 # Background images referenced in frontmatter
```

## Common Commands
```bash
# Install dependencies
bundle install

# Local dev server (requires restart after _config.yml changes)
bundle exec jekyll serve

# Build static site to _site/
bundle exec jekyll build

# Run on a specific port
bundle exec jekyll serve --port 4000
```

## Ruby Version
This project requires Ruby 3.3.x (set via `.ruby-version`). Ruby 4.0+ is incompatible with the `github-pages` gem stack (`jekyll-feed` requires Ruby < 4.0).

If you have `rbenv` installed, it will automatically use the correct Ruby version. Install with:
```bash
brew install rbenv
rbenv install 3.3.8
```

A `pre-push` hook runs `jekyll build` before each push to catch build errors early.

## Deployment
Pushes to `main` trigger a GitHub Actions workflow (`.github/workflows/jekyll-gh-pages.yml`) that builds and deploys to GitHub Pages. No manual steps needed. Manual runs via `workflow_dispatch` on the Actions tab.

## Theme Override Pattern
The remote theme (`StartBootstrap/startbootstrap-clean-blog-jekyll`) provides base layouts, includes, and SCSS. Custom files in `_layouts/`, `_includes/`, and `_sass/` **override** the theme's versions automatically — Jekyll prefers local files over theme-generated ones. To modify site-wide structure, edit the corresponding file in `_layouts/` or `_includes/` rather than trying to patch the theme.

## Key Conventions
- Pages use YAML frontmatter for layout, title, permalink, and background image
- Background images are referenced as `'/img/<name>.jpg'` in frontmatter
- Navigation is defined in `_config.yml` under `navigation:` key
- `baseurl` is empty string (site deploys at repo root, not a subpath)
- Blog posts go in `_posts/` with filenames like `YYYY-MM-DD-slug.markdown`
- The Gemfile lists both `minima` and `jekyll-theme-clean-blog`; the active theme is the remote theme in `_config.yml`. `minima` is vestigial.

## UI Verification Rule
**After any UI change (layouts, includes, SCSS, CSS), verify visually before reporting completion:**
1. Start dev server: `bundle exec jekyll serve --port 4000`
2. Navigate to affected pages with Playwright (`browser_navigate`)
3. Take snapshots (`browser_snapshot`) to verify layout, spacing, and content flow
4. Take screenshots (`browser_take_screenshot`) for visual confirmation
5. Check computed styles (`browser_evaluate`) for width/height changes
6. Verify on post page, home page, and any modified layout
7. Stop dev server when done

Builds pass != UI is correct. Always verify rendering in browser.

## Gotchas
- **Before adding scripts/includes, check `default.html`** — it already includes `google-analytics.html`, `scripts.html`, `navbar.html`, and `footer.html`. Adding the same script to both caused duplicate GA tags. Remove the old include after adding to a new file.
- **CI minification can't write in-place to `_site/`** — `actions/configure-pages` makes `_site/` read-only. Minify to a temp dir (`_minified/`) and upload that as the artifact instead.
- **`_includes/scripts.html` is the canonical place for inline scripts** (GA, etc.) since it's always included via `default.html`. Don't add scripts to `default.html` directly — use the existing include partials.
- **`jekyll serve` requires `webrick` gem** — Ruby 3.3+ doesn't include webrick in stdlib. Add `gem "webrick"` to Gemfile and run `bundle install`. Without it, `jekyll build` works but `jekyll serve` crashes with LoadError.
- **Content column width is controlled by two things**: Bootstrap grid class (`col-lg-8`/`col-lg-10`) in layouts AND `.container` max-width in SCSS. Changing only one gives partial results. Change both for full effect.
- **SRI integrity hashes for CDN resources become stale** — jQuery, Bootstrap, and Font Awesome hashes in `head.html` and `scripts.html` require periodic updates. When resources are blocked (check console errors), fetch the actual computed hash from the browser error message and update the `integrity` attribute. See May 2026 session for example: FA `sha384-DyZ88MY4...` became `sha384-DyZ88mC6Up...`
- **Font Awesome async preload (`rel="preload" onload="..."`)** is unreliable for CSS. Use direct `<link rel="stylesheet">` instead. The async preload trick's `onload` handler doesn't reliably fire when integrity checks are involved or under network delays.
- **Navbar opacity issue with vendor theme** — the Clean Blog vendor SCSS makes the navbar `background: transparent` at desktop (≥992px) with white text, then uses `.is-fixed` class (added via missing scroll JS) to make it opaque. Without scroll JS, navbar stays transparent and text becomes invisible after scrolling past the dark masthead. Override in `_sass/styles.scss` with solid white background at desktop to fix.
