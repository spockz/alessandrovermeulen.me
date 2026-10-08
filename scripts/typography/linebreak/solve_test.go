// This file tests observable search behavior where pruning, dash history, and forced breaks affect paths.
// SPDX-License-Identifier: Apache-2.0
// Adapted from Encarta's standalone kernel tests; see UPSTREAM.md.
package linebreak

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func syntheticProblem(ends ...int) Problem {
	breakpoints := make([]Breakpoint, len(ends))
	for index, end := range ends {
		kind := Opportunity
		if index == len(ends)-1 {
			kind = ParagraphEnd
		}
		breakpoints[index] = Breakpoint{EndUTF16: end, Kind: kind}
	}
	return Problem{
		Length:                 ends[len(ends)-1],
		FontSize:               10,
		Width:                  10,
		Justified:              true,
		HyphenCost:             DefaultHyphenCost,
		RuntCost:               DefaultRuntCost,
		MinRatio:               -1,
		Breakpoints:            breakpoints,
		Approximate:            func(Range, Breakpoint) Metrics { return Metrics{Width: 10} },
		Exact:                  func(Range, Breakpoint, bool) (Metrics, error) { return Metrics{Width: 10}, nil },
		EnableIncumbentPruning: true,
	}
}

func TestSeedIsExactlyRescoredBeforeBounding(t *testing.T) {
	problem := syntheticProblem(2, 4, 6)
	var measured []Range
	problem.Exact = func(rangeUTF16 Range, _ Breakpoint, _ bool) (Metrics, error) {
		measured = append(measured, rangeUTF16)
		width := 10.0
		if rangeUTF16 == (Range{0, 6}) {
			width = 9
		}
		return Metrics{Width: width, ShouldJustify: true}, nil
	}
	output, err := Solve(problem)
	if err != nil {
		t.Fatal(err)
	}
	if measured[0] != (Range{0, 6}) || !reflect.DeepEqual(lineEnds(output), []int{2, 6}) || output.UsedUnboundedRetry {
		t.Fatalf("seed rescoring did not supply a usable bound: ends=%v, first=%v, retry=%t", lineEnds(output), measured[0], output.UsedUnboundedRetry)
	}
	assertCost(t, output.TotalCost, 2)
	if output.ExactMeasurementCount != len(measured) {
		t.Fatalf("measurement count = %d, observed %d", output.ExactMeasurementCount, len(measured))
	}
}

func TestInfeasibleSeedUsesAnUnboundedSearch(t *testing.T) {
	problem := syntheticProblem(2, 4, 6)
	problem.Justified = false
	problem.MinRatio = 0
	problem.Exact = func(rangeUTF16 Range, _ Breakpoint, _ bool) (Metrics, error) {
		width := 20.0
		if rangeUTF16.EndUTF16-rangeUTF16.StartUTF16 == 2 {
			width = 10
		}
		return Metrics{Width: width}, nil
	}
	output, err := Solve(problem)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(lineEnds(output), []int{2, 4, 6}) || !output.UsedUnboundedRetry {
		t.Fatalf("infeasible seed result: ends=%v, retry=%t", lineEnds(output), output.UsedUnboundedRetry)
	}
	for _, line := range output.Lines {
		if !line.Fit {
			t.Fatalf("selected line does not fit: %+v", line)
		}
	}
}

func TestDashHistoryRetainsMoreExpensivePredecessor(t *testing.T) {
	problem := syntheticProblem(1, 2, 3, 4, 5)
	seenTailDash, seenTailNoDash := false, false
	problem.Exact = func(rangeUTF16 Range, _ Breakpoint, previousDash bool) (Metrics, error) {
		if rangeUTF16 == (Range{3, 5}) {
			seenTailDash = seenTailDash || previousDash
			seenTailNoDash = seenTailNoDash || !previousDash
		}
		metrics := Metrics{Width: 0, Stretch: 10, Spaces: 1, ShouldJustify: true}
		switch rangeUTF16 {
		case Range{0, 1}:
			metrics.Width = 10
		case Range{0, 3}:
			metrics.Width = 8
		case Range{1, 3}, Range{3, 5}:
			metrics.Width = 10
			metrics.EndsWithDash = true
		}
		return metrics, nil
	}
	output, err := Solve(problem)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(lineEnds(output), []int{3, 5}) || !seenTailDash || !seenTailNoDash {
		t.Fatalf("dash history result: ends=%v, measured tail states=%t/%t", lineEnds(output), seenTailDash, seenTailNoDash)
	}
	assertCost(t, output.TotalCost, 4.24)
}

