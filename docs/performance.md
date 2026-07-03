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
| **go-ruby (pure Go)** | 9080.9 | 1.12× |
| MRI | 8073.0 | 1.00× |
| MRI + YJIT | 8048.0 | 1.00× |
| JRuby | 14992.5 | 1.86× |
| TruffleRuby | 44100.6 | 5.46× |

#### parse-60obj

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 21866.7 | 1.80× |
| MRI | 12157.0 | 1.00× |
| MRI + YJIT | 12129.0 | 1.00× |
| JRuby | 65483.0 | 5.39× |
| TruffleRuby | 127274.9 | 10.47× |

`generate` is at parity (1.12×); `parse` is now **1.80× MRI's C extension**, down
from **2.30×** before the parse hot-path optimization (below). Both operations
remain far ahead of the JVM- and Graal-based Rubies on this document.

##### Parse hot-path optimization (2026-07-03)

An allocation profile of the parse path found two avoidable allocation sources —
a per-object `map[any]int` key-index built for *every* object, and re-boxing each
object key on *every* occurrence — that were ~75% of the ~731 allocs/op. Both
were removed with **no change to the MRI-observable result** (same parsed
structure, key order, `Integer`/`Float` types, string/encoding handling and
malformed-input errors):

- **Lazy `Map` index** — small objects (the common case) resolve keys by linear
  scan with no map allocation; the hash index is materialised only once an object
  grows past 16 pairs, so large hashes keep O(1) lookup. Duplicate-key last-wins
  and insertion order are unchanged.
- **Key dedup cache** — a key string that recurs across sibling records is boxed
  into an interface once and reused, mirroring MRI's frozen-fstring key dedup;
  only the shared interface header changes, never the contents.

Measured effect on `parse-60obj`: **731 → 318 allocs/op**, **39944 → 20216
B/op**, **2.30× → 1.80× MRI**. The GC pressure that dominated the CPU profile
(concurrent `madvise`/`kevent`) is roughly halved.

**Honest residual:** the remaining allocations are *structural* to the MRI value
model — each object is a `*Map` over `[]Pair`, each array a boxed `[]any`, and
string/array values must be boxed into `any`; these cannot be removed without
changing the observable output. MRI's mature C extension keeps a modest edge
(~1.8×) on this document, so full parse parity is not reachable while the return
type is the Ruby value tree.

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
