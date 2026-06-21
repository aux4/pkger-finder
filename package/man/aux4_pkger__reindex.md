#### Description

The `reindex` command rebuilds the local search index used by `aux4 aux4 pkger find`. It scans every installed package under `~/.aux4.config/packages`, indexes each package `README.md` and every command man page, and writes a fresh index to `~/.aux4.config/.pkg-index/index.json`.

You rarely need to run this by hand. `find` builds the index lazily the first time it runs and rebuilds it automatically whenever your set of installed packages changes (detected via the `packages/all.json` ledger); `aux4/pkger-finder` also registers hooks on `aux4:pkger/install` and `aux4:pkger/uninstall` to pre-warm the index where the toolchain supports package hooks. Use `reindex` to force a refresh — for example after manually changing package documentation during development.

The command reports how many documents were indexed.

#### Usage

```bash
aux4 aux4 pkger reindex
```

#### Example

```bash
aux4 aux4 pkger reindex
```

```text
Indexed 606 documents from installed packages
```
