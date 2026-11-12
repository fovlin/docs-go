# Command line

~~~text
docs-go <command> [options] [project]
~~~

The project directory defaults to the current directory. Flags may be placed
after the command, before the project path.

## Commands

| Command | Description |
| --- | --- |
| `build` | Generate the site. |
| `dev` | Render pages on demand and preview over HTTP. |
| `init` | Write a new site skeleton. |

## Options

| Option | Description |
| --- | --- |
| `-c, --config <file>` | Config file, default `docs.json`. |
| `-s, --src <dir>` | Override `src_dir`. |
| `-o, --out <dir>` | Override `out_dir`. |
| `-a, --addr <ip>` | Listen address for `dev`, default `0.0.0.0`. |
| `-p, --port <port>` | Listen port for `dev`, default `8080`. |
| `-f, --force` | Overwrite scaffold files in `init`. |
| `-q, --quiet` | Do not print progress. |
| `-v, --version` | Print the version. |
| `-h, --help` | Print help. |

## Examples

~~~bash
docs-go init my-docs
cd my-docs
docs-go dev -p 9000
docs-go build -q
docs-go build -c config/docs.json -o site
~~~

## Preview server

The server:

- resolves clean URLs and directory indexes;
- renders a Markdown page when it is requested;
- serves assets from `public/` and `docs/`, plus the embedded theme assets;
- serves a rendered 404 page for missing pages;
- reloads `docs.json` and `sidebar.json` when they change.

Pages are resolved and read from the source tree for each request, so editing a
Markdown file is visible after refreshing the browser. There is no background
source-tree scan.
