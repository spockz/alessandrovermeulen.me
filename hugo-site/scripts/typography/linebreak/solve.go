// This file ports the upstream search while retaining exact incumbent rescoring and dash-history states.
// SPDX-License-Identifier: Apache-2.0
// Derived from the Typst contributors' paragraph line-breaking implementation via Encarta.
// Upstream: typst/typst, crates/typst-layout/src/inline/linebreak.rs
// Revision: 9dfd3a08500b7896045f907433cf7b4b02434fad (v0.15.1)
// Modified: translated from Swift to Go; added validation and explicit measurement errors.
package linebreak

import (
	"fmt"
	"math"
	"slices"
)

const boundEpsilon = 1e-3

type pathEntry struct {
	predecessor  int
	end          int
	breakpoint   Breakpoint
	metrics      Metrics
	total        float64
	ratio        float64
	endsWithDash bool
}

type approximateLine struct {
	rangeUTF16  Range
	breakpoint  Breakpoint
	unbreakable bool
}

func Solve(problem Problem) (Output, error) {
	if err := validateProblem(problem); err != nil {
		return Output{}, err
	}
	seed, err := approximatePath(problem)
	if err != nil {
		return Output{}, err
	}
	measurements := 0
	incumbent, feasible, err := evaluateSeed(seed, problem, &measurements)
	if err != nil {
		return Output{}, err
	}
	bound := math.Inf(1)
	if problem.EnableIncumbentPruning {
		bound = incumbent
	}
	path, err := exactPath(problem, bound, &measurements)
	if err != nil {
		return Output{}, err
	}
	if len(path) > 0 {
		return makeOutput(path, problem, measurements, !feasible || !problem.EnableIncumbentPruning), nil
	}
	retryProblem := problem
	retryProblem.EnableOverfullRetirement = false
	path, err = exactPath(retryProblem, math.Inf(1), &measurements)
	if err != nil {
		return Output{}, err
	}
	if len(path) == 0 {
		return Output{}, ErrNoCompleteLayout
	}
	return makeOutput(path, problem, measurements, true), nil
}

func evaluateSeed(seed []approximateLine, problem Problem, measurements *int) (float64, bool, error) {
	incumbent := 0.0
	feasible := len(seed) > 0
	previousDash := false
	for _, line := range seed {
		metrics, err := measureExact(problem, line.rangeUTF16, line.breakpoint, previousDash, measurements)
		if err != nil {
			return 0, false, err
		}
		ratio := ratioFor(metrics, problem)
		if ratio < problem.MinRatio {
			feasible = false
		}
		params := costParams{
			breakpoint:      line.breakpoint,
			ratio:           ratio,
			unbreakable:     line.unbreakable,
			consecutiveDash: previousDash && metrics.EndsWithDash,
		}
		incumbent += lineCost(metrics, params, problem, false)
		if !finite(incumbent) {
			return 0, false, fmt.Errorf("%w: seed cost overflow", ErrInvalidProblem)
		}
		previousDash = metrics.EndsWithDash
	}
	if !feasible {
		incumbent = math.Inf(1)
	}
	return incumbent, feasible, nil
}

func initialTable() []pathEntry {
	return []pathEntry{{breakpoint: Breakpoint{Kind: Mandatory}}}
}

