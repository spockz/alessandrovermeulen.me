// Check the rendered contract: continuous fallback text, intact links, supported geometry, and idempotent processing.
package main

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestPlanPreservesContinuousSourceAndSingleLink(t *testing.T) {
	for _, mode := range []string{"576", "responsive"} {
		t.Run(mode, func(t *testing.T) { testPlanPreservesContinuousSourceAndSingleLink(t, mode) })
	}
}

func testPlanPreservesContinuousSourceAndSingleLink(t *testing.T, mode string) {
	label := strings.Repeat("asynchronous computation across services and application boundaries ", 6)
	source := []byte(`<main data-hyphenate-lang="en-GB" data-plan-lines="` + mode + `"><p>Unicode café and naïve readers. <a href="https://example.com/?x=1&amp;y=2">` + label + `</a> End of the paragraph.</p></main>`)
	output, changed, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected a generated plan")
	}
	parsed := parseDocument(t, output)
	original := parseDocument(t, source)
	if got, want := normalizeSource(parsed), normalizeSource(original); got != want {
		t.Fatalf("fallback wording changed:\ngot %q\nwant %q", got, want)
	}
	links := elements(parsed, "a")
	if len(links) != 1 || attribute(links[0], "href") != "https://example.com/?x=1&y=2" {
		t.Fatalf("link identity changed: %d links", len(links))
	}
	if normalizeSource(links[0]) != strings.TrimSpace(label) {
		t.Fatal("link label changed across planned lines")
	}
	paragraph := elements(parsed, "p")[0]
	if attribute(paragraph, "data-plan-width") != "576" || attribute(paragraph, "data-plan-size") != "16" {
		t.Fatal("missing pinned geometry")
	}
	var breaks []*html.Node
	for _, span := range elements(paragraph, "span") {
		if hasClass(span, "planned-break") {
			breaks = append(breaks, span)
		}
	}
	if len(breaks) < 2 {
		t.Fatal("expected linked text to span several chosen lines")
	}
	for _, span := range elements(paragraph, "span") {
		if !hasAttribute(span, "data-natural") {
			continue
		}
		width, err := strconv.ParseFloat(attribute(span, "data-natural"), 64)
		if err != nil {
			t.Fatal(err)
		}
		spaces, err := strconv.Atoi(attribute(span, "data-spaces"))
		if err != nil {
			t.Fatal(err)
		}
		spacing, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(attribute(span, "style"), "--plan-spacing:"), "px"), 64)
		if err != nil {
			t.Fatal(err)
		}
		if width+float64(spaces)*spacing > 576.001 || spacing < -1.3201 {
			t.Fatalf("planned line cannot fit: natural %v, spaces %v, spacing %v", width, spaces, spacing)
		}
	}
	second, changed, err := transformHTML(output)
	if err != nil {
		t.Fatal(err)
	}
	if changed || !bytes.Equal(second, output) {
		t.Fatal("reprocessing changed an existing plan")
	}
}

func TestResponsivePlansCoverFluidWidthsWithFittingLines(t *testing.T) {
	source := []byte(`<main data-hyphenate-lang="en-GB" data-plan-lines="responsive"><p>` + strings.Repeat("Asynchronous computation across services and application boundaries. ", 4) + `</p></main>`)
	output, _, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	paragraph := elements(parseDocument(t, output), "p")[0]
	var plans storedPlans
	if err := json.Unmarshal([]byte(attribute(paragraph, "data-line-plans")), &plans); err != nil {
		t.Fatal(err)
	}
	if len(plans.Widths) != 337 || len(plans.Paths) >= len(plans.Widths) {
		t.Fatalf("expected all integer widths with shared paths: %d widths, %d paths", len(plans.Widths), len(plans.Paths))
	}
	var expectedEnd int
	for _, span := range elements(paragraph, "span") {
		if hasAttribute(span, "data-final") {
			path := plans.Paths[plans.Widths["576"]]
			expectedEnd = plans.Lines[path[len(path)-1]].End
		}
	}
	for width := 240; width <= 576; width++ {
		pathID, exists := plans.Widths[strconv.Itoa(width)]
		if !exists || pathID < 0 || pathID >= len(plans.Paths) {
			t.Fatalf("width %d has no valid path", width)
		}
		path := plans.Paths[pathID]
		end := 0
		for index, id := range path {
			if id < 0 || id >= len(plans.Lines) {
				t.Fatalf("width %d references missing line %d", width, id)
			}
			line := plans.Lines[id]
			if line.Start != end || line.End <= line.Start {
				t.Fatalf("width %d has incomplete source ranges", width)
			}
			end = line.End
			spacing := 0.0
			if index != len(path)-1 || line.Natural > float64(width) {
				if line.Spaces > 0 {
					spacing = (float64(width) - line.Natural) / float64(line.Spaces)
				}
			}
			if line.Natural+float64(line.Spaces)*spacing > float64(width)+0.001 || spacing < -1.3201 {
				t.Fatalf("width %d cannot fit line %+v at spacing %v", width, line, spacing)
			}
		}
		if end != expectedEnd || expectedEnd == 0 {
			t.Fatalf("width %d changed source coverage: %d, expected %d", width, end, expectedEnd)
		}
	}
}

