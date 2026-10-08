<!-- This file records source identity, licensing, and the limits of this numerical translation. -->

# Line-breaking kernel provenance

This package translates Encarta's standalone Swift numerical kernel into Go. Its scoring and search derive from Typst v0.15.1, revision `9dfd3a08500b7896045f907433cf7b4b02434fad`, `crates/typst-layout/src/inline/linebreak.rs` (SHA-256 `8ff2d0a71ce8ed96005571bef7689a9674b1863358929b918b808b72b3dc6107`). The Typst contributors' Apache-2.0 license is copied unchanged into `LICENSE`.

The immediate sources are from Encarta's `ParagraphLineBreaker` directory:

| Source                                                  | SHA-256                                                            |
| ------------------------------------------------------- | ------------------------------------------------------------------ |
| `Sources/ParagraphLineBreaker/KnuthPlassCore.swift`     | `15d1647cf265553ec0305dd588145e0c19097d4186d49bad9270065c46af01ad` |
| `Sources/ParagraphLineBreaker/TypstLineBreakCost.swift` | `ac9325593962e87d2bee7ac913550a56c189372b1ca957274aac21f4b19a6530` |
| `Sources/ParagraphLineBreaker/KernelTypes.swift`        | `b3d228b242bbe0b8b96ddb6e4a6e93db649863d838b7f12a2dba82e8c8e39bb8` |
| `Sources/ParagraphLineBreaker/Types.swift`              | `8a2321c77bdbdc2f776da939612b22ab65da5f1d357122f7ae67d66892da50eb` |

The translation retains approximate-path seeding followed by exact rescoring, incumbent pruning, optional overfull-candidate retirement, mandatory-break resets, later-predecessor ties, and separate best paths for the two dash-history states. It omits shaping, language segmentation, glyph adjustment, and rendering. It adds Go input validation and wraps callback errors without discarding their identity. Cost overflow and missing complete paths return explicit errors. All lengths and capacities use the same caller-selected units; font size and line width must use those units too. Source positions are UTF-16 offsets.

`EnableIncumbentPruning` must be enabled explicitly in a Go struct literal. `EnableOverfullRetirement` defaults to false: retirement assumes candidate widths remain monotone as a line grows. When enabled, it is a fast path whose complete result can differ from exhaustive search for nonmonotone measurements. If the first exact search has no complete path, a second unbounded search disables retirement. `UsedUnboundedRetry` follows Encarta's diagnostics: it is true for an infeasible seed, disabled incumbent pruning, or an actual retry. Negative finite width and capacity values remain accepted; adjustment capacities are clamped to zero by the upstream ratio equation. Space counts must be nonnegative.

Encarta reviewed the upstream `NOTICE` at SHA-256 `1778244777547c281b6f5fa9fc0c18ab21f8d4491c803f64e09046800f5fcb26`. The notices cover other upstream assets and components; none applies to this extracted numerical kernel.

## Fixtures

The JSON files under `testdata` are copied unchanged from Encarta's `Tests/ParagraphLineBreakerTests/Fixtures`. `fixtures_test.go` adapts `KnuthPlassPinnedFixtureTests.swift` and `TypstTestFixtures.swift` without changing the JSON format.

| Fixture                          | SHA-256                                                            | Scope                                                                                                                |
| -------------------------------- | ------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------- |
| `TypstNumericalFixtures.json`    | `01daf0c9fd8bea82eb53a4d5a8daf74178fea523ef02d2e6ec0e33f8b7254464` | Four controlled single-line numerical vectors captured from pinned Typst `raw_cost`.                                 |
| `TypstOptimizedPathFixture.json` | `82507a0812cbcf2f798ef8b7fd33a516e41a70a874a61063ea5860eda3f062da` | 136 candidate metrics and a selected four-line path captured by instrumenting the real pinned `typst-cli` optimizer. |

The optimized fixture covers one paragraph in Libertinus Serif at 10 pt with a 100 pt measure. Measurements depend on the generator host's font files. Its source text, generator name, font, and revision are retained in the JSON. These tests establish numerical and selected-path agreement for those recorded inputs; they do not establish general font, document, language, or rendering parity with Typst.

Encarta generated the fixtures using `Tools/TypstReference/capture-pinned-raw-costs.py` and `capture_pinned_optimized.py`. The latter compiles `TypstOptimizedParagraph.typ` through the real CLI in a temporary detached worktree and instruments `linebreak_optimized_bounded`. Neither generator nor the Typst renderer is part of this Go package.
