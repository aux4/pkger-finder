# aux4 aux4 pkger reindex

These tests point `AUX4_PKGER_FINDER_HOME` at a fixture directory so the indexed
document count is deterministic.

```file:fxr/packages/demo/widget/README.md
# demo/widget

Render colorful charts from CSV data.
```

```file:fxr/packages/demo/widget/man/widget__render.md
#### Description

The render command draws a bar chart from a CSV file.
```

```afterAll
rm -rf fxr
```

## rebuilding the index

### should report the number of indexed documents

```execute
AUX4_PKGER_FINDER_HOME=fxr aux4 aux4 pkger reindex
```

```expect
Indexed 2 documents from installed packages
```

### should make the indexed commands searchable

```execute
NO_COLOR=1 AUX4_PKGER_FINDER_HOME=fxr aux4 aux4 pkger find "bar chart"
```

```expect:partial
aux4 widget render  (demo/widget)
```
