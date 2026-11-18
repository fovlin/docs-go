# DOCS_GO

> A small, plain static documentation generator in Go, in the spirit of

[next-index](https://github.com/fovlin/next-index) and VitePress.

- **One binary.** No Node.js, no bundler, no plugin graph. The only external
  Go module is the small `github.com/fovlin/opt` flag parser; logging is a
  vendored `record` package.
- **Two config files.** `config/docs.json` for the site and `config/sidebar.json` for
  navigation, kept separate on purpose.
- **A real static site.** Every page is an HTML file. JavaScript only adds
  theme switching, copy buttons, the mobile drawer, the page menu and the
  outline scrollspy.
- **A useful dev server.** Clean URLs, a 404 page and on-demand page rendering.
  Refresh the browser after editing; no background disk scan.

## Quick start

Requires Go 1.27.2 or newer.

```bash
git clone https://github.com/fovlin/docs-go
cd docs-go
go build -o docs-go .
./docs-go init my-docs
cd my-docs
../docs-go dev
```

Open <http://localhost:7280/>. Refresh the browser after editing a file under
`docs/`; the page is read from disk on every request. Changes to `config/docs.json`
and `config/sidebar.json` are picked up on the next request.

To build once:

```bash
docs-go build
```

The output goes to `dist/`; docs-go clears it and writes the site directly.

## Project layout

```
my-docs/
  config/
    docs.json                site settings
    sidebar.json             navigation, kept out of the main config
  docs/
    index.md           /
    guide/
      getting-started.md
      configuration.md
  public/              copied to the root of the generated site
  dist/                generated; do not edit
```

## Configuration

`config/docs.json` covers the title, description, language, directories,
navigation and theme. Every field is optional; missing fields use
defaults. Paths are relative to the project directory.

`config/sidebar.json` is a separate file. It accepts one array for every page:

```json
[
  {
    "text": "Guide",
    "items": [
      { "text": "Getting started", "link": "/guide/getting-started/" }
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
        { "text": "Getting started", "link": "/guide/getting-started/" }
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
| `-c, --config <file>` | Config file, default `config/docs.json`. |
| `-s, --src <dir>` | Override `src_dir`. |
| `-o, --out <dir>` | Override `out_dir`. |
| `-a, --addr <ip>` | Listen address for `dev` (default `localhost`). |
| `-p, --port <port>` | Listen port for `dev` (default `7280`). |
| `-f, --force` | Overwrite scaffold files in `init`. |
| `-q, --quiet` | Do not print progress. |
| `-v, --version` | Print the version. |
| `-h, --help` | Print help. |

## Features

- Top navigation with active state, social links and an optional logo. No
  placeholder logo is rendered when `logo` is omitted.
- Sidebar groups, nesting, collapsible sections, longest-prefix selection and
  automatic opening of the active group. Sidebars may link to pages in other
  sidebars; the destination uses its own matching sidebar.
- Right-hand outline with scrollspy; levels configurable globally.
- Plain `<pre><code>` blocks with a copy button. No syntax highlighting, no
  line numbers and no code themes.
- A native Markdown subset: tables, heading anchors, nested lists, blockquotes,
  callout containers and inline HTML tags.
- No frontmatter. The first H1 is the page title and the first paragraph is the
  meta description; missing values fall back to the file name and site config.
- Links are written into the page as authored; they are not rewritten or
  validated, so the author points them at the URL the build emits.
- Clean URLs (`<name>/index.html` folders), canonical URLs, `404.html` and an optional `sitemap.xml`.
- Direct output generation: `dist/` is cleared and rewritten on each build.
- A dev server that renders individual pages from source on request and
  reloads config/sidebar snapshots without a full rebuild.

## Markdown subset

The parser is deliberately small. Unsupported syntax is not an error; it
usually renders as plain text.

- ATX headings (`#` to `######`), GitHub-style anchors and duplicate suffixes.
- Paragraphs, entity references, inline code, emphasis and strong.
- Links and images.
- Fenced code blocks (backticks or tildes), indented code blocks, blockquotes
  and thematic breaks.
- Ordered and unordered lists and nested lists. Blank lines do not end a list.
- GFM-style tables with column alignment.
- `::: tip`-style containers (`tip`, `info`, `note`, `warning`, `danger`,
  `details`), including nested ones.
- Inline HTML tags inside paragraphs.

Not supported: setext headings, reference links and definitions, footnotes,
task lists, strikethrough, autolinks, custom container names, explicit heading
IDs, block-level HTML, and links nested inside links.

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
