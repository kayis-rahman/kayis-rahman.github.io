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

## Newsletter & Subscriber Strategy
**High-impact placements (prioritize):**
1. **Post footer** (`_layouts/post.html`) — readers most engaged, drives 10-15% conversion
2. **Author card** (`_includes/author-card.html`) — visible on every post
3. **Dedicated `/subscribe/` page** — landing page for newsletter value prop

**Newsletter copy guide:**
- Strong heading: "Join engineers in the loop" (not "Stay updated")
- Value prop + reassurance: "PKI, Java, LLM inference... No spam. Unsubscribe anytime."
- Use reusable `_includes/newsletter-cta.html` include for consistency

**Buttondown integration:**
- Config: `newsletter_username: kayis` in `_config.yml`
- Form action: `https://buttondown.email/api/emails/embed-subscribe/{username}`
- Submit button label should have arrow: "Subscribe →"

## Post Frontmatter Fields
Use these fields in `_posts/` frontmatter for better SEO + discovery:
```yaml
---
layout: post
title: "Post Title"
date: 2026-05-19
categories: [llm, infrastructure, self-hosting]  # Use lowercase, for /categories/ page
tags: [vllm, gpuhub, claude-code]                # Individual topic keywords
description: "Unique 160-char description for meta tags and preview text"
background: /assets/img/posts/hero.jpg           # Post hero/masthead image
image: /assets/img/og-image.jpg                   # og:image meta tag (social sharing)
word_count: 4500                                  # For schema.org BlogPosting
reading_time: 12                                  # Displayed in post meta
series: "Series Name"                             # For series-nav.html include
series_part: 1
---
```

Categories show up on `/categories/` page and post cards. Tags link to `/tags/` page. Use `background:` for post hero image (checked in `_layouts/post.html`), not `image:`.

## Growth Audit Pattern
**Before writing new posts, audit for quick wins across three dimensions:**

| Dimension | Quick Win | Impact |
|-----------|-----------|--------|
| Traffic | Hero CTA buttons | Direct action, stops bounces |
| | Category pages | SEO + discovery without content |
| | Post excerpt length | Better hooks on home/listing |
| Subscribers | Post footer newsletter | 10-15% conversion from engaged readers |
| | Author card subscribe link | Visible on every post |
| | Dedicated /subscribe/ page | Value prop landing page |
| Design | Hero avatar + CTA | Humanize + drive action |
| | Category badges | Visual type indicators |
| | Footer nav | Discoverability + subscribe link |

See memory `growth_audit_pattern.md` for detailed approach and when to apply.

## Color System & Theme
**Aesthetic: Brutalist minimalism + technical accent**

- **Base**: Pure black (`#000000`) and white (`#ffffff`) — brutal, unforgettable, zero distraction
- **Accent**: Bright blue (`#0099ff`) — reserved for CTAs (Subscribe, Read Blog), links, and interactive states
- **Dark theme**: Black background, white text, same bright blue accent

**CSS Variables** (`_sass/styles.scss`):
```scss
:root {
  --color-primary: #0099ff;    // Bright blue accent only
  --color-text: #000000;       // Black text on white
  --color-bg: #ffffff;         // White background
  --color-border: #cccccc;     // Light gray borders
  --color-link-hover: #0099ff; // Bright blue on hover
}

html[data-theme="dark"] {
  --color-primary: #0099ff;
  --color-text: #ffffff;
  --color-bg: #000000;
  --color-border: #333333;
  --color-link-hover: #0099ff;
}
```

Use `var(--color-primary)` for all accents. Bright blue should appear only on interactive elements (buttons, links, hover states), not as a dominant color.

**STRICT RULE: No Hardcoded Colors**
Never use hex values (`#fff`), `rgb()`, or `rgba()` with literal numbers directly in component styles in `_sass/styles.scss`. Every color must come from a CSS variable defined in `:root` or `html[data-theme="dark"]`. When adding a new color:
1. Define the semantic variable in both `:root` (light) and `html[data-theme="dark"]` (dark)
2. Use `var(--color-name)` in component styles

Documented exceptions (leave as-is):
- Syntax highlighting token colors in `.highlight` block (GitHub Solarized theme)
- Print media colors in `@media print` (CSS variables unreliable in print)
- Semantic alert colors in `blockquote.prompt-*` (tip/info/warning/danger)

## Home Header Positioning
Push name and buttons down slightly to reveal more of the face background image. Use `margin-top` on `.home-heading`:
- Mobile: `20px`
- Desktop: `30px`

Don't change masthead padding — only adjust content margin. Keeps header visual balance while showing more of the background photo.

## Gotchas
- **Before adding scripts/includes, check `default.html`** — it already includes `google-analytics.html`, `scripts.html`, `navbar.html`, and `footer.html`. Adding the same script to both caused duplicate GA tags. Remove the old include after adding to a new file.
- **CI minification can't write in-place to `_site/`** — `actions/configure-pages` makes `_site/` read-only. Minify to a temp dir (`_minified/`) and upload that as the artifact instead.
- **`_includes/scripts.html` is the canonical place for inline scripts** (GA, etc.) since it's always included via `default.html`. Don't add scripts to `default.html` directly — use the existing include partials.
- **`jekyll serve` requires `webrick` gem** — Ruby 3.3+ doesn't include webrick in stdlib. Add `gem "webrick"` to Gemfile and run `bundle install`. Without it, `jekyll build` works but `jekyll serve` crashes with LoadError.
- **Content column width is controlled by two things**: Bootstrap grid class (`col-lg-8`/`col-lg-10`) in layouts AND `.container` max-width in SCSS. Changing only one gives partial results. Change both for full effect.
- **SRI integrity hashes for CDN resources become stale** — jQuery, Bootstrap, and Font Awesome hashes in `head.html` and `scripts.html` require periodic updates. When resources are blocked (check console errors), fetch the actual computed hash from the browser error message and update the `integrity` attribute. See May 2026 session for example: FA `sha384-DyZ88MY4...` became `sha384-DyZ88mC6Up...`
- **Font Awesome async preload (`rel="preload" onload="..."`)** is unreliable for CSS. Use direct `<link rel="stylesheet">` instead. The async preload trick's `onload` handler doesn't reliably fire when integrity checks are involved or under network delays.
- **Navbar opacity issue with vendor theme** — the Clean Blog vendor SCSS makes the navbar `background: transparent` at desktop (≥992px) with white text, then uses `.is-fixed` class (added via missing scroll JS) to make it opaque. Without scroll JS, navbar stays transparent and text becomes invisible after scrolling past the dark masthead. Override in `_sass/styles.scss` with solid white background at desktop to fix.
- **Remove `future: true` from config** — prevents accidental publication of draft posts with future dates. With this setting enabled, any post dated in the future will publish immediately. Use post frontmatter draft flag or manual scheduling instead.
