# go-ruby-json documentation

**Ruby's JSON parser and generator in pure Go — MRI-compatible, no cgo.**

`go-ruby-json/json` is a faithful, pure-Go (zero cgo) reimplementation of Ruby's JSON,
matching reference Ruby (MRI) byte-for-byte. The module path is
`github.com/go-ruby-json/json`.

It was **extracted from rbgo's prelude/internals into a reusable standalone
library**: the module is standalone and importable by any Go program, and it is
the backend bound into [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)
by `rbgo` as a native module — just like
[go-ruby-regexp](https://github.com/go-ruby-regexp) and
[go-ruby-erb](https://github.com/go-ruby-erb). The dependency runs the other
way: this library has **no dependency on the Ruby runtime**.

!!! success "Status: parser + generator complete — MRI byte-exact"
    Faithful port of Ruby's JSON: **`JSON.parse`** parser and **`JSON.generate`** / **`JSON.pretty_generate`** generator, with **`symbolize_names`**, depth-limited **nesting**, and the full **`JSON::ParserError` / `JSON::NestingError` / `JSON::GeneratorError`** taxonomy. Validated by a **differential oracle** against the system `ruby` / `json` — parsed and generated values compared byte-for-byte — at 100% coverage, `gofmt` + `go vet` clean, CI green across the six 64-bit Go targets and three OSes.

## Quick taste

```go
v, _ := json.Parse(`{"a":1,"b":[2,3]}`)        // Ruby value graph
s, _ := json.Generate(map[string]any{"a": 1})  // `{"a":1}`
p, _ := json.PrettyGenerate(v)                 // indented form
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`json`](https://github.com/go-ruby-json/json) | the library — Ruby's JSON in pure Go |
| [`docs`](https://github.com/go-ruby-json/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-ruby-json.github.io`](https://github.com/go-ruby-json/go-ruby-json.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-ruby-json/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain.
- **MRI byte-exact.** Output matches reference Ruby exactly, not approximately,
  validated by a differential oracle against the `ruby` binary.
- **Standalone & reusable.** Extracted from rbgo's internals; no dependency on
  the Ruby runtime — the dependency runs the other way.
- **100% test coverage** is the target, enforced as a CI gate, across 6 arches
  and 3 OSes.

## Where to go next

- [Why pure Go](why.md) — why this slice of Ruby is deterministic enough to live
  as a standalone, interpreter-independent Go library.
- [Usage & API](api.md) — the public surface and worked examples.
- [Roadmap](roadmap.md) — what is done and what is downstream by design.

Source lives at [github.com/go-ruby-json/json](https://github.com/go-ruby-json/json).
