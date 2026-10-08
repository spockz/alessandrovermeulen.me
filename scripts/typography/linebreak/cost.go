// This file retains the upstream ratio and demerit equations in a single numerical implementation.
// SPDX-License-Identifier: Apache-2.0
// Derived from the Typst contributors' paragraph line-breaking implementation via Encarta.
// Upstream: typst/typst, crates/typst-layout/src/inline/linebreak.rs
// Revision: 9dfd3a08500b7896045f907433cf7b4b02434fad (v0.15.1)
// Modified: translated from Rust through Swift to Go; removed engine-specific item types.
package linebreak

import "math"

const geometryZeroTolerance = 1e-9

type costParams struct {
	breakpoint      Breakpoint
	ratio           float64
	unbreakable     bool
	consecutiveDash bool
}

func ratioFor(metrics Metrics, problem Problem) float64 {
	delta := problem.Width - metrics.Width
	if math.Abs(delta) <= geometryZeroTolerance {
		delta = 0
	}
	capacity := metrics.Shrink
	if delta >= 0 {
		capacity = metrics.Stretch
	}
	capacity = math.Max(0, capacity)
	ratio := delta / capacity
	if math.IsNaN(ratio) {
		ratio = 0
	}
	if ratio > 1 {
		extra := (delta - capacity) / float64(max(metrics.Spaces, 1))
		ratio = 1 + extra/(problem.FontSize/2)
	}
	return math.Min(10, math.Max(-2, ratio))
}

func lineCost(metrics Metrics, params costParams, problem Problem, approximate bool) float64 {
	shouldJustify := metrics.ShouldJustify || (!params.breakpoint.Kind.isMandatory() && problem.Justified)
	overfullRatio := 0.0
	if problem.Justified {
		overfullRatio = -1
		if approximate {
			overfullRatio = -0.5
		}
	}
	badness := 0.0
	if params.ratio < overfullRatio {
		badness = 1_000_000
	} else if !params.breakpoint.Kind.isMandatory() || shouldJustify || params.ratio < 0 {
		badness = 100 * math.Pow(math.Abs(params.ratio), 3)
	}
	penalty := 0.0
	if params.unbreakable && params.breakpoint.Kind.isMandatory() {
		penalty += problem.RuntCost
	}
	if params.breakpoint.Kind == Hyphen {
		steps := 5 - min(5, int(params.breakpoint.HyphenLeft)) + 5 - min(5, int(params.breakpoint.HyphenRight))
		penalty += (1 + 0.15*float64(steps)) * problem.HyphenCost
	}
	if params.consecutiveDash {
		penalty += problem.HyphenCost
	}
	return math.Pow(1+badness+penalty, 2)
}
