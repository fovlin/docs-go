# Markdown

The parser implements a subset of Markdown. Unsupported syntax is not an error; it usually renders as plain text.

## Supported

- ATX headings (`#` through `######`). Heading IDs are generated from the text; duplicate headings receive a numeric suffix.
- Paragraphs, entity references, inline code, emphasis and strong emphasis.
- Links and images, including titles. Links nested inside links are parsed and produce nested anchors.
- Fenced code blocks (backticks or tildes), indented code blocks, blockquotes and thematic breaks.
- Ordered and unordered lists, including nested lists. Blank lines do not end a list; loose lists wrap item paragraphs in `<p>`.
- GFM-style tables with column alignment.
- `::: tip` style containers named `tip`, `info`, `note`, `warning`, `danger` or `details`, including nested containers. A closing `:::` is recognized at a block boundary, so markers inside fences or list content do not close the container.
- Inline HTML tags inside paragraphs.

## Not supported

- Setext headings.
- Reference links and link definitions.
- Footnotes and task lists.
- Strikethrough and autolinks.
- Custom container names.
- Explicit heading IDs.
- Block-level HTML.

## Page metadata

There is no frontmatter. The first level-1 heading is the page title and the first paragraph is the meta description. When either is missing the file name and the site description are used.
