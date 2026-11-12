# Getting started

Install docs-go, create a project and preview it locally.

## Install

Requires Go 1.27.2 or newer.

~~~bash
go install acovia.net/docs-go@latest
~~~

Or build this checkout:

~~~bash
git clone https://github.com/fovlin/docs-go
cd docs-go
go build .
~~~

## Create a project

~~~bash
docs-go init my-docs
cd my-docs
~~~

`init` writes a small, working site:

~~~text
my-docs/
  docs.json          site settings
  sidebar.json       navigation, kept separate
  docs/
    index.md         /
    guide/
      index.md       /guide/
      getting-started.md
      configuration.md
  public/            copied to the root of the output
~~~

## Preview

~~~bash
docs-go dev
~~~

The dev server reads `docs.json` and `sidebar.json` at startup, then resolves
and renders each Markdown page when it is requested. Page bodies are read from
disk on every request, so refresh the browser after editing a page. Config and
sidebar changes are picked up on the next request; there is no background disk
scan.

Open <http://localhost:8080/>.

## Build

~~~bash
docs-go build
~~~

The result is written to `dist/` (configurable with `out_dir`). The directory
is cleared first, then the site is written directly. A failed build can leave a
partial output tree; remove `dist/` and rebuild if that happens.

::: warning
`dist/` is generated. Edit files in `docs/` and `public/`, not in `dist/`.
:::

## Add a page

1. Create `docs/guide/my-page.md`.
2. Add an entry to `sidebar.json`.
3. The URL is `/guide/my-page` when `clean_urls` is enabled.

Links can use any of these forms; they all resolve to the same page:

~~~markdown
[My page](/guide/my-page)
[My page](/guide/my-page.md)
[My page](/guide/my-page.html)
~~~

The first form is the canonical one.

## Titles and descriptions

There is no frontmatter. The first `# heading` is the page title. If a page has
no H1, the file name becomes the title. The first paragraph is used as the meta
description; if a page has no paragraph, the site description is used.

Page-level switches do not exist. Outline levels, last-updated dates,
edit links and previous/next links are all controlled from `docs.json`, and the
sidebar comes from `sidebar.json`.
