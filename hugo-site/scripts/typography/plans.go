// Generate blog and explicit paragraph plans beside the HTML transform so preview and release use identical font geometry.
package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"github.com/spockz/alessandrovermeulen.me/hugo-site/scripts/typography/linebreak"
	"golang.org/x/image/math/fixed"
	"golang.org/x/net/html"
)

const plannedFontSize = 16.0
const plannedColumnWidth = 576.0
const minimumPlannedWidth = 240
const plannedCodeSize = 14.0
const plannedCodePadding = 5.0

type storedLine struct {
	Start   int     `json:"start"`
	End     int     `json:"end"`
	Natural float64 `json:"natural"`
	Spaces  int     `json:"spaces"`
	Hyphen  bool    `json:"hyphen,omitempty"`
	Ragged  bool    `json:"ragged,omitempty"`
}

type storedPlans struct {
	Lines  []storedLine   `json:"lines"`
	Paths  [][]int        `json:"paths"`
	Widths map[string]int `json:"widths"`
}

//go:embed font/ArticleSans-Regular.ttf
var plannedFont []byte

//go:embed font/ArticleSans-Bold.ttf
var plannedBoldFont []byte

//go:embed font/ArticleSans-Italic.ttf
var plannedItalicFont []byte

//go:embed font/ArticleSans-BoldItalic.ttf
var plannedBoldItalicFont []byte

//go:embed font/ArticleMono-Regular.ttf
var plannedCodeFont []byte

type plannedFaces struct {
	prose      *font.Face
	code       *font.Face
	bold       *font.Face
	italic     *font.Face
	boldItalic *font.Face
}

type inlineRun struct {
	text       string
	path       []*html.Node
	code       *html.Node
	format     textFormat
	atom       *html.Node
	image      *html.Node
	imageWidth float64
}

type measuredRun struct {
	start      int
	end        int
	code       *html.Node
	format     textFormat
	atom       *html.Node
	image      *html.Node
	imageWidth float64
	boxes      []*html.Node
}

type paragraphMap struct {
	text         []rune
	utf16Offsets []int
	runeOffsets  map[int]int
	byteOffsets  map[int]int
	length       int
}

type paragraphMeasure struct {
	faces         plannedFaces
	shaper        shaping.HarfbuzzShaper
	mapText       paragraphMap
	runs          []measuredRun
	codeAt        []bool
	prefix        []float64
	stretchPrefix []float64
	cache         map[linebreak.Range]linebreak.Metrics
	size          float64
	boxFormats    map[*html.Node]textFormat
}

