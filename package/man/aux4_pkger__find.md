#### Description

The `find` command searches the documentation of every **installed** aux4 package and returns the commands and packages most relevant to your query, ranked by relevance.

It differs from `aux4 aux4 pkger search`, which queries the remote hub for packages you could install. `find` searches what is already on your machine and resolves a query down to the specific command to run.

- **Command-level results** — each command man page is indexed as its own document, so a result points at an exact command such as `aux4 config get`.
- **Package-level results** — each package `README.md` is indexed too, surfacing the package as a whole with its install command.
- **Relevance ranking** — results are ranked by how well each document's text matches your query; the command invocation, scope and name are folded in so naming a command or package ranks it highly.
- **Self-refreshing index** — the index lives at `~/.aux4.config/.pkg-index/index.json` and is built on first use. Before each search it is rebuilt automatically if your installed packages have changed since it was last built (detected via the `packages/all.json` ledger), so results always reflect what is currently installed.

Each result prints the exact command to run on the first line (in yellow), the package `(scope/name)` beside it (in cyan), and a short description from its documentation on the line directly below. Results are separated by a blank line.

#### Usage

```bash
aux4 aux4 pkger find <query> [--limit <n>]
```

--query   Words to search for (positional). Quote multi-word queries so they are treated as a single search.
--limit   Maximum number of results to show (default: 10)

#### Example

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

Limit the number of results:

```bash
aux4 aux4 pkger find "encrypt secrets" --limit 5
```
