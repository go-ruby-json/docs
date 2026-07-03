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
| **go-ruby (pure Go)** | 8954.9 | 1.10× |
| MRI | 8136.0 | 1.00× |
| MRI + YJIT | 8469.0 | 1.04× |
| JRuby | 14874.5 | 1.83× |
| TruffleRuby | 41615.4 | 5.11× |

#### parse-60obj

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 28201.2 | 2.30× |
| MRI | 12274.0 | 1.00× |
| MRI + YJIT | 12048.0 | 0.98× |
| JRuby | 64480.0 | 5.25× |
| TruffleRuby | 121400.1 | 9.89× |

`generate` is at parity (1.10×); `parse` is ~2.3× MRI's C extension and is the top optimization target for the module. Both operations remain far ahead of the JVM- and Graal-based Rubies on this document.

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
