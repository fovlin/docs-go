# docs-go

docs-go generates a static documentation site from Markdown files. A project directory contains a configuration, a sidebar file, Markdown sources and optional public assets. The `build` command renders the sources to HTML; the `dev` command serves the same pages from source over HTTP.

## Requirements

Building from source requires Go 1.27.2 or newer. The generated site has no runtime dependencies; it is HTML, CSS and a small amount of JavaScript.

## Installation

~~~bash
git clone https://github.com/fovlin/docs-go
cd docs-go
go build -o docs-go .
~~~

`./build.sh` builds the release targets. Release notes are in [CHANGELOG.md](CHANGELOG.md).

## Quick start

~~~bash
docs-go init my-docs
cd my-docs
docs-go dev
~~~

Open <http://localhost:7280/>. Files under `docs/` are read on every request, so refresh the browser after editing. To write the site to `dist/`:

~~~bash
docs-go build
~~~

## Commands

~~~text
docs-go <command> [options] [project]
~~~

| Command | Description |
| --- | --- |
| `build` | Render the site. This is the default command. |
| `dev` | Serve pages from source over HTTP. |
| `init` | Write a new project skeleton. |

`[project]` defaults to the current directory. Paths in the configuration are resolved against it.

## Options

| Option | Description |
| --- | --- |
| `-c, --config <file>` | Configuration file. Default: `config/docs.json`. |
| `-s, --src <dir>` | Override `src_dir`. |
| `-o, --out <dir>` | Override `out_dir`. |
| `-a, --addr <ip>` | Address for `dev`. Default: `localhost`. |
| `-p, --port <port>` | Port for `dev`. Default: `7280`. |
| `-f, --force` | Overwrite existing files in `init`. |
| `-t, --time` | Add timestamps to log lines. |
| `--color` | Force colored log output. |
| `-v, --version` | Print the version and exit. |
| `-h, --help` | Print usage and exit. |

## Project layout

~~~text
my-docs/
  config/
    docs.json            site settings
    sidebar.json         sidebar trees
  docs/
    index.md             /
    guide/
      getting-started.md
      configuration.md
  public/                copied to the root of the output
  dist/                  generated
~~~

## Pages and URLs

Every `.md` file under `src_dir` is a page. `clean_urls` selects the URL and the output layout:

| Source | `clean_urls: true` | `clean_urls: false` |
| --- | --- | --- |
| `index.md` | `/` → `index.html` | `/` → `index.html` |
| `guide/index.md` | `/guide` → `guide/index.html` | `/guide` → `guide/index.html` |
| `guide/start.md` | `/guide/start` → `guide/start/index.html` | `/guide/start.html` → `guide/start.html` |

With `clean_urls` enabled a page URL is its source path without the extension. With it disabled a non-index page is linked by the file it generates, while an index page keeps the directory URL `/dir` in both cases.

## Links

Link destinations are written to the output after normalization:

- destinations with a scheme or host (`http`, `https`, `mailto`, `//`) and in-page anchors (`#`) are left unchanged;
- other destinations are site paths: a leading `/` is expected, a `.md` or `.html` extension is removed, and the result is converted to the page URL;
- a trailing `/` is removed, so an index page is linked as `/guide/` and renders as `/guide`;
- query strings and fragments are preserved.

Links are not validated. `config/docs.json` and `config/sidebar.json` links are normalized the same way.

## Configuration

A site uses two files: `config/docs.json` for settings and `config/sidebar.json` for navigation. Every field is optional. See [docs/guide/configuration.md](docs/guide/configuration.md) for the full reference.

## Markdown

The parser supports a subset of Markdown. See [docs/guide/markdown.md](docs/guide/markdown.md).

## Build

`build` clears `out_dir`, copies source assets (Markdown files are skipped), writes the theme assets and favicon, renders every page, writes `404.html`, writes `sitemap.xml` when enabled, and finally copies `public_dir` over the output. Two pages that produce the same output path are an error. The output directory is rewritten in place, not swapped atomically, so a failed render can leave a partial site.

## Dev

`dev` resolves each request to a source file and renders it. It accepts the page URLs the build produces, serves assets from `src_dir` and `public_dir`, serves the theme assets under `/assets/`, renders the 404 page for unknown paths, and reloads the configuration when the config, sidebar or source directory changes. Markdown sources are not served directly. The browser is not reloaded automatically.

## Documentation

The site in `docs/` is built with docs-go itself:

~~~bash
docs-go build
~~~

## License

GNU General Public License v3. See [LICENSE](LICENSE).
