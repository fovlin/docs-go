# Getting started

This guide covers the small set of files that make up a docs-go site.

## Pages are files

Every `.md` file under `docs/` becomes a page:

| Source | URL |
| --- | --- |
| `docs/index.md` | `/` |
| `docs/guide/index.md` | `/guide/` (optional directory index) |
| `docs/guide/getting-started.md` | `/guide/getting-started/` |

With `clean_urls` enabled each non-index page is written to `guide/getting-started/index.html` and linked as `/guide/getting-started/`. Set `"clean_urls": false` to write `guide/getting-started.html` instead.

## Build the site

~~~bash
docs-go build
~~~

The output goes to `dist/` by default. The directory is cleared and written directly; it is not an atomic swap, so a failed render can leave a partial site.

## Preview

~~~bash
docs-go dev
~~~

The dev server listens on <http://localhost:7280> by default. Pages are read from source on each request. Changes to `config/docs.json` and `config/sidebar.json` are picked up on the next request, but the browser is not reloaded automatically; refresh manually.

`serve` is accepted as an alias for `dev`, but use `dev` in new scripts.

## Add a page

Create `docs/guide/my-page.md`, then add it to `config/sidebar.json`:

~~~json
{
  "/guide/": [
    {
      "text": "Guide",
      "items": [
        { "text": "My page", "link": "/guide/my-page/" }
      ]
    }
  ]
}
~~~

Links are used exactly as written, so point them at the URL the build emits: `/guide/my-page/` with clean URLs, `/guide/my-page.html` without.
