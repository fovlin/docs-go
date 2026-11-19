# Changelog

## 1.1.0

- Page links in Markdown (`.md`, `.html`, extensionless) are rewritten to the page URL of the current layout; anchors, queries, images and other file extensions are left alone.
- `nav[].active_match` is honoured: a nav entry is active when the page URL starts with the given prefix.
- Previous and next links skip sidebar group entries and follow leaf entries only.
- Sidebar group links and the group arrow render as a single link row.
- Documentation: added a guide overview and a Markdown reference; the configuration reference covers every field.
- Removed the undocumented `serve` alias; use `dev`.
- Fixed `init` with an absolute path printing an empty location.
- `--help` now lists `-t`, `--color`, `-v` and `-h`.
