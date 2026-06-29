# Usage & API

The public API lives at the module root (`github.com/go-ruby-json/json`). It is **Ruby-shaped but Go-idiomatic**: `Parse` / `Generate` mirror Ruby's `JSON.parse` / `JSON.generate`, while the surface follows Go conventions — an explicit `error`, value types, no global state.

!!! success "Status: implemented"
    The library is built and importable as `github.com/go-ruby-json/json`, bound into
    `rbgo` as a native module; see [Roadmap](roadmap.md).

## Install

```sh
go get github.com/go-ruby-json/json
```

## Worked example

```go
v, _ := json.Parse(`{"a":1,"b":[2,3]}`)        // Ruby value graph
s, _ := json.Generate(map[string]any{"a": 1})  // `{"a":1}`
p, _ := json.PrettyGenerate(v)                 // indented form
```

## Shape

```go
// Parse parses JSON text into the Ruby value graph (JSON.parse).
func Parse(s string, opts ...Option) (Value, error)

// Generate emits compact JSON for a Ruby value (JSON.generate).
func Generate(v Value) (string, error)

// PrettyGenerate emits the indented form (JSON.pretty_generate).
func PrettyGenerate(v Value) (string, error)
```

## MRI conformance

Correctness is defined by reference Ruby. A **differential oracle** runs a wide
corpus through both the system `ruby` and this library and compares the results
**byte-for-byte** — not approximated from memory. The oracle tests skip
themselves where `ruby` is not on `PATH` (e.g. the qemu arch lanes), so the
cross-arch builds still validate the library.

## Relationship to Ruby

`go-ruby-json/json` is **standalone and reusable**, and is the backend bound into
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) by `rbgo` as a
native module — the same way [go-ruby-regexp](https://github.com/go-ruby-regexp)
and [go-ruby-erb](https://github.com/go-ruby-erb) are bound. The dependency runs
the other way: this library has no dependency on the Ruby runtime.
