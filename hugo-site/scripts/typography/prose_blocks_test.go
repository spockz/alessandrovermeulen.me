// Verify compound prose partitioning preserves source DOM nodes and remains safe to repeat.
package main

import (
	"bytes"
	"testing"

	"golang.org/x/net/html"
)

func TestProseBlockPartitionPreservesNativeStructure(t *testing.T) {
	source := []byte(`<main><ul><li id="compound">Before <strong data-mark="kept">bold</strong><br> between breaks<br> <ul id="nested"><li>Nested wording.</li></ul> <p id="tail">After block.</p></li><li id="simple">Simple item.</li></ul><p id="plain">Plain paragraph.</p><blockquote>Standalone quote<br>still native.</blockquote></main>`)
	root := parseDocument(t, source)
	items := elements(root, "li")
	compound, simple := items[0], items[2]
	paragraph := elements(root, "p")[1]
	quote := elements(root, "blockquote")[0]
	nestedList := elements(root, "ul")[1]
	tailParagraph := elements(root, "p")[0]
	strong := elements(compound, "strong")[0]
	breaks := elements(compound, "br")
	if len(breaks) != 2 {
		t.Fatalf("expected two original breaks, got %d", len(breaks))
	}
	originalText := normalizeSource(root)

	if !prepareProseBlocks(root) {
		t.Fatal("expected compound list item to be partitioned")
	}
	if got := normalizeSource(root); got != originalText {
		t.Fatalf("partition changed source wording: got %q, want %q", got, originalText)
	}
	if hasAttribute(simple, "data-plan-width") || len(elements(simple, "span")) != 0 || len(elements(paragraph, "span")) != 0 {
		t.Fatal("simple list item or paragraph was changed")
	}
	if len(elements(quote, "span")) != 0 {
		t.Fatal("direct text in a blockquote without block children was wrapped")
	}
	if elements(root, "ul")[1] != nestedList || elements(root, "p")[0] != tailParagraph || elements(compound, "strong")[0] != strong {
		t.Fatal("partition replaced an original block or inline node")
	}
	if attribute(nestedList, "id") != "nested" || attribute(tailParagraph, "id") != "tail" || attribute(strong, "data-mark") != "kept" {
		t.Fatalf("partition changed original node attributes: nested=%q tail=%q mark=%q", attribute(nestedList, "id"), attribute(tailParagraph, "id"), attribute(strong, "data-mark"))
	}
	if got := elements(compound, "br"); len(got) != 2 || got[0] != breaks[0] || got[1] != breaks[1] {
		t.Fatal("partition replaced or moved an original BR node")
	}
	wrappers := 0
	for _, span := range elements(compound, "span") {
		if hasClass(span, "planned-prose-block") {
			wrappers++
		}
	}
	if wrappers < 2 {
		t.Fatalf("expected prose groups on both sides of block children, got %d", wrappers)
	}
	whitespaceNodes := 0
	for child := compound.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode && child.Data == " " {
			whitespaceNodes++
		}
	}
	if whitespaceNodes == 0 {
		t.Fatal("whitespace-only groups were not preserved as direct text")
	}
	for _, span := range elements(compound, "span") {
		if hasClass(span, "planned-prose-block") && normalizeSource(span) == "" {
			t.Fatal("whitespace-only group received a wrapper")
		}
	}

	var before bytes.Buffer
	if err := html.Render(&before, root); err != nil {
		t.Fatal(err)
	}
	if prepareProseBlocks(root) {
		t.Fatal("second call reported a mutation")
	}
	var after bytes.Buffer
	if err := html.Render(&after, root); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before.Bytes(), after.Bytes()) {
		t.Fatal("second call mutated the DOM")
	}
}
