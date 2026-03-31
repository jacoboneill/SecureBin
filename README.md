# SecureBin

[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue)](LICENSE) [![Go](https://img.shields.io/badge/Go-1.25-%2300ADD8?logo=go&logoColor=white)](#) [![HTMX](https://img.shields.io/badge/HTMX-4.0.0--alpha7-36C?logo=htmx&logoColor=fff)](#) [![SQLite](https://img.shields.io/badge/SQLite-3.51.3-%2307405e?logo=sqlite&logoColor=white)](#)

An end-to-end encrypted pastebin. Content is encrypted client-side using
AES-256-GCM via the Web Crypto API before it ever leaves the browser. The
server only stores ciphertext and never sees plaintext, paste keys, or
master keys.

## How It Works

Each paste gets its own randomly generated encryption key. That key travels
in the URL fragment (the part after `#`), which
[is never sent to the server](https://www.rfc-editor.org/rfc/rfc3986#section-3.5).
Registered users can also access their pastes via a master key derived from
their password using PBKDF2, so pastes remain accessible even without the
original link.

## Prerequisites

- [Go](https://go.dev/) 1.25+

## Getting Started

```sh
git clone https://github.com/jacoboneill/SecureBin.git
cd SecureBin
go run ./cmd/server
```

## Project Structure

```
├── cmd/server/main.go          # Application entry point
├── Dockerfile                  # Multi-stage production build
├── docs/                       # Project documentation
├── internal/
│   ├── db/
│   │   ├── migrations/         # DB table schemas, used by golang-migrate
│   │   └── queries/            # SQLC query definitions (source of truth for internal/db/)
│   ├── contextkeys/            # typed context key constants
│   ├── templates/              # templ components and generated Go code
│   ├── testutil/               # utilities for tests
│   └── handlers/               # HTTP handlers
├── static/                     # Client-side assets (JS, images)
├── sqlc.yaml                   # SQLC configuration
├── go.mod
└── go.sum
```

## Documentation

- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) -- project design and
  handler conventions
- [`docs/API.md`](docs/API.md) -- full route reference

## License

[AGPL-3.0](LICENSE)
