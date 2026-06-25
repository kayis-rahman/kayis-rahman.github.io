source "https://rubygems.org"

# Explicit deps for Ruby 3.4+ (removed from stdlib)
gem "csv"
gem "bigdecimal"
gem "json"
gem "webrick"
# Liquid 4.0.4+ drops String#tainted? call removed in Ruby 3.2+
gem "liquid", "~> 4.0.4"

# Use the same gem stack as GitHub Pages CI for consistent local builds.
gem "github-pages", group: :jekyll_plugins

group :jekyll_plugins do
  gem "jekyll-feed", "~> 0.12"
end
