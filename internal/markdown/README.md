# docs-go Markdown parser

The docs-go parser supports the small Markdown subset below. Syntax outside
it usually renders as plain text.

## Headings

```text
# Heading level 1

## Heading level 2

### Heading level 3

#### Heading level 4

##### Heading level 5

###### Heading level 6
```

ATX headings only (`#` to `######`). A space or tab after the hashes is
required. Heading anchors are generated for the site outline; duplicate
headings get `-1`, `-2` suffixes.

## Paragraphs

```text
A paragraph is one or more lines of text.

A blank line separates two paragraphs.
```

A block that starts with four spaces or a tab is code, not a paragraph.
Two trailing spaces or a trailing backslash create a hard line break.

## Bold and italic

```text
This is a **bold** font.

This is an *italic* font.

This is a ***bold and italic*** font.
```

Underscores work the same way: `_italic_`, `__bold__` and `___both___`.

## Blockquotes

```text
> This is a quote.

> Dorothy followed her through many of the beautiful rooms in her castle.
>
> The Witch bade her clean the pots and kettles and sweep the floor and keep the fire fed with wood.
```

A blank line starts a new blockquote. Put `>` on an empty line to keep
multiple paragraphs in one quote. Quotes can be nested with `> >`.

## Lists

Bullets may be `-`, `*` or `+`; ordered markers may use `.` or `)`. A different
marker starts a new list. The first number sets the ordered list start; later
numbers are ignored.

### Ordered list

```text
1. First item
2. Second item
3. Third item
    1. Indented item
    2. Indented item
4. Fourth item
```

Result:

```text
<ol>
<li>First item</li>
<li>Second item</li>
<li>Third item
<ol>
<li>Indented item</li>
<li>Indented item</li>
</ol>
</li>
<li>Fourth item</li>
</ol>
```

### Unordered list

```text
- First item
- Second item
- Third item
    - Indented item
    - Indented item
- Fourth item
```

Result:

```text
<ul>
<li>First item</li>
<li>Second item</li>
<li>Third item
<ul>
<li>Indented item</li>
<li>Indented item</li>
</ul>
</li>
<li>Fourth item</li>
</ul>
```

Blank lines do not end a list; they stay inside the list context. There are no
loose/tight items and no lazy continuation lines. Nested lists are recognized
at two spaces from the marker, or deeper:

```text
- Parent
  - Child
```

## Containers

Six callout types are available: `tip`, `info`, `note`, `warning`, `danger` and
`details`. An optional title follows the type, containers can be nested, and a
line containing only `:::` closes one:

```text
::: tip
This is a tip.
:::

::: warning Custom title
Container content.
:::
```

`::: details` renders as a collapsible `<details>` block. Unknown container
names and `:::` lines inside fenced code are left as text.

## Tables

Tables use a header row, a delimiter row and body rows. Leading and trailing
pipes are optional, but the header must contain a pipe. Colons set column
alignment:

```text
| Left | Center | Right |
| :--- | :----: | ---: |
| a    | b      | c    |
```

`:---` aligns left, `:---:` centers and `---:` aligns right. The header and
delimiter rows must have the same number of cells. Delimiter cells need at
least three dashes. Use `\|` for a literal pipe inside a cell. Code spans do not protect pipes.
Body rows may have fewer cells (padded with empty cells) or more cells (the
extras are ignored).

Result:

```text
<table>
<thead>
<tr>
<th style="text-align:left">Left</th>
<th style="text-align:center">Center</th>
<th style="text-align:right">Right</th>
</tr>
</thead>
<tbody>
<tr>
<td style="text-align:left">a</td>
<td style="text-align:center">b</td>
<td style="text-align:right">c</td>
</tr>
</tbody>
</table>
```

## Code

### Code blocks

Fenced code uses at least three backticks or three tildes and an optional
language:

````text
```go
// This is a code block
package main

import "fmt"

func main() {
    fmt.Println("Hello docs-go")
}
```
````

Indented code uses four spaces or a tab:

```text
    // This is a code block
    package main

    import "fmt"

    func main() {
        fmt.Println("Hello docs-go")
    }
```

Result:

```go
// This is a code block
package main

import "fmt"

func main() {
    fmt.Println("Hello docs-go")
}
```

### Inline code

```text
This is an `inline code` example.
```

Result:

This is an `inline code` example.

## Thematic breaks

Three or more `*`, `-` or `_` characters on their own line make a thematic
break. Leave a blank line around it:

```text
***

---
```

Result:

***

---

## Links and images

Inline links and images only. An optional title is allowed:
`[text](url "title")`. Reference links and autolinks are not supported.

```text
My GitHub account: [fovlin](https://github.com/fovlin)
```

Result:

My GitHub account: [fovlin](https://github.com/fovlin)

```text
![Picture](https://sufox.net/sufox.png)
```

Result:

![Picture](https://sufox.net/sufox.png)

## Escapes

A backslash makes the following ASCII punctuation character literal. HTML
entities such as `&amp;` pass through.

```text
\-\-\-
```

Result:

\-\-\-

## Inline HTML

Inline tags such as `<kbd>`, `<br>`, `<span>` and `<sup>` pass through inside
a paragraph, and Markdown around them keeps working. Block-level HTML such as
`<div>`, `<table>` or `<details>` is not supported; use Markdown blocks
instead.

## Not supported

Setext headings, reference links, footnotes, frontmatter, task lists,
strikethrough, autolinks, custom container names, explicit heading IDs and
block-level HTML.
