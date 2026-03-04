# section

`section` is a minimal parser for section-based configuration files with an INI-like shape.

Its purpose is intentionally narrow: it detects sections and groups their raw lines, but it does not parse keys, values, types, quoting, interpolation, or inline comments.

That design is deliberate. The package focuses on structure, not semantics.

## Why this package exists

Many configuration files share a common pattern:

- optional lines before any named section
- named sections such as `[server]` or `[database]`
- free-form content inside each section

In some projects, the exact syntax inside a section is domain-specific:

- `key = value`
- custom directives
- shell-like tokens
- embedded mini DSLs
- repeated values

Instead of forcing a single opinionated parser for all those cases, this package only splits the file into sections and preserves each content line exactly as it was written.

This lets the caller decide how to parse each section.

## What it does

The parser supports:

- blank lines are ignored
- full-line comments starting with `#` or `;` are ignored
- section headers in the form `[section]`
- all other non-empty lines are stored as raw content for the current section

Lines that appear before the first explicit section belong to an implicit root section.

## What it does not do

This package does not:

- parse `key=value`
- split tokens
- validate section content
- support nested sections
- preserve blank lines
- interpret inline comments

If you need those behaviors, you can build them on top of `Content(unit)`.

## API overview

```go
doc, err := section.Parse(r)
if err != nil {
	// handle error
}

units := doc.Units()
root := doc.Content("")
server := doc.Content("server")
hasFeature := doc.Has("feature")
```

### `Parse`

`Parse(io.Reader)` scans the input and returns a `Section`.

Parsing fails when:

- a section header is empty (`[]`)
- a section name is duplicated
- the underlying reader returns a scan error

### `Units`

`Units()` returns only explicit section names, in the same order they appeared in the file.

The implicit root section is not included.

### `Content`

`Content(unit)` returns the raw lines for a section.

Passing `""` selects the implicit root section.

The returned lines are the original raw lines from the input, so spacing is preserved.

### `Has`

`Has(unit)` reports whether the internal parsed data contains that section key.

Passing `""` checks the implicit root section.

## Example

Input:

```ini
title = demo

[server]
host = 127.0.0.1
port = 8080

[feature]
enabled = true
```

Usage:

```go
doc, err := section.Parse(strings.NewReader(src))
if err != nil {
	return err
}

fmt.Println(doc.Units())           // [server feature]
fmt.Println(doc.Content(""))       // [title = demo]
fmt.Println(doc.Content("server")) // [host = 127.0.0.1 port = 8080]
```

See `example_test.go` for a runnable example.

## When this approach is useful

This package is a good fit when:

- you want a lightweight section splitter
- section bodies are parsed differently depending on the section
- you want to preserve the original text lines
- you do not want a rigid INI schema in the core parser

It is less suitable when:

- you need a full INI parser with typed values
- you need automatic key/value lookup
- you need advanced validation or schema enforcement

## Design tradeoff

The main tradeoff is simple:

- the package is flexible because it stays generic
- the caller must do the semantic parsing work

That is the intended model.
