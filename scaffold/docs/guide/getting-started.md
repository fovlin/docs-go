# Getting started

This guide covers the small set of files that make up a docs-go site.

## Pages are files

Every `.md` file under `docs/` becomes a page:

| Source | URL |
| --- | --- |
| `docs/index.md` | `/` |
| `docs/guide/index.md` | `/guide/` (optional directory index) |
| `docs/guide/getting-started.md` | `/guide/getting-started` |

With `clean_urls` enabled the generated file is still
`guide/getting-started.html`, but links omit the extension. Set
`"clean_urls": false` if your host cannot serve extensionless URLs.

## Build the site

~~~bash
docs-go build
~~~

The output goes to `dist/` by default. The directory is cleared and written
directly; it is not an atomic swap, so a failed render can leave a partial site.

## Preview

~~~bash
docs-go dev
~~~

The dev server listens on <http://localhost:7280> by default. Pages are read
from source on each request. Changes to `docs.json` and `sidebar.json` are
picked up on the next request, but the browser is not reloaded automatically;
refresh manually.

`serve` is accepted as an alias for `dev`, but use `dev` in new scripts.

## Add a page

Create `docs/guide/my-page.md`, then add it to `sidebar.json`:

~~~json
{
  "/guide/": [
    {
      "text": "Guide",
      "items": [
        { "text": "My page", "link": "/guide/my-page" }
      ]
    }
  ]
}
~~~

Links may point at `/guide/my-page`, `/guide/my-page.md` or
`/guide/my-page.html`; all three resolve to the page.
