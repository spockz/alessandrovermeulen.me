// This file keeps the numerical optimizer independent of shaping, segmentation, and rendering.
// SPDX-License-Identifier: Apache-2.0
// Derived from the Typst contributors' paragraph line-breaking implementation via Encarta.
// Modified: translated from Swift to Go; retained only the numerical kernel contract.
package linebreak

import "errors"

type Kind uint8

const (
	Opportunity Kind = iota
	Mandatory
	Hyphen
	ParagraphEnd
)

const (
	DefaultHyphenCost = 135.0
	DefaultRuntCost   = 100.0
)

var (
	ErrNoCompleteLayout = errors.New("linebreak: no complete layout")
	ErrInvalidProblem   = errors.New("linebreak: invalid problem")
	ErrInvalidMetrics   = errors.New("linebreak: invalid metrics")
)

type Breakpoint struct {
	EndUTF16    int
	Kind        Kind
	HyphenLeft  uint8
	HyphenRight uint8
}

func (kind Kind) isMandatory() bool {
	return kind == Mandatory || kind == ParagraphEnd
}

type Range struct {
	StartUTF16 int
	EndUTF16   int
}

type Metrics struct {
	Width         float64
	Stretch       float64
	Shrink        float64
	Spaces        int
	EndsWithDash  bool
	ShouldJustify bool
}

type Problem struct {
	Length                   int
	FontSize                 float64
	Width                    float64
	Justified                bool
	HyphenCost               float64
	RuntCost                 float64
	MinRatio                 float64
	Breakpoints              []Breakpoint
	Approximate              func(Range, Breakpoint) Metrics
	Exact                    func(Range, Breakpoint, bool) (Metrics, error)
	EnableIncumbentPruning   bool
	EnableOverfullRetirement bool
}

type Line struct {
	StartUTF16 int
	EndUTF16   int
	Breakpoint Breakpoint
	Metrics    Metrics
	Fit        bool
}

type Output struct {
	Lines                 []Line
	TotalCost             float64
	ExactMeasurementCount int
	UsedUnboundedRetry    bool
}