func approximatePath(problem Problem) ([]approximateLine, error) {
	table := initialTable()
	active := 0
	previousEnd := 0
	for _, breakpoint := range problem.Breakpoints {
		var bestWithoutDash, bestWithDash *pathEntry
		for index := active; index < len(table); index++ {
			predecessor := table[index]
			rangeUTF16 := Range{StartUTF16: predecessor.end, EndUTF16: breakpoint.EndUTF16}
			metrics := problem.Approximate(rangeUTF16, breakpoint)
			if err := validateMetrics(metrics); err != nil {
				return nil, fmt.Errorf("approximate candidate [%d,%d): %w", rangeUTF16.StartUTF16, rangeUTF16.EndUTF16, err)
			}
			ratio := ratioFor(metrics, problem)
			params := costParams{
				breakpoint:      breakpoint,
				ratio:           ratio,
				unbreakable:     previousEnd == predecessor.end,
				consecutiveDash: predecessor.breakpoint.Kind == Hyphen && breakpoint.Kind == Hyphen,
			}
			total := predecessor.total + lineCost(metrics, params, problem, true)
			if !finite(total) {
				return nil, fmt.Errorf("%w: approximate cost overflow", ErrInvalidProblem)
			}
			entry := pathEntry{
				predecessor:  index,
				end:          breakpoint.EndUTF16,
				breakpoint:   breakpoint,
				total:        total,
				ratio:        ratio,
				endsWithDash: breakpoint.Kind == Hyphen,
			}
			updateBest(entry, &bestWithoutDash, &bestWithDash)
			if ratio < problem.MinRatio && index == active {
				active++
			}
		}
		if breakpoint.Kind.isMandatory() {
			active = len(table)
		}
		table = appendBest(table, bestWithoutDash, bestWithDash)
		previousEnd = breakpoint.EndUTF16
	}
	path := buildPath(table, problem.Length)
	seed := make([]approximateLine, len(path))
	for index, entry := range path {
		start := table[entry.predecessor].end
		unbreakable := true
		for _, breakpoint := range problem.Breakpoints {
			if breakpoint.EndUTF16 > start && breakpoint.EndUTF16 < entry.end {
				unbreakable = false
				break
			}
		}
		seed[index] = approximateLine{
			rangeUTF16:  Range{StartUTF16: start, EndUTF16: entry.end},
			breakpoint:  entry.breakpoint,
			unbreakable: unbreakable,
		}
	}
	return seed, nil
}

func exactPath(problem Problem, upperBound float64, measurements *int) ([]pathEntry, error) {
	table := initialTable()
	active := 0
	previousEnd := 0
	for _, breakpoint := range problem.Breakpoints {
		var bestWithoutDash, bestWithDash *pathEntry
		for index := active; index < len(table); index++ {
			predecessor := table[index]
			rangeUTF16 := Range{StartUTF16: predecessor.end, EndUTF16: breakpoint.EndUTF16}
			metrics, err := measureExact(problem, rangeUTF16, breakpoint, predecessor.endsWithDash, measurements)
			if err != nil {
				return nil, err
			}
			ratio := ratioFor(metrics, problem)
			params := costParams{
				breakpoint:      breakpoint,
				ratio:           ratio,
				unbreakable:     previousEnd == predecessor.end,
				consecutiveDash: predecessor.endsWithDash && metrics.EndsWithDash,
			}
			total := predecessor.total + lineCost(metrics, params, problem, false)
			if !finite(total) {
				return nil, fmt.Errorf("%w: exact cost overflow", ErrInvalidProblem)
			}
			if problem.EnableOverfullRetirement && ratio < problem.MinRatio && index == active {
				active++
			}
			if total > upperBound+boundEpsilon {
				continue
			}
			entry := pathEntry{
				predecessor:  index,
				end:          breakpoint.EndUTF16,
				breakpoint:   breakpoint,
				metrics:      metrics,
				total:        total,
				ratio:        ratio,
				endsWithDash: metrics.EndsWithDash,
			}
			updateBest(entry, &bestWithoutDash, &bestWithDash)
		}
		if breakpoint.Kind.isMandatory() {
			active = len(table)
		}
		table = appendBest(table, bestWithoutDash, bestWithDash)
		previousEnd = breakpoint.EndUTF16
	}
	return buildPath(table, problem.Length), nil
}

func updateBest(entry pathEntry, bestWithoutDash, bestWithDash **pathEntry) {
	best := bestWithoutDash
	if entry.endsWithDash {
		best = bestWithDash
	}
	if *best == nil || (*best).total >= entry.total {
		*best = &entry
	}
}

func appendBest(table []pathEntry, bestWithoutDash, bestWithDash *pathEntry) []pathEntry {
	if bestWithoutDash != nil {
		table = append(table, *bestWithoutDash)
	}
	if bestWithDash != nil {
		table = append(table, *bestWithDash)
	}
	return table
}

