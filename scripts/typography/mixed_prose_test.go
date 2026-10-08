// Verify mixed article elements retain source identity while receiving their own font and box contracts.
package main

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestMixedProsePlansPreserveFormattingAndNativeCode(t *testing.T) {
	source := []byte(`<main class="typography-prose" data-hyphenate-lang="en-GB" data-plan-lines="responsive"><p>Some <strong>important <em>asynchronous</em> services</strong> use <small>smaller text</small> and a footnote<sup id="fnref:1"><a href="#fn:1">1</a></sup>.</p><blockquote><p>Quoted services use <code>value</code> across application boundaries.</p></blockquote><figure><img src="/diagram.png" width="40" height="20"><figcaption><strong>Figure 1.</strong> A caption about asynchronous services.</figcaption></figure><div class="footnotes"><ol><li id="fn:1"><p>A smaller footnote body with a <a href="#fnref:1">back</a> link.</p></li></ol></div><p>A styled <span style="font-size:70%;padding-left:3px;padding-right:4px;text-transform:uppercase">box</span> and an inline <img src="/icon.png" width="12" height="12"> image.</p><p>A first line.<br>A second line with <em>emphasis</em>.</p><ul><li>Parent item prose.<ul><li>Nested item prose.</li></ul></li></ul><table><tbody><tr><th>Important heading text.</th><td>Asynchronous services in a cell.</td></tr></tbody></table><dl><dt>Term</dt><dd>An explanation of asynchronous services.</dd></dl><pre id="code"><code>  asynchronous\n source &lt;tag&gt;</code></pre></main>`)
	output, changed, err := transformHTML(source)
	if err != nil || !changed {
		t.Fatalf("mixed prose not transformed: %v", err)
	}
	root := parseDocument(t, output)
	if normalizeSource(root) != normalizeSource(parseDocument(t, source)) {
		t.Fatal("mixed prose source wording changed")
	}
	for _, tag := range []string{"strong", "em", "small", "sup", "a", "code", "img"} {
		if len(elements(root, tag)) != len(elements(parseDocument(t, source), tag)) {
			t.Fatalf("%s identity duplicated", tag)
		}
	}
	for _, tag := range []string{"p", "figcaption", "th", "td", "dt", "dd"} {
		for _, node := range elements(root, tag) {
			if tag == "p" && len(elements(node, "br")) > 0 && !hasAttribute(node, "data-plan-width") {
				continue
			}
			if !hasAttribute(node, "data-plan-width") {
				t.Fatalf("%s received no plan", tag)
			}
		}
	}
	segments := 0
	for _, span := range elements(root, "span") {
		if hasClass(span, "planned-prose-block") {
			segments++
			if !hasAttribute(span, "data-plan-width") {
				t.Fatal("compound prose or explicit break segment has no plan")
			}
		}
	}
	if segments != 3 {
		t.Fatalf("compound prose produced %d segments, expected 3", segments)
	}
	quote := elements(elements(root, "blockquote")[0], "p")[0]
	if attribute(quote, "data-plan-style") != "italic" {
		t.Fatal("quote lost italic contract")
	}
	caption := elements(root, "figcaption")[0]
	if attribute(caption, "data-plan-size") != "14" {
		t.Fatal("caption size not modeled")
	}
	for _, sup := range elements(root, "sup") {
		if attribute(sup, "id") != "fnref:1" {
			t.Fatal("footnote reference ID changed")
		}
	}
	var format textFormat
	for _, span := range elements(root, "span") {
		if attribute(span, "style") == "font-size:70%;padding-left:3px;padding-right:4px;text-transform:uppercase" {
			if err := json.Unmarshal([]byte(attribute(span, "data-plan-inline")), &format); err != nil {
				t.Fatal(err)
			}
			if math.Abs(format.Size-11.2) > 0.0001 || format.PaddingLeft != 3 || format.PaddingRight != 4 || format.Transform != "uppercase" {
				t.Fatalf("styled box geometry wrong: %+v", format)
			}
		}
	}
	if !strings.Contains(string(output), `<pre id="code"><code>  asynchronous\n source &lt;tag&gt;</code></pre>`) {
		t.Fatal("code block changed")
	}
	second, changed, err := transformHTML(output)
	if err != nil || changed || !bytes.Equal(second, output) {
		t.Fatalf("mixed transform not idempotent: %v", err)
	}
}

