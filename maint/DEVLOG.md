# Dev Log / Developer Guide

This document is a short onboarding guide for engineers who are new to this repository.
It explains how the project is structured, how to run it locally, and where to make changes.

## What is tinyenv?

`tinyenv` is a small, self-contained CLI that manages language runtimes in a style similar to `*env` tools
(e.g., rbenv, pyenv, goenv). It downloads prebuilt archives for multiple languages, installs them under a
common root, and generates lightweight shims into a shared `bin` directory.

## Quick Start (Local)

- Build and run with Go:

```bash
# runs the CLI using the local main package
$ go run . --help

# convenience wrapper (sets TINYENV_ROOT to ./_root)
$ ./test.sh versions
```

- Local root selection:
  - If `TINYENV_ROOT` is set, it is used.
  - Otherwise, the root is inferred from the installed binary path.

## Repository Layout

- `main.go`: CLI entrypoint and command wiring.
- `language/`: per-language logic and shared helpers.
  - `language/language.go`: common operations (list, install, rehash, etc.).
  - `language/util.go`: HTTP helpers, tar extraction, OS/arch logic.
  - `language/<lang>.go`: language-specific list/install logic.
- `config/config.go`: optional `config.json` parsing for rehash filtering.
- `test.sh`: convenience wrapper for local execution.

## CLI Behavior Overview

- Global commands (e.g., `versions`, `latest`, `rehash`) operate across all languages.
- Language commands (e.g., `tinyenv python install`) operate on a single language.
- The `rehash` command generates shims in `<root>/bin` that `exec` the active version.

## Config: `config.json`

`config.json` can live at the root directory. It currently supports `rehash` filtering:

```json
{
  "rehash": {
    "python": {
      "includes": ["^python"],
      "excludes": ["-config$"]
    }
  }
}
```

This lets you constrain which executables are turned into shims.

## Adding a New Language

1. Create `language/<name>.go` implementing the `Specific` interface:
   - `List(ctx, all)`
   - `Latest(ctx)`
   - `Install(ctx, version)`
   - `BinDirs()`
   - `Untar(tarball, targetDir)` (often use the base implementation)
2. Register the language name in `language/language.go` under `var All`.
3. Ensure list/install logic uses the helpers in `language/util.go` for HTTP download and untar.

## Testing

- There are a few lightweight tests under `language/`.
- Note: some tests hit live network endpoints and may fail without network access.

```bash
$ go test ./...
```

## Notes and Pitfalls

- Many language installs depend on platform-specific assets (OS/arch).
- Some installers use HTTP HEAD/GET checks and can fail if an upstream URL changes.
- `rehash` relies on filesystem permissions in `<root>/bin`.
- `tinyenv --version` prints just the version string.

## Common Tasks

- List installed versions:
  - `tinyenv versions`
- Install a language version:
  - `tinyenv python install 3.12.5+20240814`
- Set a global version:
  - `tinyenv python global 3.12.5+20240814`
- Regenerate shims:
  - `tinyenv rehash`

## Development Checklist

- Update or add a language implementation in `language/`.
- Run `gofmt` on modified Go files.
- Run `go test ./...` if your change affects language logic.
