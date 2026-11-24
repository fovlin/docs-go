# Changelog

## 1.1.0

- Markdown containers close only at block boundaries, so fences and nested containers inside list items are preserved; an indented closing `:::` that belongs to list content is text.
- Markdown loose lists wrap every direct item paragraph in `<p>`; tight lists keep the first paragraph inline.
- Markdown list continuation and code indentation honour tab stops when stripping indentation.
- An unclosed fenced code block inside a container keeps its trailing blank lines, matching the behaviour at the top level.
- Page links in Markdown (`.md`, `.html`, extensionless) are rewritten to the page URL of the current layout; anchors, queries, images and other file extensions are left alone.
- `nav[].active_match` is honoured: a nav entry is active when the page URL starts with the given prefix.
- Previous and next links include a sidebar group’s own link when it has one.
- Sidebar group arrows are buttons beside the link and toggle the group; groups with a link still navigate from the label.
- The palette is warm: every colour is a root variable, and the browser theme colour follows the page background.
- The theme toggle locks an explicit choice and returns to following the system theme when the chosen theme matches it.
- Code blocks no longer render a language label; the copy button is hidden until the block is hovered or the button is focused.
- Code blocks span the full width on narrow screens.
- The outline heading links back to the top of the page.
- Documentation: added a guide overview and a Markdown reference; the configuration reference covers every field.
- Removed the undocumented `serve` alias; use `dev`.
- Fixed `init` with an absolute path printing an empty location.
- `--help` now lists `-t`, `--color`, `-v` and `-h`.
