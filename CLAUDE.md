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
