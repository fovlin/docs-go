# Docs

docs-go generates this site from the Markdown files in `docs/`.

## Quick start

~~~bash
docs-go init my-docs
cd my-docs
docs-go dev
~~~

`docs-go build` writes the site to `dist/`.

## Where things live

| Path | Purpose |
| --- | --- |
| `docs/` | Markdown source. Folders become URL segments. |
| `config/docs.json` | Site settings. |
| `config/sidebar.json` | Sidebar entries. |
| `public/` | Files copied to the root of the output. |
| `dist/` | Generated site. |

## Guide

- [Overview](/guide/)
- [Getting started](/guide/getting-started)
- [Configuration](/guide/configuration)
- [Markdown](/guide/markdown)
