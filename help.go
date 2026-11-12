package main

const helpText = `docs-go — a small, plain static documentation generator

usage:
  docs-go build [options] [project]
  docs-go dev   [options] [project]
  docs-go init  [options] [directory]

commands:
  build      generate the site (default)
  dev        render pages on demand and preview over http
  init       write a new site skeleton

options:
  -c, --config <file>   config file (default: docs.json)
  -s, --src <dir>       override src_dir
  -o, --out <dir>       override out_dir
  -a, --addr <ip>       dev listen address (default: 0.0.0.0)
  -p, --port <port>     dev listen port (default: 8080)
  -f, --force           overwrite scaffold files in init
  -q, --quiet           quiet output

examples:
  docs-go init my-docs
  cd my-docs && docs-go build
  docs-go dev -p 9000
  docs-go build -c docs.json .
`