func TestNestedExplicitBreakKeepsSourceAndRaggedLines(t *testing.T) {
	source := []byte(`<main class="typography-prose" data-plan-lines="responsive" data-hyphenate-lang="en-GB"><p>Some <em>emphasized text<br>with a second line</em> remains intact.</p></main>`)
	output, _, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	root := parseDocument(t, output)
	paragraph := elements(root, "p")[0]
	if !hasAttribute(paragraph, "data-plan-width") {
		t.Fatal("nested break prevented planning")
	}
	if normalizeSource(root) != normalizeSource(parseDocument(t, source)) {
		t.Fatal("nested break wording changed")
	}
	originalBreaks := 0
	for _, br := range elements(root, "br") {
		if !hasClass(br, "planned-break") {
			originalBreaks++
		}
	}
	if originalBreaks != 1 {
		t.Fatal("original explicit break duplicated or removed")
	}
	var plans storedPlans
	if err := json.Unmarshal([]byte(attribute(paragraph, "data-line-plans")), &plans); err != nil {
		t.Fatal(err)
	}
	for _, path := range plans.Paths {
		mandatory := 0
		for _, id := range path {
			if plans.Lines[id].Ragged {
				mandatory++
			}
		}
		if mandatory != 1 {
			t.Fatal("explicit break missing from a responsive plan")
		}
	}
}

func TestNarrowTableCellUsesHyphenationAndOwnFontSize(t *testing.T) {
	source := []byte(`<main class="typography-prose" data-plan-lines="responsive" data-hyphenate-lang="en-GB"><table><tbody><tr><td>Table prose contains <em>italic emphasis</em> and <code>fixed code</code> across asynchronous application boundaries.</td><td><p>Asynchronous services in a cell.</p></td></tr></tbody></table></main>`)
	output, _, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	cell := elements(parseDocument(t, output), "td")[0]
	if attribute(cell, "data-plan-size") != "14" || !strings.Contains(documentText(t, output), "\u00ad") {
		t.Fatal("table cell lacks its smaller font or discretionary hyphens")
	}
	var plans storedPlans
	if err := json.Unmarshal([]byte(attribute(cell, "data-line-plans")), &plans); err != nil {
		t.Fatal(err)
	}
	if _, ok := plans.Widths["171"]; !ok {
		t.Fatal("mixed table prose has no plan for a phone cell")
	}
	paragraph := elements(parseDocument(t, output), "p")[0]
	if err := json.Unmarshal([]byte(attribute(paragraph, "data-line-plans")), &plans); err != nil {
		t.Fatal(err)
	}
	if attribute(paragraph, "data-plan-size") != "14" {
		t.Fatal("paragraph inside a cell lost the table font size")
	}
	if _, ok := plans.Widths["171"]; !ok {
		t.Fatal("paragraph inside a cell did not inherit the table width range")
	}
	if normalizeSource(parseDocument(t, output)) != normalizeSource(parseDocument(t, source)) {
		t.Fatal("table source text changed")
	}
}

func TestInlineLengthsUseFinalFontSizeAndInheritedPercentage(t *testing.T) {
	source := []byte(`<main class="typography-prose" data-plan-lines="576" data-hyphenate-lang="en-GB"><p>Text <sup style="padding-left:0.5em;font-size:70%;font-size:75%">reference</sup> text.</p></main>`)
	output, _, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	var format textFormat
	if err := json.Unmarshal([]byte(attribute(elements(parseDocument(t, output), "sup")[0], "data-plan-inline")), &format); err != nil {
		t.Fatal(err)
	}
	if format.Size != 12 || format.PaddingLeft != 6 {
		t.Fatalf("relative CSS resolved in declaration order instead of final font size: %+v", format)
	}
}
