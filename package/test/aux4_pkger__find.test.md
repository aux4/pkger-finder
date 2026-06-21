# aux4 aux4 pkger find

These tests run against a controlled fixture of "installed" packages by
pointing `AUX4_PKGER_FINDER_HOME` at a fixture directory, so results are
deterministic regardless of what is actually installed.

```file:fx/packages/demo/mailer/README.md
# demo/mailer

Send transactional email through your provider.
```

```file:fx/packages/demo/mailer/man/mailer__send.md
#### Description

The send command delivers an email message to a recipient with a subject and body.
```

```file:fx/packages/demo/widget/README.md
# demo/widget

Render colorful charts from CSV data.
```

```file:fx/packages/demo/widget/man/widget__render.md
#### Description

The render command draws a bar chart from a CSV file and writes it to a PNG image.
```

```file:fx/packages/demo/crypto/man/crypto__encrypt.md
#### Description

The encrypt command protects secrets using AES encryption before storing them.
```

```afterAll
rm -rf fx
```

## ranking by command documentation

### should rank the matching command first

```execute
NO_COLOR=1 AUX4_PKGER_FINDER_HOME=fx aux4 aux4 pkger find "deliver an email message"
```

```expect:partial
aux4 mailer send  (demo/mailer)
The send command delivers an email message to a recipient with a subject and body.
```

### should find a command from a different package

```execute
NO_COLOR=1 AUX4_PKGER_FINDER_HOME=fx aux4 aux4 pkger find "bar chart from csv"
```

```expect:partial
aux4 widget render  (demo/widget)
```

### should match secrets encryption

```execute
NO_COLOR=1 AUX4_PKGER_FINDER_HOME=fx aux4 aux4 pkger find "encrypt secrets"
```

```expect:partial
aux4 crypto encrypt  (demo/crypto)
```

## limit

### should respect the result limit

```execute
NO_COLOR=1 AUX4_PKGER_FINDER_HOME=fx aux4 aux4 pkger find "recipient subject body" --limit 1
```

```expect:partial
aux4 mailer send  (demo/mailer)
```

## no results

### should report when nothing matches

```execute
NO_COLOR=1 AUX4_PKGER_FINDER_HOME=fx aux4 aux4 pkger find "zzzznomatchqqq"
```

```expect:partial
No commands or packages found for "zzzznomatchqqq"
```

## missing query

### should error when no query is given

```execute
NO_COLOR=1 AUX4_PKGER_FINDER_HOME=fx aux4 aux4 pkger find ""
```

```error:partial
missing search query
```