func TestResponsivePlansOmitOverfullWidths(t *testing.T) {
	source := []byte(`<main data-hyphenate-lang="en-GB" data-plan-lines="responsive"><p>A long identifier ` + strings.Repeat("X", 48) + ` requires native wrapping in narrow columns.</p></main>`)
	output, _, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	var plans storedPlans
	paragraph := elements(parseDocument(t, output), "p")[0]
	if err := json.Unmarshal([]byte(attribute(paragraph, "data-line-plans")), &plans); err != nil {
		t.Fatal(err)
	}
	if _, exists := plans.Widths["240"]; exists {
		t.Fatal("overfull narrow paragraph received a fixed plan")
	}
	if _, exists := plans.Widths["576"]; !exists {
		t.Fatal("fitting desktop paragraph lost its plan")
	}
	if normalizeSource(paragraph) != normalizeSource(elements(parseDocument(t, source), "p")[0]) {
		t.Fatal("native fallback wording changed")
	}
}

func TestArticlePlanUsesWordSplitsToKeepSpacesNarrow(t *testing.T) {
	source := []byte(`<main data-hyphenate-lang="en-GB" data-plan-lines="responsive"><p>The use of Futures makes it easy to test asynchronous computations. Services and filters both can be created separately, each containing their specialized logic. This modularity makes it easy to test and reason about them separately. Services and filters are easily composed, just like functions do, which makes it convenient to test chains. Services and filters are meant to separate behaviour from domain logic.</p></main>`)
	output, _, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	paragraph := elements(parseDocument(t, output), "p")[0]
	var plans storedPlans
	if err := json.Unmarshal([]byte(attribute(paragraph, "data-line-plans")), &plans); err != nil {
		t.Fatal(err)
	}
	pathID, exists := plans.Widths["354"]
	if !exists {
		t.Fatal("phone-width paragraph must have a plan")
	}
	path := plans.Paths[pathID]
	hyphens := 0
	for index, id := range path {
		line := plans.Lines[id]
		if line.Hyphen {
			hyphens++
		}
		if index == len(path)-1 || line.Spaces == 0 {
			continue
		}
		spacing := (354 - line.Natural) / float64(line.Spaces)
		if spacing > 3 {
			t.Errorf("article line stretches spaces by %.3fpx despite available word splits: %+v", spacing, line)
		}
	}
	if hyphens == 0 {
		t.Error("article plan omitted every hyphenated candidate")
	}
	if normalizeSource(paragraph) != normalizeSource(parseDocument(t, source)) {
		t.Fatal("hyphenation changed the source wording")
	}
}

func TestUnsupportedProseKeepsNativeLayout(t *testing.T) {
	source := []byte(`<main data-hyphenate-lang="en-GB" data-plan-lines="576"><p>Use <ruby>unsupported markup</ruby> with asynchronous computation.</p><p>Use <span style="font-family: unknown">another font</span> in paragraphs.</p><p>Unsupported glyph 🙂.</p></main>`)
	output, _, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	paragraphs := elements(parseDocument(t, output), "p")
	for _, paragraph := range paragraphs {
		if hasAttribute(paragraph, "data-plan-width") {
			t.Fatal("unsupported prose received fixed lines")
		}
	}
	if !strings.Contains(string(output), "<ruby>un") || !strings.Contains(string(output), "🙂") {
		t.Fatal("fallback content changed")
	}
}

func TestCodeBoxesIncludeFontAndPaddingWithoutAdjustableSpaces(t *testing.T) {
	source := []byte(`<main data-hyphenate-lang="en-GB" data-plan-lines="responsive"><p><code>x y-z</code></p><ul><li><code>x y-z</code></li></ul></main>`)
	output, _, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	root := parseDocument(t, output)
	for _, tag := range []string{"p", "li"} {
		block := elements(root, tag)[0]
		var plans storedPlans
		if err := json.Unmarshal([]byte(attribute(block, "data-line-plans")), &plans); err != nil {
			t.Fatal(err)
		}
		for width, pathID := range plans.Widths {
			path := plans.Paths[pathID]
			if len(path) != 1 {
				t.Fatalf("%s at %s split a code box", tag, width)
			}
			line := plans.Lines[path[0]]
			if line.Start != 0 || line.End != 5 || line.Spaces != 0 || line.Hyphen || line.Natural < 51.99 || line.Natural > 52.01 {
				t.Fatalf("%s at %s: five 8.4px mono advances plus two 5px pads, no adjustable spaces: %+v", tag, width, line)
			}
		}
	}
}

