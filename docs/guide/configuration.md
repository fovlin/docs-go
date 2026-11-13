# Configuration

The site is configured with two JSON files.

## docs.json

~~~json
{
  "title": "My Docs",
  "src_dir": "docs",
  "out_dir": "dist",
  "sidebar_file": "sidebar.json",
  "clean_urls": true,
  "nav": [{ "text": "Guide", "link": "/guide/" }]
}
~~~

The main config holds the title, description, language, base path, navigation,
theme switches and output settings.

## sidebar.json

The sidebar is deliberately a separate file. It accepts either one array for
every page:

~~~json
[
  {
    "text": "Guide",
    "items": [{ "text": "Getting started", "link": "/guide/getting-started" }]
  }
]
~~~

or an object keyed by URL prefix:

~~~json
{
  "/guide/": [
    {
      "text": "Guide",
      "items": [{ "text": "Getting started", "link": "/guide/getting-started" }]
    }
  ]
}
~~~

The longest matching prefix wins. Groups can be nested and collapsed with
`"collapsed": true`.
