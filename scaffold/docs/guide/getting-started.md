# Getting started

## Build the site

~~~bash
docs-go build
~~~

The output goes to `dist/` by default. The directory is replaced atomically, so
a failed build never leaves half a site behind.

## Preview with live reload

~~~bash
docs-go serve -p 8080
~~~

`serve` watches the source directory, the config and the sidebar file. When
something changes it rebuilds and the browser reloads automatically.

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
