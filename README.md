# DOCS_GO

> A small, plain static documentation generator in Go, in the spirit of


[next-index](https://github.com/fovlin/next-index) and VitePress.

- **One binary.** No Node.js, no bundler, no plugin graph. The only external
  Go module is the small `github.com/fovlin/opt` flag parser; logging is a
  vendored `record` package.
- **Two config files.** `docs.json` for the site and `sidebar.json` for
  navigation, kept separate on purpose.
- **A real static site.** Every page is an HTML file. JavaScript only adds
  theme switching, copy buttons, the mobile drawer and the outline scrollspy.
- **A useful dev server.** Clean URLs, a 404 page and on-demand page rendering.
  Refresh the browser after editing; no background disk scan.

## Quick start

Requires Go 1.27.2 or newer.

```bash
go install acovia.net/docs-go@latest

docs-go init my-docs
cd my-docs
docs-go dev
```

Open <http://localhost:8080/>. Refresh the browser after editing a file under
`docs/`; the page is read from disk on every request.

To build once:

```bash
docs-go build
```

The output goes to `dist/`; docs-go clears it and writes the site directly.

## Project layout

```
my-docs/
  docs.json            site settings
  sidebar.json         navigation, kept out of the main config
  docs/
    index.md           /
    guide/
      index.md         /guide/
      getting-started.md
  public/              copied to the root of the generated site
  dist/                generated; do not edit
```

## Configuration

`docs.json` covers the title, description, language, base path, directories,
navigation, theme and footer. Every field is optional; missing fields use
defaults. Paths are relative to the project directory.

`sidebar.json` is a separate file. It accepts one array for every page:

```json
[
  {
    "text": "Guide",
    "items": [
      { "text": "Getting started", "link": "/guide/getting-started" }
    ]
  }
]
```

or an object keyed by URL prefix:

```json
{
  "/guide/": [
    {
      "text": "Guide",
      "items": [
        { "text": "Getting started", "link": "/guide/getting-started" }
      ]
    }
  ]
}
```

The longest matching prefix wins. See the
[documentation](https://github.com/fovlin/docs-go) for the full reference.
The source for this project's own documentation is in `docs/`; build it with:

```bash
docs-go build
```

## Commands

```
docs-go <command> [options] [project]
```

| Command | Description |
| --- | --- |
| `build` | Generate the site. |
| `dev` | Render pages on demand and preview over HTTP. |
| `init` | Write a new site skeleton. |

| Option | Description |
| --- | --- |
| `-c, --config <file>` | Config file, default `docs.json`. |
| `-s, --src <dir>` | Override `src_dir`. |
| `-o, --out <dir>` | Override `out_dir`. |
| `-a, --addr <ip>` | Listen address for `dev` (default `0.0.0.0`). |
| `-p, --port <port>` | Listen port for `dev` (default `8080`). |
| `-f, --force` | Overwrite scaffold files in `init`. |
| `-q, --quiet` | Do not print progress. |
| `-v, --version` | Print the version. |
| `-h, --help` | Print help. |

## Features

- Top navigation with active state, social links and a logo or built-in mark.
- Sidebar groups, nesting, collapsible sections, longest-prefix selection and
  automatic opening of the active group.
- Right-hand outline with scrollspy; levels configurable globally.
- Plain `<pre><code>` blocks with a copy button. No syntax highlighting, no
  line numbers and no code themes.
- A native Markdown subset: tables, task lists, strikethrough, autolinks,
  heading anchors, nested lists and blockquotes.
- VitePress-style `::: tip`, `info`, `note`, `warning`, `danger` and `details`
  blocks, including nested ones.
- No frontmatter. The first H1 is the page title and the first paragraph is the
  meta description; missing values fall back to the file name and site config.
- Markdown links to `.md` files are rewritten to the generated pages; missing
  targets are reported as warnings.
- Clean URLs, configurable base path, canonical URLs, `404.html` and an
  optional `sitemap.xml`.
- Direct output generation: `dist/` is cleared and rewritten on each build.
- A dev server that renders individual pages from source on request.

## Build

```bash
git clone https://github.com/fovlin/docs-go
cd docs-go
go build .
```

`./build.sh` cross-compiles the release targets.

## Dependencies

- [`github.com/fovlin/opt`](https://github.com/fovlin/opt) — command-line parsing.
- `internal/record` — vendored logging package, based on `next-index/record`.


## License

GNU General Public License v3. See [LICENSE](LICENSE).
