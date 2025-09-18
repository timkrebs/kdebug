# kdebug Website

This directory contains the GitHub Pages website for kdebug, built with Jekyll and inspired by the Kubernetes documentation design.

## Local Development

### Prerequisites

- Ruby 3.0 or later
- Bundler gem
- Git

### Quick Start

1. **Install dependencies**:
   ```bash
   make website-deps
   ```

2. **Start development server**:
   ```bash
   make website-serve
   ```

3. **Open your browser** to `http://localhost:4000`

The site will automatically reload when you make changes to files.

### Manual Setup

If you prefer to run Jekyll commands directly:

```bash
# Install dependencies
bundle install

# Serve locally with live reload
bundle exec jekyll serve --livereload --drafts

# Build for production
bundle exec jekyll build
```

## Site Structure

```
kdebug/
├── _config.yml          # Jekyll configuration
├── _layouts/            # Page layouts
│   ├── default.html     # Base layout
│   ├── home.html        # Homepage layout
│   └── docs.html        # Documentation layout
├── _includes/           # Reusable components
│   ├── head.html        # HTML head section
│   ├── header.html      # Site header
│   └── footer.html      # Site footer
├── _sass/               # Sass stylesheets
│   ├── _variables.scss  # Design variables
│   ├── _base.scss       # Base styles
│   ├── _header.scss     # Header styles
│   ├── _home.scss       # Homepage styles
│   ├── _docs.scss       # Documentation styles
│   └── _footer.scss     # Footer styles
├── _docs/               # Documentation pages
│   ├── getting-started.md
│   ├── installation.md
│   ├── commands.md
│   ├── examples.md
│   └── contributing.md
├── assets/              # Static assets
│   ├── main.scss        # Main stylesheet
│   └── images/          # Images and logos
├── index.md             # Homepage
└── Gemfile              # Ruby dependencies
```

## Design System

The website uses a Kubernetes-inspired design system with:

### Colors

- **Primary Blue**: `#326ce5` - Main brand color
- **Secondary Blue**: `#f0f7ff` - Light backgrounds
- **Dark Blue**: `#1a1a2e` - Headings and dark text
- **Success Green**: `#28a745` - Success states
- **Warning Orange**: `#fd7e14` - Warning states

### Typography

- **Font Family**: System fonts (-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, etc.)
- **Headings**: Weights 600, clean hierarchy
- **Body Text**: Medium gray (`#6c757d`) for readability

### Components

- **Cards**: Rounded corners, subtle shadows, hover effects
- **Buttons**: Primary/secondary variants with hover animations
- **Code Blocks**: Dark theme with syntax highlighting
- **Navigation**: Sticky header with responsive mobile menu

## Content Guidelines

### Writing Style

- **Clear and concise**: Use simple, direct language
- **Action-oriented**: Start with verbs (Install, Configure, Run)
- **Scannable**: Use headings, lists, and code blocks
- **Helpful**: Include examples and context

### Documentation Structure

1. **Overview**: What the feature does
2. **Prerequisites**: What users need first
3. **Step-by-step**: Clear instructions
4. **Examples**: Real-world usage
5. **Troubleshooting**: Common issues and solutions

### Code Examples

Use realistic examples with proper syntax highlighting:

```bash
# Good: Real command with context
kdebug pod nginx-deployment-7d4b8c6f9-x8k2l --verbose

# Avoid: Generic placeholder
kdebug pod <pod-name> [flags]
```

## Adding New Content

### Documentation Pages

1. **Create markdown file** in `_docs/` directory
2. **Add front matter**:
   ```yaml
   ---
   layout: docs
   title: Your Page Title
   description: Brief description for SEO
   permalink: /docs/your-page/
   order: 3
   ---
   ```
3. **Add to navigation** in `_config.yml` if needed
4. **Test locally** with `make website-serve`

### Homepage Updates

Edit `_layouts/home.html` and update:
- Hero section content
- Feature cards in `_config.yml`
- Quick start examples

### Styling Changes

1. **Edit Sass files** in `_sass/` directory
2. **Follow existing patterns** and color scheme
3. **Test responsive design** on mobile/tablet
4. **Maintain accessibility** (contrast, focus states)

## Deployment

The website deploys automatically via GitHub Actions when changes are pushed to the main branch.

### Automatic Deployment

- **Trigger**: Push to main branch with website files
- **Process**: Jekyll build → GitHub Pages deployment
- **URL**: `https://your-username.github.io/kdebug/`

### Manual Deployment

For testing deployment locally:

```bash
# Build production site
make website-build

# Check for issues
make website-check
```

## Configuration

### GitHub Pages Setup

1. **Enable GitHub Pages** in repository settings
2. **Set source** to "GitHub Actions"
3. **Configure custom domain** (optional) in `_config.yml`

### Site Configuration

Key settings in `_config.yml`:

```yaml
title: kdebug
description: Your site description
baseurl: "/kdebug"  # Repository name
url: "https://your-username.github.io"

# Update these with your details
github_username: your-username
github_repo: kdebug
```

## Testing

### Local Testing

```bash
# Serve locally
make website-serve

# Build and check
make website-prod
```

### Checks to Perform

- ✅ All links work (internal and external)
- ✅ Images load correctly
- ✅ Responsive design on mobile/tablet
- ✅ Fast page load times
- ✅ Good SEO metadata
- ✅ Accessible navigation

## Troubleshooting

### Common Issues

**Jekyll won't start**:
```bash
# Check Ruby version
ruby --version

# Reinstall dependencies
bundle clean --force
bundle install
```

**Build fails**:
```bash
# Check for syntax errors
bundle exec jekyll build --verbose

# Clean and rebuild
make website-clean
make website-build
```

**Styles not updating**:
```bash
# Clear Jekyll cache
rm -rf .jekyll-cache .sass-cache

# Hard refresh browser (Cmd+Shift+R)
```

### Getting Help

- Check Jekyll documentation: https://jekyllrb.com/docs/
- Review GitHub Pages docs: https://docs.github.com/pages
- Test with different browsers and devices
- Validate HTML: https://validator.w3.org/

## Contributing

When contributing to the website:

1. **Test locally** before submitting PRs
2. **Follow design patterns** established in existing pages
3. **Optimize images** and assets for web
4. **Check accessibility** (color contrast, keyboard navigation)
5. **Update documentation** if adding new features

## Resources

- [Jekyll Documentation](https://jekyllrb.com/docs/)
- [GitHub Pages Documentation](https://docs.github.com/pages)
- [Liquid Template Language](https://shopify.github.io/liquid/)
- [Sass Documentation](https://sass-lang.com/documentation)
- [Kubernetes.io Design Reference](https://kubernetes.io/docs/home/)