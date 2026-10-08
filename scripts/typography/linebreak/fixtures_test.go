// This file checks the numerical translation against unchanged fixtures captured from pinned Typst.
// SPDX-License-Identifier: Apache-2.0
// Fixture adaptation derived from Encarta's KnuthPlassPinnedFixtureTests.swift; see UPSTREAM.md.
package linebreak

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"unicode/utf16"
)

const pinnedRevision = "9dfd3a08500b7896045f907433cf7b4b02434fad"

type fixtureMetric struct {
	NaturalWidthPoints float64 `json:"naturalWidthPoints"`
	StretchPoints      float64 `json:"stretchPoints"`
	ShrinkPoints       float64 `json:"shrinkPoints"`
	JustifiableCount   int     `json:"justifiableCount"`
	EndsWithDash       bool    `json:"endsWithDash"`
}

func (metric fixtureMetric) metrics(justified bool) Metrics {
	return Metrics{
		Width:         metric.NaturalWidthPoints,
		Stretch:       metric.StretchPoints,
		Shrink:        metric.ShrinkPoints,
		Spaces:        metric.JustifiableCount,
		EndsWithDash:  metric.EndsWithDash,
		ShouldJustify: justified,
	}
}

type numericalFixtureSet struct {
	UpstreamRevision string `json:"upstreamRevision"`
	Cases            []struct {
		ID                string        `json:"id"`
		UTF16Length       int           `json:"utf16Length"`
		LineWidthPoints   float64       `json:"lineWidthPoints"`
		FontSizePoints    float64       `json:"fontSizePoints"`
		Justified         bool          `json:"justified"`
		MinRatio          float64       `json:"minRatio"`
		Metric            fixtureMetric `json:"metric"`
		ExpectedEnds      []int         `json:"expectedEnds"`
		ExpectedTotalCost float64       `json:"expectedTotalCost"`
	} `json:"cases"`
}

type pathFixtureKey struct {
	StartUTF16 int
	EndUTF16   int
	Kind       string
}

type pathFixtureMetric struct {
	fixtureMetric
	StartUTF16     int     `json:"startUTF16"`
	EndUTF16       int     `json:"endUTF16"`
	Kind           string  `json:"kind"`
	Justified      bool    `json:"justified"`
	ReferenceRatio float64 `json:"referenceRatio"`
	ReferenceCost  float64 `json:"referenceCost"`
}

type optimizedPathFixture struct {
	UpstreamRevision  string  `json:"upstreamRevision"`
	Text              string  `json:"text"`
	FontSizePoints    float64 `json:"fontSizePoints"`
	LineWidthPoints   float64 `json:"lineWidthPoints"`
	IsJustified       bool    `json:"isJustified"`
	MinRatio          float64 `json:"minRatio"`
	HyphenCost        float64 `json:"hyphenCost"`
	RuntCost          float64 `json:"runtCost"`
	ExpectedTotalCost float64 `json:"expectedTotalCost"`
	Breakpoints       []struct {
		EndUTF16 int    `json:"endUTF16"`
		Kind     string `json:"kind"`
	} `json:"breakpoints"`
	CandidateMetrics []pathFixtureMetric `json:"candidateMetrics"`
	ExpectedPath     []struct {
		StartUTF16   int     `json:"startUTF16"`
		EndUTF16     int     `json:"endUTF16"`
		Cumulative   float64 `json:"cumulativeCost"`
		EndsWithDash bool    `json:"endsWithDash"`
	} `json:"expectedPath"`
}

