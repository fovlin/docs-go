# Configuration

A docs-go site uses two JSON files:

- `config/docs.json` for site settings, navigation and output paths
- `config/sidebar.json` for one or more sidebar trees

Every field is optional; missing fields use defaults.

## config/docs.json

~~~json
{
  "title": "My Docs",
  "description": "Documentation for my project.",
  "lang": "en",
  "src_dir": "docs",
  "out_dir": "dist",
  "public_dir": "public",
  "sidebar_file": "config/sidebar.json",
  "clean_urls": true,
  "logo": "/logo.svg",
  "nav": [
    { "text": "Guide", "link": "/guide/getting-started" }
  ],
  "social_links": [
    { "icon": "github", "link": "https://github.com/me", "label": "GitHub" }
  ],
  "outline": { "label": "On this page", "level": [2, 3] },
  "prev_next": true,
  "last_updated": true
}
~~~

## Core fields

| Field | Default | Description |
| --- | --- | --- |
| `title` | `DOCS_GO` | Site title. |
| `description` | `Documentation built with docs-go.` | Default meta description. |
| `lang` | `en` | Document language. |
| `src_dir` | `docs` | Markdown and asset source directory. |
| `out_dir` | `dist` | Build output directory. |
| `public_dir` | `public` | Files copied to the output root. |
| `sidebar_file` | `config/sidebar.json` | Sidebar file, relative to the project. |
| `clean_urls` | `true` | Omit `.html` from generated links. |
| `site_url` | empty | Absolute URL used for canonical links and the sitemap. |
| `sitemap` | `false` | Write `sitemap.xml`; requires `site_url`. |

## Header

| Field | Default | Description |
| --- | --- | --- |
| `logo` | empty | Image URL in the topbar. Omitting it renders no `<img>`. |
| `favicon` | empty | Favicon URL. |
| `appearance` | `auto` | Initial theme: `auto`, `light` or `dark`. |

`nav` is an array of topbar links:

~~~json
"nav": [
  { "text": "Guide", "link": "/guide/" },
  { "text": "GitHub", "link": "https://github.com/me" }
]
~~~

`active_match` can override the URL prefix used for the active state.

`social_links` accepts `github`, `gitlab`, `x`, `twitter`, `discord`, `youtube`, `linkedin`, `rss`, `mail`, `email` and `link`. Unknown icons fall back to the generic link icon. Social links are hidden at `1240px` and below; they move into the **On this page** menu together with the outline.

## Page chrome

| Field | Default | Description |
| --- | --- | --- |
| `edit_link` | empty | Edit URL template. `:path` is the source path; `:url` is the page URL. |
| `last_updated` | `true` | Show the source modification date. |
| `prev_next` | `true` | Show previous/next links. |
| `prev_text` | `Previous page` | Previous label. |
| `next_text` | `Next page` | Next label. |
| `outline.label` | `On this page` | Outline heading. |
| `outline.level` | `[2, 3]` | Heading levels included in the outline. |

## UI strings

Override interface labels with `ui`:

~~~json
"ui": {
  "menu": "Menu",
  "theme": "Toggle theme",
  "edit_page": "Edit this page",
  "last_updated": "Last updated",
  "not_found_title": "Page not found",
  "not_found_text": "The page you are looking for does not exist.",
  "back_home": "Back to home",
  "copy": "Copy",
  "copied": "Copied"
}
~~~

## config/sidebar.json

The sidebar is separate from `config/docs.json`. It accepts one array for every page:

~~~json
[
  {
    "text": "Guide",
    "items": [
      { "text": "Getting started", "link": "/guide/getting-started" }
    ]
  }
]
~~~

or an object keyed by URL prefix:

~~~json
{
  "/guide/": [
    {
      "text": "Guide",
      "items": [
        { "text": "Getting started", "link": "/guide/getting-started" },
        { "text": "Configuration", "link": "/guide/configuration" }
      ]
    }
  ]
}
~~~

The longest matching prefix wins. Sidebars may link to pages in other sidebars; the destination page uses its own matching sidebar and its own prev/next context.

## No footer

There is no footer configuration. A `footer` object in an older `config/docs.json` is ignored.