func TestListAndParagraphCodePreserveSourceAndLinkIdentity(t *testing.T) {
	text := strings.Repeat("asynchronous services and application boundaries ", 5)
	source := []byte(`<main data-hyphenate-lang="en-GB" data-plan-lines="responsive"><p>Use <a href="/code"><code>x y-z</code></a> with ` + text + `.</p><ul><li>Use <code>x y-z</code> with ` + text + `.</li><li><p>A separate paragraph uses <code>x y-z</code>.</p><ul><li>A nested item uses <code>x y-z</code>.</li></ul></li></ul></main>`)
	output, _, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	root := parseDocument(t, output)
	if normalizeSource(root) != normalizeSource(parseDocument(t, source)) {
		t.Fatal("list or paragraph source wording changed")
	}
	links := elements(root, "a")
	if len(links) != 1 || attribute(links[0], "href") != "/code" || normalizeSource(links[0]) != "x y-z" {
		t.Fatal("linked code identity changed")
	}
	for _, code := range elements(root, "code") {
		if normalizeSource(code) != "x y-z" || len(elements(code, "br")) != 0 {
			t.Fatal("code box was split or changed")
		}
	}
	for _, paragraph := range elements(root, "p") {
		if !hasAttribute(paragraph, "data-plan-width") {
			t.Fatal("paragraph containing code received no plan")
		}
	}
	items := elements(root, "li")
	if !hasAttribute(items[0], "data-plan-width") || !hasAttribute(items[2], "data-plan-width") || hasAttribute(items[1], "data-plan-width") {
		t.Fatal("simple items or nested blocks were planned at the wrong scope")
	}
	second, changed, err := transformHTML(output)
	if err != nil || changed || !bytes.Equal(second, output) {
		t.Fatalf("mixed-font reprocessing was not idempotent: %v", err)
	}
}

func TestPlanGeometryFailsExplicitly(t *testing.T) {
	_, _, err := transformHTML([]byte(`<main data-hyphenate-lang="en-GB" data-plan-lines="640"><p>Some prose.</p></main>`))
	if err == nil || !strings.Contains(err.Error(), "unsupported planned column") {
		t.Fatalf("error = %v", err)
	}
}

func TestDeepMixedListsPlanLeafItemsAndChildParagraphs(t *testing.T) {
	source := []byte(`<main data-hyphenate-lang="en-GB" data-plan-lines="responsive"><ol><li><p>Asynchronous services operate across application boundaries.</p><ul><li>Parent text remains continuous.<ol><li><p>Nested services operate across application boundaries.</p><ul><li>A deeply nested service uses <code>value</code> across boundaries.</li></ul></li></ol></li></ul></li></ol></main>`)
	output, changed, err := transformHTML(source)
	if err != nil || !changed {
		t.Fatalf("deep list not transformed: %v", err)
	}
	root := parseDocument(t, output)
	if normalizeSource(root) != normalizeSource(parseDocument(t, source)) {
		t.Fatal("deep list source wording changed")
	}
	items := elements(root, "li")
	if len(items) != 4 || len(elements(root, "ol")) != 2 || len(elements(root, "ul")) != 2 {
		t.Fatal("mixed list structure changed")
	}
	for i, item := range items {
		if hasAttribute(item, "data-plan-width") != (i == 3) {
			t.Fatalf("list level %d planned at the wrong scope", i+1)
		}
	}
	blocks := append(elements(root, "p"), items[3])
	for _, block := range blocks {
		var plans storedPlans
		if err := json.Unmarshal([]byte(attribute(block, "data-line-plans")), &plans); err != nil {
			t.Fatalf("nested block has no readable plans: %v", err)
		}
		for _, width := range []string{"448", "576"} {
			if _, exists := plans.Widths[width]; !exists {
				t.Fatalf("nested block missing %s-pixel plan", width)
			}
		}
	}
}

func parseDocument(t *testing.T, source []byte) *html.Node {
	t.Helper()
	root, err := html.Parse(bytes.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func elements(root *html.Node, name string) []*html.Node {
	var result []*html.Node
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == name {
			result = append(result, node)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(root)
	return result
}

func normalizeSource(root *html.Node) string {
	var text strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if attribute(node, "class") == "planned-hyphen" {
			return
		}
		if node.Type == html.TextNode {
			text.WriteString(strings.ReplaceAll(node.Data, "\u00ad", ""))
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(root)
	return strings.Join(strings.Fields(text.String()), " ")
}
