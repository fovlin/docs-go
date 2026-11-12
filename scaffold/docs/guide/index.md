# Guide

This guide covers the small set of files that make up a docs-go site.

## Pages are files

Every `.md` file under `docs/` becomes a page:

| Source | URL |
| --- | --- |
| `docs/index.md` | `/` |
| `docs/guide/index.md` | `/guide/` |
| `docs/guide/getting-started.md` | `/guide/getting-started` |

With `clean_urls` enabled the generated file is still
`guide/getting-started.html`, but links omit the extension. Set
`"clean_urls": false` if your host cannot serve extensionless URLs.
