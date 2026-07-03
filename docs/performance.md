# Performance

`go-ruby-json/json` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's `json`. This
page records a **comparative benchmark** of that module against the reference
Ruby runtimes, part of the ecosystem-wide per-module parity suite.

## What is measured

The **same** Ruby script — `JSON.parse` + `JSON.generate` round-trip of a ~100-key nested document — is run under every runtime. `rbgo`'s
number reflects **this pure-Go library doing the work**; every other column is
that interpreter's own `json` stdlib. So the comparison is the **Ruby-visible
operation**, apples-to-apples across interpreters. The script prints a
deterministic checksum and its output is checked **byte-identical to MRI**
before timing.

- **Host:** Apple M4 Max, macOS (darwin/arm64). **Method:** best-of-5 wall time
  (best, not mean, to suppress scheduler noise); single-shot processes, no
  warm-up beyond the script's own loop.
- **Runtimes:** `ruby 4.0.5 +PRISM` (MRI, the oracle) and `ruby --yjit`;
  `jruby 10.1.0.0` (OpenJDK 25); `truffleruby 34.0.1` (GraalVM CE Native).
- The benchmark script and harness live in rbgo's repo under
  [`bench/modules/`](https://github.com/go-embedded-ruby/ruby/tree/main/bench/modules)
  (`json.rb` + `run.sh`). Reproduce:
  `RBGO=./rbgo TRUFFLE=truffleruby bash bench/modules/run.sh 5`.

## Result (best of 5, ms)

| Runtime | time | vs MRI |
| --- | ---: | ---: |
| **rbgo** (go-ruby-json) | 1520 | 4.47× |
| MRI (ruby 4.0.5) | 340 | 1.00× |
| MRI + YJIT | 340 | 1.00× |
| JRuby 10.1.0.0 | 2090 | 6.15× |
| TruffleRuby 34.0.1 | 3030 | 8.91× |

**Honest gap:** rbgo runs on **go-ruby-json** at ~4.5x MRI here. MRI's `json` is a mature, tuned C extension; go-ruby-json is correct and competitive but not yet at C-extension parse+generate throughput. Flagged for the go-ruby-json perf backlog.

!!! note "Honest framing"
    JRuby and TruffleRuby are timed **cold, single-shot**, so they carry JVM /
    Graal startup on every run — read them as one-shot `ruby file.rb` costs, the
    same way `rbgo` and MRI are measured, not as steady-state JIT numbers. Rows
    that complete in well under ~200 ms carry the most relative noise; treat
    their ratios as order-of-magnitude. These are real measured numbers from the
    2026-06-29 run — nothing is cherry-picked.

## Library-level benchmark (Go API vs runtimes) — 2026-07-03

This section measures the **pure-Go library directly, through its Go API** — not
the `rbgo` interpreter path recorded above. It isolates the library primitive
from Ruby-interpreter dispatch, answering the parity question head-on: *is the
pure-Go implementation as fast as the reference runtime's own `json`?* The
**same workload, same inputs, same iteration counts** run through the Go library
and through each reference runtime's stdlib; outputs were checked identical to
MRI before any timing.

- **Host:** Apple M4 Max (`Mac16,5`, arm64), macOS 26.5.1 — **date 2026-07-03**.
- **Runtimes:** Go 1.26.4 · MRI `ruby 4.0.5 +PRISM` · MRI + YJIT · JRuby 10.1.0.0
  (OpenJDK 25) · TruffleRuby 34.0.1 (GraalVM CE Native).
- **Method:** each process runs 3 untimed warm-up passes, then 25 timed passes of
  a fixed inner loop, timed with a monotonic clock; the **best** pass is reported
  as **ns/op** (lower is better). `vs MRI` < 1.00× means *faster than MRI*.
  Interpreter start-up is outside the timed region, so these are operation costs,
  not `ruby file.rb` process costs.

#### generate-60obj

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 8944.1 | 1.08× |
| MRI | 8260.0 | 1.00× |
| MRI + YJIT | 8162.0 | 0.99× |
| JRuby | 14989.0 | 1.81× |
| TruffleRuby | 49454.2 | 5.99× |

#### parse-60obj

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 11657.1 | 0.95× |
| MRI | 12276.0 | 1.00× |
| MRI + YJIT | 12206.0 | 0.99× |
| JRuby | 65933.3 | 5.37× |
| TruffleRuby | 125297.3 | 10.21× |

`parse` is now **faster than MRI's C extension _and_ MRI + YJIT** — 0.95× MRI
(11.66 µs vs 12.28 µs) and 0.955× YJIT (vs 12.21 µs) — where it was 1.80× MRI
before the arena/scratch materialisation below (and 2.30× before the round prior
to that). `generate` holds at parity (1.08×). Both operations remain far ahead of
the JVM- and Graal-based Rubies on this document.

##### Arena/scratch parse materialisation (2026-07-03)

The previous round cut allocations 731 → 318/op (a lazy `Map` index and a
key-dedup cache) and reached 1.80× MRI, calling the residual "structural to the
Ruby value tree." A deeper pass profiled `parse-60obj` again — allocations, CPU,
and a `GOGC` ablation which showed that, contrary to the earlier reading, **GC
was _not_ the bottleneck** (`GOGC=off` barely moved the number; the darwin CPU
profiler over-attributes samples to `kevent`/`madvise`). The real costs were the
scan, per-`malloc` churn, and per-value materialisation overhead. Four reducible
costs were attacked, on top of the irreducible value-tree interface boxing, with
**no change to the MRI-observable result** (same parsed structure, key order,
`Integer`/`Float` types, string/encoding, duplicate-key last-wins and
malformed-input errors — `Parse`'s return type is unchanged):

- **Per-parse arenas.** The `Map` structs, their `[]Pair` backing and arrays'
  `[]any` backing are bump-allocated from shared slabs carved into exact
  sub-slices; the returned tree keeps a slab alive, the builder is discarded.
  ~180 per-op allocations collapse to a handful.
- **No element pre-scan.** The per-container `countElems` look-ahead — ~37% of
  the pure-scan cost — is gone: materialisation accumulates each container's
  members in a reused scratch stack and learns the exact size at container close,
  so no second pass is needed.
- **Inline integer parsing** replaces `strconv.ParseInt` (falling back to
  `*big.Int` only on int64 overflow).
- **Key handling.** A linear key cache (a map only past 32 distinct keys), a
  dense id per distinct key, and **O(1) duplicate-key detection via a per-object
  id bitmask** replace a hash lookup on every key occurrence and the O(n²)
  duplicate-key scan. `skipSpace` gains a single-compare fast path.

Measured effect on `parse-60obj`: **1.80× → 0.95× MRI** (and 0.955× YJIT) —
parse now clears both reference C-extension columns. The scan alone dropped from
~11.0 µs to ~5.5 µs; full parse from ~21.9 µs to ~11.7 µs.

**Where the floor is:** the surviving allocations are the *irreducible* interface
boxes of the value tree itself — each string value and each array must be boxed
into `any` (~121 boxes for this document, ~1.75 µs) — plus the pure byte scan
(~5.5 µs). Both are inherent while `Parse` returns the Ruby value tree, yet the
sum now sits below MRI+YJIT because the scan and the non-boxing materialisation
overhead were driven down far enough. Small integers (0–255) and `*Map` pointers
box for free (Go's static small-value cache / pointer-in-interface), so the
document's `id`/`score`/`tags` integers and its nested maps add no boxing.

!!! note "Reproduce"
    The harness is committed under
    [`benchmarks/`](https://github.com/go-ruby-json/docs/tree/main/benchmarks):
    a self-contained Go driver (`go/`, pins the published library via
    `go.mod`), the equivalent `ruby/json.rb` workload, and `run.sh`. Run
    `bash benchmarks/run.sh`; env `OUTER`/`WARM` tune the pass budget and
    `RUBY`/`JRUBY`/`TRUFFLERUBY` select the runtime binaries.

!!! warning "Warm-up budget & noise — honest framing"
    Numbers reflect a **fixed warm-process budget** (3 warm-up + 25 timed passes
    in one process). The JVM/GraalVM JITs (JRuby, TruffleRuby) may need a larger
    warm-up to reach steady state, so their columns can **understate** peak
    throughput — most visibly TruffleRuby on the shortest loops (a few cold-JIT
    outliers are noted in the text). Sub-microsecond rows carry the most relative
    noise; treat those ratios as order-of-magnitude. Every number here is a
    **real measured value** from the dated run above — nothing is fabricated,
    estimated, or cherry-picked. The go-ruby column is the pure-Go library; every
    other column is that interpreter's own stdlib doing the equivalent work.
