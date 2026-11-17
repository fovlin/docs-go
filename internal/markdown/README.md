# docs-go Markdown parser

## Paragraphs

Consecutive lines render as one paragraph with `<br>` between lines; a blank
line starts a new paragraph. Explicit breaks (two trailing spaces or a trailing
backslash) are supported but render the same `<br>`.

```md
first line
second line

new paragraph
```

## Headings

ATX headings from `#` to `######`; a space after the hashes is required.

```md
# Heading 1

## Heading 2

### Heading 3
```

## Bold and italic

Use `**` for bold, `*` for italic and `***` for both.

```md
This is **bold**, *italic* and ***both***.
```

## Inline code

Wrap code in backticks.

```md
Use `docs-go dev` to preview.
```

## Links and images

Inline links and images are supported.

```md
[Getting started](/guide/getting-started)

![Logo](/logo.svg)
```

## Blockquotes

Start quoted lines with `>`; blank lines separate quotes.

```md
> A quoted paragraph.

> Another quote.
```

## Lists

Bullets may be `-`, `*` or `+`; ordered markers may use `.` or `)`.

```md
- First item
- Second item
  - Nested item
```

```md
1. First step
2. Second step
```

## Code blocks

Use backtick or tilde fences, or indent code by four spaces.

~~~~md
```go
fmt.Println("hello")
```
~~~~

```md
    indented code
```

## Tables

A header, delimiter and body rows; colons set alignment.

```md
| Left | Center | Right |
| :--- | :----: | ---: |
| a    | b      | c    |
```

## Containers

Use `tip`, `info`, `note`, `warning`, `danger` or `details`.

```md
::: tip
Helpful information.
:::

::: details
Hidden content.
:::
```

## Thematic breaks

Three or more `*`, `-` or `_` characters on their own line.

```md
---

***
```

## Escapes and entities

Backslash escapes punctuation; HTML entities pass through.

```md
\*not italic\*

A &amp; B
```

## Inline HTML

Inline tags pass through inside a paragraph.

```md
Press <kbd>Ctrl</kbd> + <kbd>C</kbd>.
```

## Differences from CommonMark

- Every line break inside a paragraph renders as `<br>`; CommonMark renders
  a plain newline as a space.
- Indented code can interrupt a paragraph.
- A blank line starts a new block; there is no lazy blockquote continuation.
- Tables split cells before inline parsing, so `|` inside code spans is not protected.
- Emphasis matching is greedy and does not implement CommonMark's delimiter-stack rules;
  malformed delimiter runs degrade to text.
- Only ATX headings are supported; no setext headings, autolinks, reference links,
  footnotes or frontmatter.
- Raw HTML is trusted content and is passed through unchanged.