func buildPath(table []pathEntry, length int) []pathEntry {
	terminal := -1
	for index, entry := range table {
		if entry.end == length && entry.breakpoint.Kind == ParagraphEnd && (terminal < 0 || table[terminal].total >= entry.total) {
			terminal = index
		}
	}
	if terminal < 0 {
		return nil
	}
	var path []pathEntry
	for index := terminal; index != 0; index = table[index].predecessor {
		path = append(path, table[index])
	}
	slices.Reverse(path)
	return path
}

func makeOutput(path []pathEntry, problem Problem, measurements int, retried bool) Output {
	lines := make([]Line, len(path))
	start := 0
	for index, entry := range path {
		fit := entry.metrics.Width <= problem.Width+geometryZeroTolerance
		if problem.Justified {
			delta := problem.Width - entry.metrics.Width
			needsStretch := delta > geometryZeroTolerance && entry.breakpoint.Kind != ParagraphEnd
			stretchPerSlot := math.Inf(1)
			if entry.metrics.Spaces > 0 {
				stretchPerSlot = delta / float64(entry.metrics.Spaces)
			}
			canStretch := !needsStretch || (entry.metrics.Spaces > 0 && stretchPerSlot <= problem.FontSize)
			fit = entry.ratio >= problem.MinRatio && canStretch
		}
		lines[index] = Line{
			StartUTF16: start,
			EndUTF16:   entry.end,
			Breakpoint: entry.breakpoint,
			Metrics:    entry.metrics,
			Fit:        fit,
		}
		start = entry.end
	}
	return Output{
		Lines:                 lines,
		TotalCost:             path[len(path)-1].total,
		ExactMeasurementCount: measurements,
		UsedUnboundedRetry:    retried,
	}
}

func measureExact(problem Problem, rangeUTF16 Range, breakpoint Breakpoint, previousDash bool, count *int) (Metrics, error) {
	metrics, err := problem.Exact(rangeUTF16, breakpoint, previousDash)
	if err != nil {
		return Metrics{}, fmt.Errorf("exact candidate [%d,%d): %w", rangeUTF16.StartUTF16, rangeUTF16.EndUTF16, err)
	}
	*count++
	if err := validateMetrics(metrics); err != nil {
		return Metrics{}, fmt.Errorf("exact candidate [%d,%d): %w", rangeUTF16.StartUTF16, rangeUTF16.EndUTF16, err)
	}
	return metrics, nil
}

func validateProblem(problem Problem) error {
	if problem.Length < 0 || !finite(problem.FontSize) || problem.FontSize <= 0 || !finite(problem.Width) || problem.Width <= 0 {
		return fmt.Errorf("%w: length must be nonnegative and font size and width must be finite and positive", ErrInvalidProblem)
	}
	if !finite(problem.HyphenCost) || problem.HyphenCost < 0 || !finite(problem.RuntCost) || problem.RuntCost < 0 || !finite(problem.MinRatio) {
		return fmt.Errorf("%w: costs must be finite and nonnegative and minimum ratio must be finite", ErrInvalidProblem)
	}
	if problem.Approximate == nil || problem.Exact == nil {
		return fmt.Errorf("%w: approximate and exact callbacks are required", ErrInvalidProblem)
	}
	previousEnd := 0
	for index, breakpoint := range problem.Breakpoints {
		if breakpoint.EndUTF16 < previousEnd || breakpoint.EndUTF16 > problem.Length {
			return fmt.Errorf("%w: breakpoint %d is outside the ordered paragraph range", ErrInvalidProblem, index)
		}
		if breakpoint.Kind > ParagraphEnd || (breakpoint.Kind == ParagraphEnd && breakpoint.EndUTF16 != problem.Length) {
			return fmt.Errorf("%w: breakpoint %d has an invalid kind or paragraph end", ErrInvalidProblem, index)
		}
		previousEnd = breakpoint.EndUTF16
	}
	return nil
}

func validateMetrics(metrics Metrics) error {
	if !finite(metrics.Width) || !finite(metrics.Stretch) || !finite(metrics.Shrink) || metrics.Spaces < 0 {
		return fmt.Errorf("%w: widths and capacities must be finite and space count must be nonnegative", ErrInvalidMetrics)
	}
	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
