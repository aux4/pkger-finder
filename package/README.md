# aux4/pkger-finder

`aux4/pkger-finder` makes the commands you already have installed discoverable. It builds a local search index over the documentation of every installed aux4 package — each package `README.md` and every command man page — and lets you search it with `aux4 aux4 pkger find`.

Unlike `aux4 aux4 pkger search`, which queries the remote hub for packages you could install, `pkger-finder` searches what is **already installed** on your machine and resolves a query down to the specific command to run.

## Installation

```bash
aux4 aux4 pkger install aux4/pkger-finder
```

The index is built automatically the first time you run `find`, and is kept up to date through hooks that re-index whenever a package is installed or uninstalled.

## Quick Start

Search by what you want to do — results are ranked by relevance:

```bash
aux4 aux4 pkger find "get a configuration value"
```

```text
aux4 config get  (aux4/config)
If file is not provided it will look for a file named config.yaml, config.yml, or config.json in the current directory.

aux4 config set  (aux4/config)
If file is not provided it will look for a file named config.yaml, config.yml, or config.json in the current directory.

aux4 config merge  (aux4/config)
If file is not provided it will look for a file named config.yaml, config.yml, or config.json in the current directory.
```

The first line of each result is the exact command to run (shown in yellow), with its package in cyan; the line below is a short description pulled from its documentation.

## How It Works

- **Documents.** Each command man page (`man/<profile>__<command>.md`) becomes one searchable document, keyed by the command it documents. Each package `README.md` becomes one document representing the package as a whole.
- **Ranking.** Results are ranked by how well each document's text matches your query. The command invocation, scope and name are folded into each document so naming a command or package ranks it highly.
- **Index location.** The index is a single JSON file at `~/.aux4.config/.pkg-index/index.json`. It is self-contained — searching never re-reads package files.
- **Freshness.** `find` keeps itself up to date automatically: it rebuilds the index whenever your set of installed packages has changed since the index was last built. It does this by comparing the index against the package ledger (`packages/all.json`), which is rewritten on every install and uninstall — so results always reflect what is currently installed, with no manual step.
- **Pre-warming.** `aux4/pkger-finder` also registers hooks on `aux4:pkger/install` and `aux4:pkger/uninstall` to rebuild the index proactively. Where the installed toolchain supports package hooks this avoids the rebuild cost on the next `find`; where it does not, the freshness check above still guarantees correct results.
- **Manual rebuild.** You can always force a rebuild with `aux4 aux4 pkger reindex`.

**Note:** The first `find` after installing `aux4/pkger-finder` builds the index lazily, so it works immediately with no setup.

## Commands

### `aux4 aux4 pkger find <query>`

Search installed packages and commands by their documentation.

```bash
aux4 aux4 pkger find "send an email"
aux4 aux4 pkger find "encrypt secrets" --limit 5
```

| Option | Description | Default |
|--------|-------------|---------|
| `query` | Words to search for (positional). Quote multi-word queries. | required |
| `--limit` | Maximum number of results to show | `10` |

Multi-word queries must be quoted so they are treated as a single search, e.g. `find "database migration"`.

### `aux4 aux4 pkger reindex`

Rebuild the search index from scratch over all installed packages. This is a **private** command (hidden from `--help`) since it runs automatically on install/uninstall — but it's still callable if you ever need to force a refresh.

```bash
aux4 aux4 pkger reindex
```

```text
Indexed 606 documents from installed packages
```

## Environment Variables

| Variable | Effect |
|----------|--------|
| `AUX4_PKGER_FINDER_HOME` | Override the base directory under which `packages/` is scanned and the index is written (default `~/.aux4.config`). |
| `NO_COLOR` | Disable ANSI styling in results. Color is also disabled automatically when output is not a terminal. |
