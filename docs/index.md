# Docs

A small documentation site generated from the markdown files in `docs/`.

::: tip
Edit the files under `docs/` and run `docs-go dev` while you write. The page
you are reading was generated from `docs/index.md`.
:::

## Quick start

~~~bash
docs-go build      # write the site to dist/
docs-go dev      # build, watch for changes and preview on :8080
~~~

xxx <https://sufox.net>

## Where things live

| Path | Purpose |
| --- | --- |
| `docs/` | Markdown source. Folders become URL segments. |
| `docs.json` | Site settings: title, navigation, theme. |
| `sidebar.json` | Sidebar entries, kept out of the main config. |
| `public/` | Static files copied to the root of the output. |
| `dist/` | Generated site. Never edit it by hand. |

## Next steps

- [Getting started](/guide/getting-started)
- [Configuration](/guide/configuration)
