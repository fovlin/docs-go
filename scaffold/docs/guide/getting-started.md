# Getting started

This chapter covers the project files, the first page and the two commands used while writing.

## Requirements

Go 1.27.2 or newer builds docs-go from source. The generated site has no runtime dependencies.

## Create a project

~~~bash
docs-go init my-docs
~~~

`init` writes a configuration, a sidebar, a few example pages and a public directory. Existing files are kept unless `--force` is given. `-c` selects a different configuration file name, for example `-c site.json`.

~~~text
my-docs/
  config/
    docs.json            site settings
    sidebar.json         sidebar trees
  docs/
    index.md             /
    guide/
      index.md           /guide
      getting-started.md
      configuration.md
  public/                copied to the root of the output
  dist/                  generated
~~~

## Add a page

Create a Markdown file under `docs/`:

~~~markdown
# My page

Body text.
~~~

`docs/guide/my-page.md` becomes `/guide/my-page` with `clean_urls` enabled and `/guide/my-page.html` with it disabled; the file is written to `guide/my-page/index.html` or `guide/my-page.html`. An index page keeps the directory URL `/dir` in both cases.

Add the page to `config/sidebar.json`:

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

Links use the URL the build emits. Relative links and `.md` links are resolved to that URL automatically; see [Configuration](/guide/configuration#links).

## Build

~~~bash
docs-go build
~~~

The output goes to `dist/` by default. `build` clears the directory and writes the site in place; it is not an atomic swap, so a failed render can leave a partial site. Two pages that map to the same output path are an error.

## Preview

~~~bash
docs-go dev
~~~

The dev server listens on <http://localhost:7280/>; use `-a` and `-p` to change the address and port. Each request is resolved to a source file and rendered, so edits to Markdown appear on refresh. Changes to `config/docs.json`, `config/sidebar.json` and the source directory are picked up on the next request. The browser is not reloaded automatically.

The dev server accepts the page URLs the build produces, serves assets from `src_dir` and `public_dir`, serves the theme assets under `/assets/`, and renders the 404 page for unknown paths. Markdown sources are not served directly.

## Deploy

`dist/` is an ordinary static site. With `clean_urls` enabled a page URL is extensionless and the server serves a directory index, so `/guide/start` is `guide/start/index.html`. With it disabled the links point at the generated files, so `/guide/start.html` is `guide/start.html`; an index page is `/guide`, which the server resolves to `guide/index.html`. The disabled layout needs no URL rewriting; a plain nginx configuration still needs a directory index and, for the enabled layout, a `try_files` rule.

Set `site_url` to the absolute site address to emit canonical links and, with `sitemap` enabled, `sitemap.xml`.

## Next

- [Configuration](/guide/configuration)
- [Markdown](/guide/markdown)
