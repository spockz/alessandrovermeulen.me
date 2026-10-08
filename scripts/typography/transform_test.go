// This file verifies observable scope, safety, idempotence, and Unicode behavior.
package main

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/speedata/hyphenation"
	"golang.org/x/net/html"
)

func TestKnownEnglishWordGetsSoftHyphen(t *testing.T) {
	h, err := newHyphenator("en-GB")
	if err != nil {
		t.Fatal(err)
	}
	for word, expected := range map[string]string{
		"hyphenation": "hy\u00adphen\u00ada\u00adtion",
		"university":  "uni\u00adver\u00adsity",
		"things":      "things",
		"think":       "think",
		"thinking":    "think\u00ading",
		"other":       "other",
	} {
		got, _ := hyphenateWord(word, h)
		if got != expected {
			t.Errorf("hyphenation of %q = %q, want %q", word, got, expected)
		}
	}
}

func TestTransformUsesOnlyMarkedEligiblePlainText(t *testing.T) {
	source := []byte(`<div><p>hyphenation</p></div><main data-hyphenate-lang="en-GB"><div>hyphenation</div><p><strong>hyphenation</strong> <a href="https://example.com/hyphenation">hyphenation</a> https://example.com/hyphenation user@example.com SoundCloud LoggingFilter already-hyphenated.</p><li>hyphenation<h2>hyphenation</h2></li><pre>hyphenation</pre><code>hyphenation</code><script>hyphenation</script><style>hyphenation</style><textarea>hyphenation</textarea><kbd>hyphenation</kbd><samp>hyphenation</samp><svg><text>hyphenation</text></svg><math><mi>hyphenation</mi></math></main>`)
	gotBytes, changed, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	got := string(gotBytes)
	if !changed {
		t.Fatal("expected marked paragraph text to change")
	}
	originalText := strings.ReplaceAll(documentText(t, source), "\u00ad", "")
	transformedText := strings.ReplaceAll(documentText(t, gotBytes), "\u00ad", "")
	if transformedText != originalText {
		t.Fatalf("transformation changed visible wording:\noriginal: %q\ntransformed: %q", originalText, transformedText)
	}
	if strings.Count(got, "hy\u00adphen") < 3 {
		t.Fatalf("expected prose, bold text, and link label to hyphenate: %q", got)
	}
	for _, preserved := range []string{
		"<div><p>hyphenation</p></div>",
		"<div>hyphenation</div>",
		"href=\"https://example.com/hyphenation\"",
		"https://example.com/hyphenation",
		"user@example.com",
		"SoundCloud",
		"LoggingFilter",
		"already-hyphenated",
		"<h2>hyphenation</h2>",
		"<pre>hyphenation</pre>",
		"<code>hyphenation</code>",
		"<script>hyphenation</script>",
		"<style>hyphenation</style>",
		"<textarea>hyphenation</textarea>",
		"<kbd>hyphenation</kbd>",
		"<samp>hyphenation</samp>",
		"<svg><text>hyphenation</text></svg>",
		"<math><mi>hyphenation</mi></math>",
	} {
		if !strings.Contains(got, preserved) {
			t.Errorf("expected preserved content %q in transformed document %q", preserved, got)
		}
	}
}

func documentText(t *testing.T, source []byte) string {
	t.Helper()
	doc, err := html.Parse(bytes.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	var text strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(doc)
	return text.String()
}

func TestUnmarkedHTMLRemainsByteIdentical(t *testing.T) {
	source := []byte("<p>hyphenation &amp; café</p>\n")
	got, changed, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	if changed || string(got) != string(source) {
		t.Fatalf("unmarked HTML changed: %q (changed %v)", got, changed)
	}
}

func TestSentencePunctuationAllowsBreaksButDomainLabelsDoNot(t *testing.T) {
	source := []byte(`<p data-hyphenate-lang="en-GB">hyphenation. Another sentence; hyphenation.a</p>`)
	output, changed, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	words := strings.Fields(documentText(t, output))
	if !changed || !strings.ContainsRune(words[0], '\u00ad') {
		t.Fatalf("sentence-final word lacks soft hyphens: %q", words)
	}
	if words[len(words)-1] != "hyphenation.a" {
		t.Fatalf("domain label changed: %q", words[len(words)-1])
	}
}

func TestExistingSoftHyphenIsIdempotent(t *testing.T) {
	source := []byte(`<p data-hyphenate-lang="en-GB">hy&#173;phenation</p>`)
	first, changed, err := transformHTML(source)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatalf("existing soft hyphen should remain unchanged: %q", first)
	}
	second, changed, err := transformHTML(first)
	if err != nil {
		t.Fatal(err)
	}
	if changed || string(second) != string(first) {
		t.Fatalf("second transform changed output: %q, changed %v", second, changed)
	}
}