func readFixture[T any](t *testing.T, path string) T {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture T
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func TestPinnedTypstNumericalFixtures(t *testing.T) {
	fixtures := readFixture[numericalFixtureSet](t, "testdata/TypstNumericalFixtures.json")
	if fixtures.UpstreamRevision != pinnedRevision || len(fixtures.Cases) != 4 {
		t.Fatal("unexpected numerical fixture revision or case count")
	}
	for _, fixture := range fixtures.Cases {
		t.Run(fixture.ID, func(t *testing.T) {
			metrics := fixture.Metric.metrics(fixture.Justified)
			problem := Problem{
				Length:                 fixture.UTF16Length,
				FontSize:               fixture.FontSizePoints,
				Width:                  fixture.LineWidthPoints,
				Justified:              fixture.Justified,
				HyphenCost:             DefaultHyphenCost,
				RuntCost:               DefaultRuntCost,
				MinRatio:               fixture.MinRatio,
				Breakpoints:            []Breakpoint{{EndUTF16: fixture.UTF16Length, Kind: ParagraphEnd}},
				Approximate:            func(Range, Breakpoint) Metrics { return metrics },
				Exact:                  func(Range, Breakpoint, bool) (Metrics, error) { return metrics, nil },
				EnableIncumbentPruning: true,
			}
			output, err := Solve(problem)
			if err != nil {
				t.Fatal(err)
			}
			if got := lineEnds(output); !reflect.DeepEqual(got, fixture.ExpectedEnds) {
				t.Fatalf("line ends = %v, want %v", got, fixture.ExpectedEnds)
			}
			assertCost(t, output.TotalCost, fixture.ExpectedTotalCost)
			if output.Lines[0].Fit != (fixture.ID != "hard-overfull") {
				t.Fatalf("unexpected fit for %s: %v", fixture.ID, output.Lines[0].Fit)
			}
		})
	}
}

var fixtureHyphenKind = regexp.MustCompile(`^Hyphen\((\d+), (\d+)\)$`)

func fixtureBreakpoint(end int, kind string) (Breakpoint, error) {
	breakpoint := Breakpoint{EndUTF16: end}
	switch kind {
	case "Normal":
		breakpoint.Kind = Opportunity
	case "Mandatory":
		breakpoint.Kind = ParagraphEnd
	default:
		parts := fixtureHyphenKind.FindStringSubmatch(kind)
		if len(parts) != 3 {
			return Breakpoint{}, fmt.Errorf("invalid fixture breakpoint %q", kind)
		}
		left, err := strconv.ParseUint(parts[1], 10, 8)
		if err != nil {
			return Breakpoint{}, err
		}
		right, err := strconv.ParseUint(parts[2], 10, 8)
		if err != nil {
			return Breakpoint{}, err
		}
		breakpoint.Kind = Hyphen
		breakpoint.HyphenLeft = uint8(left)
		breakpoint.HyphenRight = uint8(right)
	}
	return breakpoint, nil
}

func fixtureKindName(breakpoint Breakpoint) string {
	switch breakpoint.Kind {
	case Opportunity:
		return "Normal"
	case Mandatory, ParagraphEnd:
		return "Mandatory"
	case Hyphen:
		return fmt.Sprintf("Hyphen(%d, %d)", breakpoint.HyphenLeft, breakpoint.HyphenRight)
	default:
		panic("invalid fixture kind")
	}
}

func TestPinnedTypstOptimizedParagraph(t *testing.T) {
	fixture := readFixture[optimizedPathFixture](t, "testdata/TypstOptimizedPathFixture.json")
	if fixture.UpstreamRevision != pinnedRevision || len(fixture.CandidateMetrics) != 136 {
		t.Fatal("unexpected path fixture revision or candidate count")
	}
	length := len(utf16.Encode([]rune(fixture.Text)))
	if length != fixture.ExpectedPath[len(fixture.ExpectedPath)-1].EndUTF16 {
		t.Fatal("fixture path does not cover its source text")
	}
	metricsByKey := make(map[pathFixtureKey]pathFixtureMetric, len(fixture.CandidateMetrics))
	for _, metric := range fixture.CandidateMetrics {
		key := pathFixtureKey{metric.StartUTF16, metric.EndUTF16, metric.Kind}
		if _, exists := metricsByKey[key]; exists {
			t.Fatalf("duplicate fixture candidate %v", key)
		}
		metricsByKey[key] = metric
	}
	problem := Problem{
		Length:     length,
		FontSize:   fixture.FontSizePoints,
		Width:      fixture.LineWidthPoints,
		Justified:  fixture.IsJustified,
		HyphenCost: fixture.HyphenCost,
		RuntCost:   fixture.RuntCost,
		MinRatio:   fixture.MinRatio,
	}
	for _, item := range fixture.Breakpoints {
		breakpoint, err := fixtureBreakpoint(item.EndUTF16, item.Kind)
		if err != nil {
			t.Fatal(err)
		}
		problem.Breakpoints = append(problem.Breakpoints, breakpoint)
	}
	lookup := func(rangeUTF16 Range, breakpoint Breakpoint) (pathFixtureMetric, error) {
		key := pathFixtureKey{rangeUTF16.StartUTF16, rangeUTF16.EndUTF16, fixtureKindName(breakpoint)}
		metric, exists := metricsByKey[key]
		if !exists {
			return pathFixtureMetric{}, fmt.Errorf("fixture omitted candidate %v", key)
		}
		return metric, nil
	}
	problem.Approximate = func(rangeUTF16 Range, breakpoint Breakpoint) Metrics {
		metric, err := lookup(rangeUTF16, breakpoint)
		if err != nil {
			t.Fatal(err)
		}
		return metric.fixtureMetric.metrics(false)
	}
	problem.Exact = func(rangeUTF16 Range, breakpoint Breakpoint, _ bool) (Metrics, error) {
		metric, err := lookup(rangeUTF16, breakpoint)
		if err != nil {
			return Metrics{}, err
		}
		return metric.fixtureMetric.metrics(metric.Justified), nil
	}
	for _, pruning := range []bool{false, true} {
		t.Run(fmt.Sprintf("pruning=%t", pruning), func(t *testing.T) {
			problem.EnableIncumbentPruning = pruning
			output, err := Solve(problem)
			if err != nil {
				t.Fatal(err)
			}
			if len(output.Lines) != len(fixture.ExpectedPath) {
				t.Fatalf("got %d lines, want %d", len(output.Lines), len(fixture.ExpectedPath))
			}
			for index, line := range output.Lines {
				want := fixture.ExpectedPath[index]
				if line.StartUTF16 != want.StartUTF16 || line.EndUTF16 != want.EndUTF16 || line.Metrics.EndsWithDash != want.EndsWithDash {
					t.Fatalf("line %d = %+v, want source range [%d,%d), dash=%t", index, line, want.StartUTF16, want.EndUTF16, want.EndsWithDash)
				}
				metric, err := lookup(Range{line.StartUTF16, line.EndUTF16}, line.Breakpoint)
				if err != nil {
					t.Fatal(err)
				}
				if line.Metrics != metric.fixtureMetric.metrics(metric.Justified) {
					t.Fatalf("selected metrics differ for line %d", index)
				}
			}
			assertCost(t, output.TotalCost, fixture.ExpectedTotalCost)
		})
	}
}

func lineEnds(output Output) []int {
	ends := make([]int, len(output.Lines))
	for index, line := range output.Lines {
		ends[index] = line.EndUTF16
	}
	return ends
}

func assertCost(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > math.Max(1e-9, math.Abs(want)*1e-12) {
		t.Fatalf("cost = %.17g, want %.17g", got, want)
	}
}
