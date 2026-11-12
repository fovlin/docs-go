# Configuration

`docs.json` is loaded from the project directory, or from the file passed with
`-c`. Unknown fields are ignored; missing fields fall back to sensible defaults.

## A complete example

~~~json
{
  "title": "My docs",
  "description": "Documentation for My project.",
  "lang": "en",
  "base": "/",
  "src_dir": "docs",
  "out_dir": "dist",
  "public_dir": "public",
  "clean_urls": true,
  "site_url": "https://example.com",
  "sidebar_file": "sidebar.json",
  "sitemap": true,
  "logo": "/logo.svg",
  "favicon": "/favicon.svg",
  "appearance": "auto",
  "nav": [
    { "text": "Guide", "link": "/guide/" }
  ],
  "social_links": [
    { "icon": "github", "link": "https://github.com/you/project", "label": "GitHub" }
  ],
  "footer": {
    "message": "Released under the MIT license.",
    "copyright": "Copyright © {year} You"
  },
  "outline": { "label": "On this page", "level": [2, 3] },
  "edit_link": "https://github.com/you/project/edit/main/docs/:path",
  "last_updated": true,
  "prev_next": true,
  "prev_text": "Previous page",
  "next_text": "Next page"
}
~~~

## Site

| Field | Default | Description |
| --- | --- | --- |
| `title` | `DOCS_GO` | Site name in the top bar, page titles and footer. |
| `description` | — | Default meta description. A page with a first paragraph uses that paragraph instead. |
| `lang` | `en` | Value of the `lang` attribute on every page. |
| `base` | `/` | URL path the site is served from, such as `/docs/`. |
| `src_dir` | `docs` | Markdown source directory. |
| `out_dir` | `dist` | Generated site directory. Cleared and regenerated on each build; must not overlap `src_dir`, `public_dir` or the project root. |
| `public_dir` | `public` | Files copied to the root of the output after the build. |
| `clean_urls` | `true` | Link `/guide/start` instead of `/guide/start.html`. |
| `site_url` | — | Absolute origin used for canonical URLs and the sitemap. |
| `sitemap` | `false` | Write `sitemap.xml`. Requires `site_url`. |

Relative paths are resolved from the project directory, not from the config
file.

## Navigation and appearance

| Field | Description |
| --- | --- |
| `nav` | Top bar links: `text`, `link`, optional `active_match`. |
| `social_links` | `icon` (`github`, `gitlab`, `x`, `discord`, `youtube`, `linkedin`, `rss`, `mail`), `link`, optional `label`. |
| `logo` | Image shown left of the title. Empty uses the built-in mark. |
| `favicon` | Optional favicon URL. Otherwise the bundled `theme/favicon.ico` is written to the site root. |
| `appearance` | `auto` follows the system. `light` or `dark` forces the initial theme; the reader can still switch, and with `auto` toggling back to the system theme restores automatic mode. |
| `footer.message` | Optional line above the copyright. |
| `footer.copyright` | `{year}` is replaced at build time. |

## Reading experience

| Field | Default | Description |
| --- | --- | --- |
| `outline.label` | `On this page` | Heading of the right-hand outline. |
| `outline.level` | `[2, 3]` | Heading levels included in the outline. |
| `edit_link` | — | Template with `:path` (source path) and `:url`. |
| `last_updated` | `true` | Show the source file modification date. |
| `prev_next` | `true` | Show previous/next links from the sidebar order. |
| `prev_text`, `next_text` | — | Labels for the previous/next links. |

## Containers

| Field | Description |
| --- | --- |
| `containers` | Labels for `tip`, `info`, `note`, `warning`, `danger` and `details`. |

## Interface labels

Every built-in label can be translated through `ui`:

~~~json
{
  "ui": {
    "skip_to_content": "跳到正文",
    "menu": "菜单",
    "theme": "切换主题",
    "edit_page": "编辑此页",
    "last_updated": "最后更新",
    "not_found_title": "页面不存在",
    "not_found_text": "你访问的页面不存在或已被移动。",
    "back_home": "返回首页",
    "copy": "复制",
    "copied": "已复制"
  }
}
~~~

## Per-page metadata

There is no frontmatter. The first `# heading` is the page title, the first
paragraph is the meta description, and outline, edit links, dates and
previous/next links are site-wide settings.

## Deploying under a sub-path

Set `base` to the public path and `site_url` to the origin:

~~~json
{
  "base": "/docs/",
  "site_url": "https://example.com"
}
~~~

Every generated link and asset URL is prefixed with `/docs/`. The preview
server understands the prefix too.

::: tip
If your host cannot serve extensionless URLs, set `"clean_urls": false`. The
generator then writes and links `page.html` directly.
:::