func TestUnicodeBreakPositionsUseRunesAndEnforceMinimums(t *testing.T) {
	h, err := hyphenation.New(strings.NewReader("coö1p a1bc abcd1e"))
	if err != nil {
		t.Fatal(err)
	}
	got, changed := hyphenateWord("coöperation", &hyphenator{patterns: h})
	if !changed || got != "coö\u00adperation" {
		t.Fatalf("Unicode hyphenation = %q, changed %v", got, changed)
	}
	for _, word := range []string{"abcd", "abcde"} {
		if got, changed := hyphenateWord(word, &hyphenator{patterns: h}); changed || got != word {
			t.Errorf("minimum edge lengths should reject %q: got %q, changed %v", word, got, changed)
		}
	}
	markBefore := mustHyphenator(t, "ab1\u0301")
	if got, changed := hyphenateWord("ab\u0301cdef", markBefore); changed || got != "ab\u0301cdef" {
		t.Fatalf("break before a combining mark should be rejected: %q, changed %v", got, changed)
	}
}

func mustHyphenator(t *testing.T, patterns string) *hyphenator {
	t.Helper()
	loadedPatterns, err := hyphenation.New(strings.NewReader(patterns))
	if err != nil {
		t.Fatal(err)
	}
	return &hyphenator{patterns: loadedPatterns}
}

func TestUnsupportedLanguageFails(t *testing.T) {
	if _, _, err := transformHTML([]byte(`<main DATA-HYPHENATE-LANG="fr"><p>bonjour</p></main>`)); err == nil || !strings.Contains(err.Error(), "unsupported hyphenation language") {
		t.Fatalf("unsupported language error = %v", err)
	}
}

func TestTransformResponseReplacesBodyAndStaleHeaders(t *testing.T) {
	body := `<main data-hyphenate-lang="en-GB"><p>hyphenation</p></main>`
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type":   []string{"text/html; charset=utf-8"},
			"Content-Length": []string{"1"},
			"Etag":           []string{"stale"},
			"Content-Md5":    []string{"stale"},
		},
		Body:    io.NopCloser(strings.NewReader(body)),
		Request: &http.Request{Method: http.MethodGet},
	}
	if err := transformResponse(response); err != nil {
		t.Fatal(err)
	}
	transformed, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(transformed, []byte("hy\u00adphen")) {
		t.Fatalf("HTML response body was not transformed: %q", transformed)
	}
	if response.Header.Get("ETag") != "" || response.Header.Get("Content-MD5") != "" {
		t.Fatalf("stale validators remain: %#v", response.Header)
	}
	if response.ContentLength != int64(len(transformed)) || response.Header.Get("Content-Length") != strconv.Itoa(len(transformed)) {
		t.Fatalf("content length = %d / %s, body length %d", response.ContentLength, response.Header.Get("Content-Length"), len(transformed))
	}
}

func TestTransformResponseRejectsUnexpectedCompression(t *testing.T) {
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/html"}, "Content-Encoding": []string{"gzip"}},
		Body:       io.NopCloser(strings.NewReader("compressed")),
		Request:    &http.Request{Method: http.MethodGet},
	}
	if err := transformResponse(response); err == nil || !strings.Contains(err.Error(), "compressed HTML") {
		t.Fatalf("compressed response error = %v", err)
	}
}
