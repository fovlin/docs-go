# Sidebar

The sidebar lives in `sidebar.json`, not in `docs.json`. The file name is
configurable with `sidebar_file`.

There are two accepted shapes.

## One sidebar for every page

~~~json
[
  {
    "text": "Guide",
    "items": [
      { "text": "Getting started", "link": "/guide/getting-started" },
      { "text": "Configuration", "link": "/guide/configuration" }
    ]
  }
]
~~~

An array is the simplest option for a small site.

## Sidebars by URL prefix

~~~json
{
  "/guide/": [
    {
      "text": "Guide",
      "items": [
        { "text": "Overview", "link": "/guide/" },
        { "text": "Getting started", "link": "/guide/getting-started" }
      ]
    }
  ],
  "/reference/": [
    {
      "text": "Reference",
      "items": [
        { "text": "Command line", "link": "/reference/cli" }
      ]
    }
  ]
}
~~~

For each page the **longest matching prefix** wins. `/guide/start` uses
`/guide/`; `/guide/` also uses `/guide/`. A `/` key is a catch-all. Pages with
no matching prefix render without a sidebar.

## Route and page prefixes must match

A page linked from a route section must live under that route. For example,
`/guide/` may link to `/guide/`, `/guide/start` or `/guide/start.md`, but not
to `/shared`:

~~~json
{
  "/guide/": [{ "text": "Shared", "link": "/shared" }]
}
~~~

That configuration is rejected during the build with an error, because the
page prefix does not match its sidebar parent. Moving the page to
`docs/guide/shared.md` fixes it.

## Entries

| Field | Description |
| --- | --- |
| `text` | Entry label. If missing, the linked page title is used. |
| `link` | `/guide/start`, `/guide/start.md`, `/guide/start.html`, an anchor, or an external URL. |
| `items` | Nested entries. |
| `collapsed` | Start a group closed. Groups with an active child open automatically. |

Links are site paths, without the configured `base`. External links are left
alone and get a small icon.

## Active state and ordering

- An entry is active when its resolved page equals the current page.
- A group is highlighted when a child is active, and opens automatically even
  if it is marked `collapsed`.
- Previous and next links follow the order of the selected sidebar only. Pages
  that are not in that sidebar get no previous/next links; different route
  sections are not mixed.

::: info
The sidebar is site-wide. Pages cannot opt out of it, because there is no
per-page frontmatter.
:::