func TestHyphenCandidateBeatsWideWordSpacing(t *testing.T) {
	for _, pruning := range []bool{false, true} {
		t.Run(fmt.Sprintf("pruning=%t", pruning), func(t *testing.T) {
			problem := syntheticProblem(5, 9, 14)
			problem.Width = 100
			problem.EnableIncumbentPruning = pruning
			problem.EnableOverfullRetirement = true
			problem.Breakpoints[1] = Breakpoint{EndUTF16: 9, Kind: Hyphen, HyphenLeft: 5, HyphenRight: 5}
			metrics := map[Range]Metrics{
				{0, 5}:  {Width: 84, Stretch: 4, Shrink: 1.32, Spaces: 2},
				{0, 9}:  {Width: 98, Stretch: 4, Shrink: 1.32, Spaces: 2, EndsWithDash: true},
				{0, 14}: {Width: 164, Stretch: 8, Shrink: 2.64, Spaces: 4},
				{5, 9}:  {Width: 14, EndsWithDash: true},
				{5, 14}: {Width: 80, Stretch: 4, Shrink: 1.32, Spaces: 2},
				{9, 14}: {Width: 66, Stretch: 4, Shrink: 1.32, Spaces: 2},
			}
			problem.Approximate = func(r Range, _ Breakpoint) Metrics {
				m, exists := metrics[r]
				if !exists {
					t.Fatalf("unexpected approximate candidate %v", r)
				}
				return m
			}
			problem.Exact = func(r Range, _ Breakpoint, _ bool) (Metrics, error) {
				m, exists := metrics[r]
				if !exists {
					return Metrics{}, fmt.Errorf("unexpected candidate %v", r)
				}
				return m, nil
			}
			withHyphen, err := Solve(problem)
			if err != nil {
				t.Fatal(err)
			}
			problem.Breakpoints = []Breakpoint{problem.Breakpoints[0], problem.Breakpoints[2]}
			withoutHyphen, err := Solve(problem)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(lineEnds(withHyphen), []int{9, 14}) || !reflect.DeepEqual(lineEnds(withoutHyphen), []int{5, 14}) || withHyphen.Lines[0].Breakpoint.Kind != Hyphen || !withHyphen.Lines[0].Metrics.EndsWithDash {
				t.Fatalf("hyphen candidate did not compete with whole words: with=%v, without=%v", lineEnds(withHyphen), lineEnds(withoutHyphen))
			}
			for _, output := range []Output{withHyphen, withoutHyphen} {
				for _, line := range output.Lines {
					if !line.Fit {
						t.Fatal("comparison must offer two fitting layouts, not force hyphenation through overflow")
					}
				}
			}
			spacing := func(line Line) float64 { return (problem.Width - line.Metrics.Width) / float64(line.Metrics.Spaces) }
			if spacing(withHyphen.Lines[0]) != 1 || spacing(withoutHyphen.Lines[0]) != 8 || withHyphen.TotalCost >= withoutHyphen.TotalCost {
				t.Fatalf("hyphen did not reduce spacing and total cost: with=%+v, without=%+v", withHyphen, withoutHyphen)
			}
		})
	}
}

func TestMandatoryBreakResetsCandidateSearch(t *testing.T) {
	problem := syntheticProblem(2, 3, 4, 6)
	problem.Breakpoints[1].Kind = Mandatory
	problem.Justified = false
	problem.MinRatio = 0
	problem.Exact = func(rangeUTF16 Range, _ Breakpoint, _ bool) (Metrics, error) {
		if rangeUTF16.StartUTF16 < 3 && rangeUTF16.EndUTF16 > 3 {
			t.Fatalf("candidate crosses mandatory break: %v", rangeUTF16)
		}
		return Metrics{Width: 10}, nil
	}
	output, err := Solve(problem)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(lineEnds(output), []int{3, 6}) || output.Lines[0].Breakpoint.Kind != Mandatory {
		t.Fatalf("mandatory break result: %+v", output)
	}
}

