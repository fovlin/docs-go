# Configuration

A docs-go project uses two JSON files:

- `config/docs.json` holds site settings, navigation and output paths.
- `config/sidebar.json` holds one or more sidebar trees.

Every field is optional. Paths are resolved against the project directory; absolute paths are used as they are.

## config/docs.json

~~~json
{
  "title": "My Docs",
  "description": "Documentation for my project.",
  "lang": "en",
  "src_dir": "docs",
  "out_dir": "dist",
  "public_dir": "public",
  "clean_urls": true,
  "sidebar_file": "config/sidebar.json",
  "site_url": "https://example.com/docs",
  "sitemap": true,
  "logo": "/logo.svg",
  "favicon": "/favicon.svg",
  "appearance": "auto",
  "nav": [
    { "text": "Guide", "link": "/guide/" },
    { "text": "GitHub", "link": "https://github.com/me" }
  ],
  "social_links": [
    { "icon": "github", "link": "https://github.com/me", "label": "GitHub" }
  ],
  "outline": { "label": "On this page", "level": [2, 3] },
  "edit_link": "https://github.com/me/docs/edit/main/docs/:path",
  "last_updated": true,
  "prev_next": true,
  "prev_text": "Previous page",
  "next_text": "Next page",
  "ui": { "menu": "Menu" }
}
~~~

### Core fields

| Field | Default | Description |
| --- | --- | --- |
| `title` | `DOCS_GO` | Site title. |
| `description` | `Documentation built with docs-go.` | Default meta description. |
| `lang` | `en` | Document language. |
| `src_dir` | `docs` | Markdown and asset source directory. |
| `out_dir` | `dist` | Build output directory. |
| `public_dir` | `public` | Files copied to the output root. |
| `clean_urls` | `true` | Write non-index pages as `<name>/index.html` and drop the extension from page URLs. With `false` pages are written as `<name>.html` and linked by that file name. An index page uses the directory URL `/dir` in both cases. |
| `sidebar_file` | `config/sidebar.json` | Sidebar file. |
| `site_url` | empty | Absolute site address, used for canonical links and the sitemap. Must start with `http://` or `https://`. |
| `sitemap` | `false` | Write `sitemap.xml`. Requires `site_url`. |

### Header

| Field | Default | Description |
| --- | --- | --- |
| `logo` | `/docs-go.svg` | Image URL in the topbar. A site path is resolved under `public_dir`; an absolute URL is used as is. The default file is written by `init`. |
| `favicon` | empty | Favicon URL. When empty the built-in favicon is served from `/favicon.ico`. |
| `appearance` | `auto` | Initial theme: `auto`, `light` or `dark`. Other values fall back to `auto`. |

### Navigation

`nav` is an array of topbar links:

~~~json
"nav": [
  { "text": "Guide", "link": "/guide/", "active_match": "/guide" },
  { "text": "GitHub", "link": "https://github.com/me" }
]
~~~

`text` falls back to `link` when omitted. Internal links are written without an extension; an index page is linked by its directory, such as `/guide/`. External links are left unchanged. A link is marked active when its normalized href equals the current page URL, or when `active_match` is set and the page URL starts with it.

### Social links

`social_links` is an array of `{ "icon", "link", "label" }` objects. `label` falls back to the icon name. The icon names are `github`, `gitlab`, `x`, `twitter`, `discord`, `youtube`, `linkedin`, `rss`, `mail`, `email`, `bilibili` and `link`. Unknown names use the `link` icon. Social links are hidden at 1240px and below and are moved into the page menu.

### Outline

| Field | Default | Description |
| --- | --- | --- |
| `outline.label` | `On this page` | Outline heading. |
| `outline.level` | `[2, 3]` | Heading levels included in the outline. Values outside 1–6 are ignored; an empty list uses the default. |

### Page chrome

| Field | Default | Description |
| --- | --- | --- |
| `edit_link` | empty | Edit URL template. `:path` is the source path and `:url` the page URL; both are escaped. |
| `last_updated` | `true` | Show the source modification date. |
| `prev_next` | `true` | Show previous and next links. |
| `prev_text` | `Previous page` | Label used before the previous page title. |
| `next_text` | `Next page` | Label used before the next page title. |

### UI strings

`ui` overrides interface labels:

| Key | Default |
| --- | --- |
| `skip_to_content` | `Skip to content` |
| `menu` | `Menu` |
| `theme` | `Toggle theme` |
| `edit_page` | `Edit this page` |
| `last_updated` | `Last updated` |
| `not_found_title` | `Page not found` |
| `not_found_text` | `The page you are looking for does not exist or has moved.` |
| `back_home` | `Back to home` |
| `copy` | `Copy` |
| `copied` | `Copied` |

## config/sidebar.json

The sidebar accepts one array for every page:

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
      "link": "/guide/",
      "items": [
        { "text": "Getting started", "link": "/guide/getting-started" },
        { "text": "Configuration", "link": "/guide/configuration" }
      ]
    }
  ]
}
~~~

Keys are normalized by trimming slashes. The longest matching prefix selects the tree for a page; two keys that normalize to the same prefix are an error.

An item has these fields:

| Field | Description |
| --- | --- |
| `text` | Entry label. |
| `link` | Target page or external URL. Omitted for a group that only toggles. |
| `collapsed` | When `true` the group starts closed. Default `false`. |
| `items` | Child entries. |

A group with a `link` renders the link and an arrow inside it; a group without a `link` toggles its children. A group is open when it is not collapsed, when its own link is the current page, or when a descendant is the current page. Links are written without an extension and normalized like other site links; an index page is linked by its directory.

Previous and next links follow the leaf entries of the current sidebar tree in order. Group entries with children are skipped.

## Page URLs

| Source | `clean_urls: true` | `clean_urls: false` |
| --- | --- | --- |
| `index.md` | `/` → `index.html` | `/` → `index.html` |
| `guide/index.md` | `/guide` → `guide/index.html` | `/guide` → `guide/index.html` |
| `guide/start.md` | `/guide/start` → `guide/start/index.html` | `/guide/start.html` → `guide/start.html` |

With `clean_urls` enabled a page URL is its source path without the extension. With it disabled a non-index page is linked by the file it generates, while an index page keeps the directory URL `/dir` in both cases.

## Links

Link destinations are normalized when a page is rendered:

- destinations with a scheme or host (`http`, `https`, `mailto`, `//`) and in-page anchors (`#`) are written unchanged;
- other destinations are site paths: a leading `/` is expected, a `.md` or `.html` extension is removed, and the result is converted to the page URL;
- a trailing `/` is removed, so an index page is linked as `/guide/` and renders as `/guide`;
- query strings and fragments are preserved.

Links are not validated. Configuration links in `nav` and `sidebar.json` are normalized the same way.

## Sitemap and canonical links

`site_url` is the absolute site address. When it is set every page gets a canonical link. `sitemap` writes `sitemap.xml` and requires `site_url`.

## Validation

The build fails when:

- `site_url` is set but is not an absolute `http://` or `https://` URL;
- `sitemap` is enabled without `site_url`;
- two sidebar keys normalize to the same prefix;
- two pages produce the same output path;
- `out_dir` contains the project root or overlaps `src_dir` or `public_dir`.

A `footer` object from older configurations is ignored.
