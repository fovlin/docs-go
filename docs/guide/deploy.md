# Deploying

`docs-go build` writes a plain static site to `dist/`. There is no server-side
runtime.

## Before you deploy

1. Set `site_url` if you want canonical URLs or a sitemap.
2. Set `base` if the site is not served from the domain root.
3. Decide whether your host supports clean URLs.

~~~json
{
  "base": "/",
  "site_url": "https://docs.example.com",
  "sitemap": true
}
~~~

## Clean URLs

With `"clean_urls": true` a source file `guide/start.md` is written to
`guide/start.html`, but every link uses `/guide/start`. The dev server resolves it. Static hosts differ:

| Host | Clean URLs |
| --- | --- |
| Netlify, Vercel, Cloudflare Pages | Usually supported. |
| GitHub Pages | Serves `/foo` as `foo.html` in many configurations; verify yours. |
| A plain file server | Set `clean_urls` to `false`. |

If in doubt, set `"clean_urls": false`. Links then include the `.html`
extension and any static file server works without rewrites.

## Base path

For `https://example.com/docs/`:

~~~json
{
  "base": "/docs/",
  "site_url": "https://example.com"
}
~~~

The generated assets and links are prefixed with `/docs/`, and the preview
server redirects `/` to `/docs/`.

## 404

A `404.html` is generated at the root of the output. It uses absolute asset URLs,
so it works from any path. Most static hosts pick it up automatically.

## Caching

Asset URLs carry a content hash as a query string, so they can be cached
aggressively. HTML pages should be served with a short cache or revalidated.

## GitHub Pages example

~~~bash
docs-go build
# copy or publish dist/ with your preferred action
~~~

Because the output contains no client-side router, every page is a real HTML
file and works without JavaScript for reading. JavaScript adds theme
switching, copy buttons, the mobile drawer and the outline scrollspy.
