# docs-go

A static documentation generator with one job: turn a directory of markdown
files into a fast, readable documentation site. It follows the same visual and
code style as [next-index](https://github.com/fovlin/next-index), and keeps the
sidebar in its own JSON file.

::: tip
No Node.js, no bundler, no plugin graph. One Go binary, a `docs.json`, a
`sidebar.json` and a folder of markdown.
:::

## Quick start

~~~bash
go install acovia.net/docs-go@latest
docs-go init my-docs
cd my-docs
docs-go dev
~~~

The dev server renders pages on demand. Edit a page and refresh the browser;
the requested source file is read again. Config and sidebar changes apply on
the next request.

## What you get

- A VitePress-like layout: top navigation, left sidebar, right outline,
  previous/next links, an edit link and a last-updated date.
- A native, zero-dependency Markdown parser: headings, tables, task lists,
  strikethrough, autolinks, heading anchors, custom containers and `details`
  blocks.
- Plain `<pre><code>` blocks with a copy button. No syntax highlighting, no
  line numbers and no code themes.
- A light and dark theme that follows the system. The toggle stores the reader's choice; toggling back to the system theme returns to automatic mode.
- Clean URLs, an automatic `404.html`, an optional sitemap and a generated
  output directory that is cleared on each build.
- Markdown links to `.md` files are rewritten to the generated page, and broken
  internal links are reported as warnings.

## The two config files

`docs.json` holds site settings. `sidebar.json` holds navigation. They are
separate on purpose: the sidebar changes often and its shape is different from
the rest of the settings.

~~~json
{
  "title": "My docs",
  "src_dir": "docs",
  "out_dir": "dist",
  "sidebar_file": "sidebar.json",
  "clean_urls": true,
  "nav": [{ "text": "Guide", "link": "/guide/" }]
}
~~~

~~~json
{
  "/guide/": [
    {
      "text": "Guide",
      "items": [
        { "text": "Getting started", "link": "/guide/getting-started" }
      ]
    }
  ]
}
~~~

Continue with [getting started](/guide/getting-started) or read the full
[configuration reference](/guide/configuration).
