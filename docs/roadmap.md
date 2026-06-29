# Roadmap

`go-ruby-json/json` is grown **test-first**, each capability differential-tested against MRI
rather than built in isolation. Ruby's JSON — the
deterministic, interpreter-independent slice extracted from rbgo's internals — is
**complete**.

| Stage | What | Status |
| --- | --- | --- |
| Parser | `JSON.parse(s)` parses JSON text into the Ruby value graph (Hash/Array/String/Integer/Float/true/false/nil), with `symbolize_names` to key hashes by Symbol, exactly as MRI's parser does. | **Done** |
| Generator | `JSON.generate(v)` emits compact JSON and `JSON.pretty_generate(v)` the indented form, matching reference Ruby's spacing and escaping byte-for-byte. | **Done** |
| Nesting & limits | Depth tracking that raises `JSON::NestingError` past the configured limit, mirroring MRI's recursion guard on both parse and generate. | **Done** |
| Error taxonomy | The `JSON::ParserError` / `JSON::NestingError` / `JSON::GeneratorError` hierarchy, raised on the same inputs and with the same messages as reference Ruby. | **Done** |
| Number & string fidelity | Integers, floats, and escapes (including `\uXXXX` surrogate pairs) parsed and emitted exactly as MRI, including the float formatting edge cases. | **Done** |
| Differential oracle & coverage | A wide corpus round-tripped here and by the system `ruby`/`json`, compared byte-for-byte; 100% coverage, gofmt + go vet clean, green across all six 64-bit Go arches and three OSes. | **Done** |

## Documented out-of-scope boundaries

These are **deliberate**, recorded so the module's surface is unambiguous:

- **No interpreter.** The library implements the deterministic algorithm; it
  never runs arbitrary Ruby. Anything that needs a live binding or evaluation is
  the consumer's job — that is why `rbgo` binds this module rather than the
  reverse.
- **Reference is reference Ruby (MRI).** Byte-for-byte conformance targets MRI's
  behaviour; differences across MRI releases are matched to the reference used by
  the differential oracle.
- **Standalone & reusable.** The module has no dependency on the Ruby runtime;
  the dependency runs the other way.

See [Usage & API](api.md) for the surface and [Why pure Go](why.md) for the
deterministic/interpreter split.
