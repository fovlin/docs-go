# Guide

docs-go reads Markdown files and writes one HTML page per file. This guide covers the project files, the generated URLs and the configuration.

- [Getting started](/guide/getting-started) — install, create a project, build and preview.
- [Configuration](/guide/configuration) — every field in `config/docs.json` and `config/sidebar.json`.
- [Markdown](/guide/markdown) — the supported syntax.

## Pages

Every Markdown file under `src_dir` becomes a page. An `index.md` file maps to its directory; other files map to their name. `clean_urls` selects the URL and the output layout:

| Source | `clean_urls: true` | `clean_urls: false` |
| --- | --- | --- |
| `docs/index.md` | `/` → `index.html` | `/` → `index.html` |
| `docs/guide/index.md` | `/guide` → `guide/index.html` | `/guide` → `guide/index.html` |
| `docs/guide/getting-started.md` | `/guide/getting-started` → `guide/getting-started/index.html` | `/guide/getting-started.html` → `guide/getting-started.html` |

With `clean_urls` enabled a page URL is its source path without the extension. With it disabled a non-index page is linked by the file it generates, while an index page keeps the directory URL `/dir` in both cases.

## Links

Link destinations are normalized before they are written:

- destinations with a scheme or host (`http`, `https`, `mailto`, `//`) and in-page anchors (`#`) are left unchanged;
- other destinations are site paths: a leading `/` is expected, a `.md` or `.html` extension is removed, and the result is converted to the page URL;
- a trailing `/` is removed, so an index page is linked as `/guide/` and renders as `/guide`;
- query strings and fragments are preserved.

Links are not checked. The author is responsible for pointing them at the page that should be reached.