func TestEqualCostsPreferLaterPredecessor(t *testing.T) {
	problem := syntheticProblem(2, 3, 4, 5)
	problem.Exact = func(rangeUTF16 Range, _ Breakpoint, _ bool) (Metrics, error) {
		width := 0.0
		switch rangeUTF16 {
		case Range{0, 2}, Range{2, 5}, Range{0, 3}, Range{3, 5}:
			width = 10
		}
		return Metrics{Width: width, Stretch: 10, Spaces: 1, ShouldJustify: true}, nil
	}
	output, err := Solve(problem)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(lineEnds(output), []int{3, 5}) {
		t.Fatalf("equal-cost line ends = %v, want [3 5]", lineEnds(output))
	}
}

func TestNonmonotoneWidthsAndRetirementRetry(t *testing.T) {
	problem := syntheticProblem(1, 2, 3)
	problem.Justified = false
	problem.MinRatio = 0
	problem.Exact = func(rangeUTF16 Range, _ Breakpoint, _ bool) (Metrics, error) {
		width := 10.0
		if rangeUTF16.StartUTF16 == 0 && rangeUTF16.EndUTF16 < 3 {
			width = 20
		}
		return Metrics{Width: width}, nil
	}
	for _, retirement := range []bool{false, true} {
		problem.EnableOverfullRetirement = retirement
		output, err := Solve(problem)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(lineEnds(output), []int{3}) || output.UsedUnboundedRetry != retirement {
			t.Fatalf("retirement=%t: ends=%v, retry=%t", retirement, lineEnds(output), output.UsedUnboundedRetry)
		}
		assertCost(t, output.TotalCost, 1)
	}
}

func TestFitDistinguishesMandatoryAndFinalStretch(t *testing.T) {
	problem := syntheticProblem(1, 2)
	problem.Breakpoints[0].Kind = Mandatory
	problem.Approximate = func(Range, Breakpoint) Metrics { return Metrics{Width: 9} }
	problem.Exact = func(Range, Breakpoint, bool) (Metrics, error) { return Metrics{Width: 9}, nil }
	output, err := Solve(problem)
	if err != nil {
		t.Fatal(err)
	}
	if output.Lines[0].Fit || !output.Lines[1].Fit {
		t.Fatalf("fit differs from mandatory/final spacing policy: %+v", output.Lines)
	}
}

func TestExactErrorsPropagate(t *testing.T) {
	problem := syntheticProblem(1)
	want := errors.New("shaping failed")
	problem.Exact = func(Range, Breakpoint, bool) (Metrics, error) { return Metrics{}, want }
	if _, err := Solve(problem); !errors.Is(err, want) {
		t.Fatalf("measurement error = %v, want %v", err, want)
	}
}

func TestMissingParagraphEndReturnsNoCompleteLayout(t *testing.T) {
	problem := syntheticProblem(1)
	problem.Breakpoints[0].Kind = Mandatory
	if _, err := Solve(problem); !errors.Is(err, ErrNoCompleteLayout) {
		t.Fatalf("incomplete layout error = %v", err)
	}
}

func TestInvalidInputsReturnExplicitErrors(t *testing.T) {
	for _, mutate := range []func(*Problem){
		func(problem *Problem) { problem.Width = math.Inf(1) },
		func(problem *Problem) { problem.FontSize = 0 },
		func(problem *Problem) { problem.Breakpoints[0].EndUTF16 = -1 },
		func(problem *Problem) { problem.Exact = nil },
	} {
		problem := syntheticProblem(1)
		mutate(&problem)
		if _, err := Solve(problem); !errors.Is(err, ErrInvalidProblem) {
			t.Fatalf("invalid problem error = %v", err)
		}
	}
	problem := syntheticProblem(1)
	problem.Exact = func(Range, Breakpoint, bool) (Metrics, error) { return Metrics{Width: math.NaN()}, nil }
	if _, err := Solve(problem); !errors.Is(err, ErrInvalidMetrics) {
		t.Fatalf("invalid exact metrics error = %v", err)
	}
	problem.Approximate = func(Range, Breakpoint) Metrics { return Metrics{Spaces: -1} }
	if _, err := Solve(problem); !errors.Is(err, ErrInvalidMetrics) {
		t.Fatalf("invalid approximate metrics error = %v", err)
	}
}