func planDocument(root *html.Node) (bool, error) {
	var faces plannedFaces
	changed := false
	var visit func(*html.Node, []int) error
	visit = func(node *html.Node, widths []int) error {
		if hasClass(node, "typography-prose") && !hasAttribute(node, "data-plan-lines") {
			return nil
		}
		if node.Type == html.ElementNode && ((isProtectedElement(node.Data) && node.Data != "kbd" && node.Data != "samp") || hasClass(node, "not-prose")) {
			return nil
		}
		if node.Type == html.ElementNode && hasAttribute(node, "data-plan-lines") {
			if hasAttribute(node, "data-plan-generated") {
				return nil
			}
			switch attribute(node, "data-plan-lines") {
			case "576":
				widths = []int{int(plannedColumnWidth)}
			case "responsive":
				widths = make([]int, int(plannedColumnWidth)-minimumPlannedWidth+1)
				for i := range widths {
					widths[i] = minimumPlannedWidth + i
				}
			default:
				return fmt.Errorf("unsupported planned column width %q", attribute(node, "data-plan-lines"))
			}
			node.Attr = append(node.Attr, html.Attribute{Key: "data-plan-generated", Val: "true"})
			changed = true
			changed = prepareProseBlocks(node) || changed
		}
		if (node.Data == "td" || node.Data == "th") && len(widths) > 1 && widths[0] > 80 {
			widths = make([]int, int(plannedColumnWidth)-80+1)
			for i := range widths {
				widths[i] = 80 + i
			}
		}
		if len(widths) > 0 && node.Type == html.ElementNode && (isProseBlockCandidate(node) || hasClass(node, "planned-prose-block")) {
			runs, ok := paragraphRuns(node)
			if !ok {
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					if err := visit(child, widths); err != nil {
						return err
					}
				}
				return nil
			}
			if faces.prose == nil {
				var err error
				faces.prose, err = font.ParseTTF(bytes.NewReader(plannedFont))
				if err != nil {
					return fmt.Errorf("load pinned article font: %w", err)
				}
				faces.code, err = font.ParseTTF(bytes.NewReader(plannedCodeFont))
				if err != nil {
					return fmt.Errorf("load pinned code font: %w", err)
				}
				for _, variant := range []struct {
					data   []byte
					target **font.Face
				}{{plannedBoldFont, &faces.bold}, {plannedItalicFont, &faces.italic}, {plannedBoldItalicFont, &faces.boldItalic}} {
					*variant.target, err = font.ParseTTF(bytes.NewReader(variant.data))
					if err != nil {
						return fmt.Errorf("load pinned prose variant: %w", err)
					}
				}
			}
			planned, err := planParagraph(node, runs, faces, widths)
			if err != nil {
				return err
			}
			changed = changed || planned
			return nil
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := visit(child, widths); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(root, nil); err != nil {
		return false, err
	}
	return changed, nil
}

func paragraphRuns(node *html.Node) ([]inlineRun, bool) {
	var runs []inlineRun
	base, ok := blockFormat(node)
	if !ok {
		return nil, false
	}
	var collect func(*html.Node, []*html.Node, textFormat, *html.Node) bool
	collect = func(n *html.Node, path []*html.Node, format textFormat, atom *html.Node) bool {
		if n.Type == html.TextNode {
			var code *html.Node
			for _, ancestor := range path {
				if ancestor.Data == "code" {
					code = ancestor
				}
			}
			runs = append(runs, inlineRun{text: n.Data, path: path, code: code, format: format.textOnly(), atom: atom})
			return true
		}
		if n.Type != html.ElementNode || (n != node && !inlineElement(n.Data)) {
			return false
		}
		if n != node && hasClass(n, "planned-prose-block") {
			return false
		}
		if n != node {
			if n.Data == "br" {
				if atom != nil {
					return false
				}
				runs = append(runs, inlineRun{text: "\n", path: append(append([]*html.Node(nil), path...), n), format: format, image: n})
				return true
			}
			var ok bool
			format, ok = formatNode(n, format)
			if !ok {
				return false
			}
			if n.Data == "img" {
				width, valid := cssLength(attribute(n, "width"), format.Size)
				if !valid || width <= 0 || format.edges() != 0 {
					return false
				}
				encoded, err := json.Marshal(map[string]float64{"width": width})
				if err != nil {
					return false
				}
				setAttribute(n, "data-plan-inline", string(encoded))
				setAttribute(n, "data-plan-atom", "true")
				runs = append(runs, inlineRun{text: "\ufffc", path: append(append([]*html.Node(nil), path...), n), format: format, image: n, imageWidth: width, atom: n})
				return true
			}
			if err := annotateFormat(n, format); err != nil {
				return false
			}
			if atom == nil && hasAttribute(n, "data-plan-atom") {
				atom = n
			}
			path = append(append([]*html.Node(nil), path...), n)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if !collect(child, path, format, atom) {
				return false
			}
		}
		return true
	}
	if !collect(node, nil, base, nil) {
		return nil, false
	}
	seenText, lastSpace := false, false
	for i := range runs {
		if runs[i].code != nil || runs[i].image != nil {
			seenText, lastSpace = true, false
			continue
		}
		var normalized strings.Builder
		for _, r := range runs[i].text {
			if r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == '\f' {
				if seenText && !lastSpace {
					normalized.WriteByte(' ')
					lastSpace = true
				}
				continue
			}
			normalized.WriteRune(r)
			seenText, lastSpace = true, false
		}
		runs[i].text = normalized.String()
	}
	for i := len(runs) - 1; i >= 0; i-- {
		runs[i].text = strings.TrimRight(runs[i].text, " ")
		if runs[i].text != "" {
			break
		}
	}
	return runs, seenText
}

func mapParagraph(text string) paragraphMap {
	m := paragraphMap{text: []rune(text), runeOffsets: map[int]int{}, byteOffsets: map[int]int{}}
	index, offset := 0, 0
	for byteIndex, r := range text {
		m.runeOffsets[offset], m.byteOffsets[offset] = index, byteIndex
		m.utf16Offsets = append(m.utf16Offsets, offset)
		offset += utf16.RuneLen(r)
		index++
	}
	m.runeOffsets[offset], m.byteOffsets[offset] = index, len(text)
	m.utf16Offsets = append(m.utf16Offsets, offset)
	m.length = offset
	return m
}

func planParagraph(node *html.Node, runs []inlineRun, faces plannedFaces, widths []int) (bool, error) {
	base, ok := blockFormat(node)
	if !ok {
		return false, nil
	}
	var joined strings.Builder
	for _, run := range runs {
		joined.WriteString(run.text)
	}
	text := joined.String()
	if strings.Contains(text, "\n\n") || strings.HasPrefix(text, "\n") || strings.HasSuffix(text, "\n") {
		return false, nil
	}
	mapped := mapParagraph(text)
	for _, run := range runs {
		if run.image != nil {
			continue
		}
		face := faces.face(run.format)
		for _, r := range transformedText(run.text, run.format) {
			if r == '\u00ad' {
				continue
			}
			if unicode.IsLetter(r) && !unicode.In(r, unicode.Latin) {
				return false, nil
			}
			if _, supported := face.NominalGlyph(r); !supported {
				return false, nil
			}
		}
	}
	if len(mapped.text) > 5000 {
		return false, nil
	}
	measure := paragraphMeasure{faces: faces, mapText: mapped, codeAt: make([]bool, len(mapped.text)), cache: map[linebreak.Range]linebreak.Metrics{}, size: base.Size, boxFormats: map[*html.Node]textFormat{}}
	offset := 0
	for _, run := range runs {
		end := offset + len([]rune(run.text))
		if end == offset {
			continue
		}
		var boxes []*html.Node
		for _, box := range run.path {
			if box.Data == "img" {
				continue
			}
			var format textFormat
			if encoded := attribute(box, "data-plan-inline"); encoded != "" {
				if err := json.Unmarshal([]byte(encoded), &format); err != nil {
					return false, fmt.Errorf("read inline contract: %w", err)
				}
				if format.edges() != 0 {
					boxes = append(boxes, box)
					measure.boxFormats[box] = format
				}
			}
		}
		if len(measure.runs) > 0 && measure.runs[len(measure.runs)-1].image == nil && measure.runs[len(measure.runs)-1].atom == run.atom && measure.runs[len(measure.runs)-1].format == run.format && run.image == nil && len(boxes) == 0 {
			measure.runs[len(measure.runs)-1].end = end
		} else {
			measure.runs = append(measure.runs, measuredRun{start: offset, end: end, code: run.code, format: run.format, atom: run.atom, image: run.image, imageWidth: run.imageWidth, boxes: boxes})
		}
		if run.atom != nil {
			for i := offset; i < end; i++ {
				measure.codeAt[i] = true
			}
		}
		offset = end
	}
	measure.preparePrefix()
	breakpoints := paragraphBreakpoints(mapped, measure.codeAt)
	problem := linebreak.Problem{
		Length: mapped.length, FontSize: base.Size,
		Justified: true, HyphenCost: linebreak.DefaultHyphenCost, RuntCost: linebreak.DefaultRuntCost,
		MinRatio: -1, Breakpoints: breakpoints, Approximate: measure.approximate,
		Exact: measure.exact, EnableIncumbentPruning: true, EnableOverfullRetirement: true,
	}
	result, plans, err := composeWidths(problem, widths)
	if err != nil {
		return false, fmt.Errorf("compose paragraph: %w", err)
	}
	if len(result.Lines) == 0 {
		return false, nil
	}
	defaultWidth := 0.0
	for _, width := range widths {
		if _, available := plans.Widths[strconv.Itoa(width)]; available {
			defaultWidth = float64(width)
		}
	}
	serialized, err := json.Marshal(plans)
	if err != nil {
		return false, fmt.Errorf("serialize paragraph plans: %w", err)
	}
	for node.FirstChild != nil {
		node.RemoveChild(node.FirstChild)
	}
	node.Attr = append(node.Attr,
		html.Attribute{Key: "data-plan-width", Val: number(defaultWidth)},
		html.Attribute{Key: "data-plan-size", Val: number(base.Size)},
		html.Attribute{Key: "data-plan-weight", Val: base.Weight},
		html.Attribute{Key: "data-plan-style", Val: base.Style},
		html.Attribute{Key: "data-line-plans", Val: string(serialized)})
	containers := map[*html.Node]*html.Node{}
	for i, line := range result.Lines {
		from, to := measure.visibleRange(linebreak.Range{StartUTF16: line.StartUTF16, EndUTF16: line.EndUTF16})
		start, end := mapped.byteOffsets[mapped.utf16Offsets[from]], mapped.byteOffsets[mapped.utf16Offsets[to]]
		spacing := 0.0
		final := i == len(result.Lines)-1 || line.Breakpoint.Kind == linebreak.Mandatory
		if (!final || line.Metrics.Width > defaultWidth) && line.Metrics.Spaces > 0 {
			spacing = (defaultWidth - line.Metrics.Width) / float64(line.Metrics.Spaces)
		}
		pieces := appendSlice(node, runs, start, end, containers, "planned-line")
		if len(pieces) == 0 {
			return false, fmt.Errorf("composed line has no visible text")
		}
		for _, piece := range pieces {
			piece.Attr = append(piece.Attr,
				html.Attribute{Key: "data-line", Val: strconv.Itoa(i)},
				html.Attribute{Key: "style", Val: "--plan-spacing:" + number(spacing) + "px"})
		}
		first := pieces[0]
		first.Attr = append(first.Attr,
			html.Attribute{Key: "data-natural", Val: number(line.Metrics.Width)},
			html.Attribute{Key: "data-spaces", Val: strconv.Itoa(line.Metrics.Spaces)})
		if final {
			first.Attr = append(first.Attr, html.Attribute{Key: "data-final", Val: "true"})
		}
		last := pieces[len(pieces)-1]
		if line.Breakpoint.Kind == linebreak.Hyphen {
			hyphen := &html.Node{Type: html.ElementNode, Data: "span", Attr: []html.Attribute{{Key: "class", Val: "planned-hyphen"}, {Key: "aria-hidden", Val: "true"}}}
			last.AppendChild(hyphen)
		}
		if i != len(result.Lines)-1 {
			container := last.Parent
			for current := last.Parent; current != nil && current != node; current = current.Parent {
				if hasAttribute(current, "data-plan-atom") {
					container = current.Parent
				}
			}
			container.AppendChild(&html.Node{Type: html.ElementNode, Data: "span", Attr: []html.Attribute{{Key: "class", Val: "planned-break"}, {Key: "aria-hidden", Val: "true"}}})
		}
		gap := text[end:mapped.byteOffsets[line.EndUTF16]]
		if gap != "" {
			appendSlice(node, runs, end, mapped.byteOffsets[line.EndUTF16], containers, "planned-gap")
		}
	}
	return true, nil
}

func composeWidths(problem linebreak.Problem, widths []int) (linebreak.Output, storedPlans, error) {
	plans := storedPlans{Widths: map[string]int{}}
	lineIDs := map[linebreak.Range]int{}
	pathIDs := map[string]int{}
	var widest linebreak.Output
	for _, width := range widths {
		problem.Width = float64(width)
		result, err := linebreak.Solve(problem)
		if err != nil {
			return linebreak.Output{}, storedPlans{}, fmt.Errorf("width %d: %w", width, err)
		}
		fits := true
		for _, line := range result.Lines {
			// HTML breaks end ragged segments rather than requiring the kernel's mandatory-line stretch.
			lineFits := line.Fit || (line.Breakpoint.Kind == linebreak.Mandatory && line.Metrics.Width <= float64(width))
			if !lineFits || line.Metrics.Spaces == 0 && line.Metrics.Width > float64(width) {
				fits = false
				break
			}
		}
		if !fits {
			continue
		}
		path := make([]int, len(result.Lines))
		var key strings.Builder
		for i, line := range result.Lines {
			r := linebreak.Range{StartUTF16: line.StartUTF16, EndUTF16: line.EndUTF16}
			id, exists := lineIDs[r]
			if !exists {
				id = len(plans.Lines)
				lineIDs[r] = id
				plans.Lines = append(plans.Lines, storedLine{Start: line.StartUTF16, End: line.EndUTF16,
					Natural: math.Round(line.Metrics.Width*10000) / 10000, Spaces: line.Metrics.Spaces, Hyphen: line.Breakpoint.Kind == linebreak.Hyphen, Ragged: line.Breakpoint.Kind == linebreak.Mandatory})
			}
			path[i] = id
			key.WriteString(strconv.Itoa(id))
			key.WriteByte(',')
		}
		id, exists := pathIDs[key.String()]
		if !exists {
			id = len(plans.Paths)
			pathIDs[key.String()] = id
			plans.Paths = append(plans.Paths, path)
		}
		plans.Widths[strconv.Itoa(width)] = id
		widest = result
	}
	return widest, plans, nil
}

func appendSlice(parent *html.Node, runs []inlineRun, start, end int, containers map[*html.Node]*html.Node, class string) []*html.Node {
	var pieces []*html.Node
	offset := 0
	for _, run := range runs {
		from, to := max(0, start-offset), min(len(run.text), end-offset)
		if from < to {
			container := parent
			for _, original := range run.path {
				if original.Data == "img" || original.Data == "br" {
					break
				}
				clone := containers[original]
				if clone == nil {
					clone = &html.Node{Type: original.Type, Data: original.Data, Attr: append([]html.Attribute(nil), original.Attr...)}
					container.AppendChild(clone)
					containers[original] = clone
				}
				container = clone
			}
			piece := &html.Node{Type: html.ElementNode, Data: "span", Attr: []html.Attribute{{Key: "class", Val: class}}}
			if run.image != nil {
				piece.AppendChild(&html.Node{Type: html.ElementNode, Data: run.image.Data, Attr: append([]html.Attribute(nil), run.image.Attr...)})
			} else {
				piece.AppendChild(&html.Node{Type: html.TextNode, Data: run.text[from:to]})
			}
			container.AppendChild(piece)
			pieces = append(pieces, piece)
		}
		offset += len(run.text)
	}
	return pieces
}

func paragraphBreakpoints(mapped paragraphMap, codeAt []bool) []linebreak.Breakpoint {
	var points []linebreak.Breakpoint
	for index, r := range mapped.text {
		if r == '\n' {
			points = append(points, linebreak.Breakpoint{EndUTF16: mapped.utf16Offsets[index+1], Kind: linebreak.Mandatory})
			continue
		}
		if codeAt[index] {
			continue
		}
		if index == len(mapped.text)-1 {
			break
		}
		switch r {
		case ' ', '-', '\u2010':
			points = append(points, linebreak.Breakpoint{EndUTF16: mapped.utf16Offsets[index+1], Kind: linebreak.Opportunity})
		case '\u00ad':
			left, right := 0, 0
			for i := index - 1; i >= 0 && (unicode.IsLetter(mapped.text[i]) || mapped.text[i] == '\u00ad'); i-- {
				if mapped.text[i] != '\u00ad' {
					left++
				}
			}
			for i := index + 1; i < len(mapped.text) && (unicode.IsLetter(mapped.text[i]) || mapped.text[i] == '\u00ad'); i++ {
				if mapped.text[i] != '\u00ad' {
					right++
				}
			}
			points = append(points, linebreak.Breakpoint{EndUTF16: mapped.utf16Offsets[index+1], Kind: linebreak.Hyphen, HyphenLeft: uint8(min(left, 5)), HyphenRight: uint8(min(right, 5))})
		}
	}
	return append(points, linebreak.Breakpoint{EndUTF16: mapped.length, Kind: linebreak.ParagraphEnd})
}

func (f plannedFaces) face(format textFormat) *font.Face {
	if format.Family == "article mono" {
		return f.code
	}
	if format.Style == "italic" {
		if format.Weight == "700" {
			return f.boldItalic
		}
		return f.italic
	}
	if format.Weight == "700" {
		return f.bold
	}
	return f.prose
}

func (m *paragraphMeasure) shape(text []rune, format textFormat) shaping.Output {
	face := m.faces.face(format)
	return m.shaper.ShapeNoExtents(shaping.Input{Text: text, RunEnd: len(text), Direction: di.DirectionLTR,
		Face: face, Size: fixed.I(1024), Script: language.Latin, Language: language.NewLanguage("en-GB")})
}

func (m *paragraphMeasure) preparePrefix() {
	m.prefix = make([]float64, len(m.mapText.text)+1)
	m.stretchPrefix = make([]float64, len(m.mapText.text)+1)
	spaceStretch := map[textFormat]float64{}
	boxRanges := map[*html.Node][2]int{}
	for _, run := range m.runs {
		for _, box := range run.boxes {
			bounds, exists := boxRanges[box]
			if !exists {
				bounds[0] = run.start
			}
			bounds[1] = run.end
			boxRanges[box] = bounds
		}
		if run.image != nil {
			m.prefix[run.end] += run.imageWidth
			continue
		}
		for i := run.start; i < run.end; i++ {
			if m.mapText.text[i] != ' ' || m.codeAt[i] {
				continue
			}
			stretch, exists := spaceStretch[run.format]
			if !exists {
				space := m.shape([]rune{' '}, run.format)
				// Half a natural space makes loose spacing compete with discretionary hyphens.
				stretch = float64(space.Advance) / 64 * run.format.Size / 1024 * 0.5
				spaceStretch[run.format] = stretch
			}
			m.stretchPrefix[i+1] = stretch
		}
		text := []rune(transformedText(string(m.mapText.text[run.start:run.end]), run.format))
		shaped := m.shape(text, run.format)
		for _, glyph := range shaped.Glyphs {
			m.prefix[run.start+glyph.TextIndex()+1] += float64(glyph.Advance) / 64 * run.format.Size / 1024
		}
		for i, r := range text {
			if r != '\u00ad' && !unicode.IsMark(r) {
				m.prefix[run.start+i+1] += run.format.LetterSpacing
			}
		}
	}
	for box, bounds := range boxRanges {
		format := m.boxFormats[box]
		m.prefix[bounds[0]+1] += format.PaddingLeft + format.BorderLeftWidth + format.MarginLeft
		m.prefix[bounds[1]] += format.PaddingRight + format.BorderRightWidth + format.MarginRight
	}
	for i := 1; i < len(m.prefix); i++ {
		m.prefix[i] += m.prefix[i-1]
		m.stretchPrefix[i] += m.stretchPrefix[i-1]
	}
}

func (m *paragraphMeasure) visibleRange(r linebreak.Range) (int, int) {
	start, end := m.mapText.runeOffsets[r.StartUTF16], m.mapText.runeOffsets[r.EndUTF16]
	for start < end && m.mapText.text[start] == ' ' && !m.codeAt[start] {
		start++
	}
	for end > start && (m.mapText.text[end-1] == ' ' || m.mapText.text[end-1] == '\u00ad' || m.mapText.text[end-1] == '\n') && !m.codeAt[end-1] {
		end--
	}
	return start, end
}

func (m *paragraphMeasure) approximate(r linebreak.Range, point linebreak.Breakpoint) linebreak.Metrics {
	start, end := m.visibleRange(r)
	text := m.mapText.text[start:end]
	width := m.prefix[end] - m.prefix[start]
	if point.Kind == linebreak.Hyphen {
		width += m.size * 0.33
	}
	return m.metrics(start, end, text, width, point.Kind == linebreak.Hyphen)
}

func (m *paragraphMeasure) exact(r linebreak.Range, point linebreak.Breakpoint, _ bool) (linebreak.Metrics, error) {
	if cached, ok := m.cache[r]; ok {
		return cached, nil
	}
	start, end := m.visibleRange(r)
	width := 0.0
	countedBoxes := map[*html.Node]bool{}
	for _, run := range m.runs {
		from, to := max(start, run.start), min(end, run.end)
		if from >= to {
			continue
		}
		if run.code != nil && (from != run.start || to != run.end) {
			return linebreak.Metrics{}, fmt.Errorf("line splits an atomic code box")
		}
		for _, box := range run.boxes {
			if !countedBoxes[box] {
				width += m.boxFormats[box].edges()
				countedBoxes[box] = true
			}
		}
		if run.image != nil {
			width += run.imageWidth
			continue
		}
		text := []rune(transformedText(strings.ReplaceAll(string(m.mapText.text[from:to]), "\u00ad", ""), run.format))
		if to == end && point.Kind == linebreak.Hyphen {
			text = append(text, '-')
		}
		shaped := m.shape(text, run.format)
		width += float64(shaped.Advance) / 64 * run.format.Size / 1024
		for _, r := range text {
			if !unicode.IsMark(r) {
				width += run.format.LetterSpacing
			}
		}
	}
	metrics := m.metrics(start, end, m.mapText.text[start:end], width, point.Kind == linebreak.Hyphen)
	m.cache[r] = metrics
	return metrics, nil
}

func (m *paragraphMeasure) metrics(start, end int, text []rune, width float64, hyphen bool) linebreak.Metrics {
	spaces := 0
	for i := start; i < end; i++ {
		if m.mapText.text[i] == ' ' && !m.codeAt[i] {
			spaces++
		}
	}
	endsDash := hyphen || len(text) > 0 && strings.ContainsRune("-–—", text[len(text)-1])
	stretch := m.stretchPrefix[end] - m.stretchPrefix[start]
	return linebreak.Metrics{Width: width, Stretch: stretch, Shrink: float64(spaces) * m.size * 0.0825, Spaces: spaces, EndsWithDash: endsDash}
}

func number(value float64) string {
	return strconv.FormatFloat(math.Round(value*10000)/10000, 'f', -1, 64)
}
