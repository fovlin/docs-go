# Markdown

docs-go uses its own small Markdown parser. It is written against the Go
standard library only, so there is no Goldmark, Chroma or YAML in the build.

The parser is deliberately a documentation subset, not a full CommonMark
implementation. Unsupported syntax falls back to plain text instead of failing
the build.

## Headings and outline

```markdown
## A section
### A subsection
```

Every heading gets an id and a hover anchor. Chinese headings work too. The
outline shows levels from `outline.level` in `docs.json` (default `2` and `3`).
There is no per-page outline override.

## Code

Fence a block with a language:

~~~markdown
```go
func main() {
    fmt.Println("hello")
}
```
~~~

The language appears in the corner and a copy button is added at runtime. Code
is escaped and rendered as plain `<pre><code>`; there is no syntax
highlighting, no line numbering and no code theme.

## Custom containers

```markdown
::: tip
A helpful hint.
:::

::: warning Watch out
The title can be customized.
:::

::: details More
Hidden until opened.
:::
```

::: tip
Containers can be nested and may contain any markdown.
:::

::: warning
The title is optional. Missing titles use the labels in `containers`.
:::

::: details What about code inside?
The parser leaves fenced code alone, so a `:::` line inside a fence is safe.
:::

## Tables, tasks and inline syntax

| Column | Meaning |
| --- | --- |
| `text` | Label |
| `link` | Destination |

- [x] Task lists
- [x] Strikethrough
- [x] Autolinks such as https://example.com

Inline syntax includes `code`, **strong**, *emphasis*, ~~strikethrough~~,
links and images. Bare `http://` and `https://` URLs become links.

Footnotes, reference-style links, setext headings and indented code blocks are
not part of the subset.

## Links and images

Relative markdown links are resolved from the current page:

```markdown
[Configuration](./configuration.md)
[Overview](../guide/)
```

They become clean site links, with the configured base applied. If a `.md`
target does not exist, the build logs a warning and keeps the link usable.

Images work the same way:

```markdown
![Diagram](./diagram.png)
```

Put files next to the markdown to have them copied to the same path, or use an
absolute path such as `/logo.svg` for a file in `public/`.

## Raw HTML

Raw HTML is not passed through. The source is escaped and shown as text, so a
`<figure>` or `<script>` tag cannot become live HTML.

If you need a figure, use an image followed by a caption paragraph, or a
`::: details` container.
