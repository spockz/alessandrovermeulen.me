// Verify the blog default without changing explicit typography choices, controls, or native article elements.
package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestBlogDefaultScopesOnlyArticleProse(t *testing.T) {
	source := []byte(`<html><head><meta name="article-typography" content="planned"></head><body><p>Asynchronous navigation.</p><div class="article-content max-w-prose"><p>Asynchronous computation across services and application boundaries.</p><div class="not-prose"><p>Asynchronous controls.</p></div><pre><code>asynchronous source code</code></pre><blockquote><p>Asynchronous quotation.</p></blockquote><div class="typography-prose typography-pretty"><p>Asynchronous native override.</p></div><div class="typography-prose" data-hyphenate-lang="en-GB" data-plan-lines="responsive"><p>Asynchronous comparison prose.</p></div></div><footer>Asynchronous footer.</footer></body></html>`)
	output, changed, err := transformHTML(source)
	if err != nil || !changed {
		t.Fatalf("blog default not applied: %v", err)
	}
	root := parseDocument(t, output)
	if normalizeSource(root) != normalizeSource(parseDocument(t, source)) {
		t.Fatal("blog default changed source wording")
	}
	paragraphs := elements(root, "p")
	for i, paragraph := range paragraphs {
		planned := hasAttribute(paragraph, "data-plan-width")
		if planned != (i == 1 || i == 3 || i == 5) {
			t.Fatalf("paragraph %d planned=%v at the wrong scope", i, planned)
		}
	}
	for _, native := range []string{"<p>Asynchronous navigation.</p>", "<p>Asynchronous controls.</p>", "<p>Asynchronous native override.</p>", "<code>asynchronous source code</code>", "<footer>Asynchronous footer.</footer>"} {
		if !strings.Contains(string(output), native) {
			t.Fatalf("native element changed: %s", native)
		}
	}
	second, changed, err := transformHTML(output)
	if err != nil || changed || !bytes.Equal(second, output) {
		t.Fatalf("blog default is not idempotent: %v", err)
	}
}

func TestBlogDefaultRequiresItsArticleScope(t *testing.T) {
	for _, body := range []string{`<main>No article scope.</main>`, `<div class="article-content"></div><div class="article-content"></div>`} {
		_, _, err := transformHTML([]byte(`<head><meta name="article-typography" content="planned"></head><body>` + body + `</body>`))
		if err == nil || !strings.Contains(err.Error(), "requires one article-content scope") {
			t.Fatalf("missing scope error: %v", err)
		}
	}
}

func TestNonBlogPageKeepsExistingLayout(t *testing.T) {
	source := []byte(`<html><body><div class="article-content"><p>Asynchronous prose on an ordinary page.</p></div></body></html>`)
	output, changed, err := transformHTML(source)
	if err != nil || changed || !bytes.Equal(output, source) {
		t.Fatalf("ordinary page received blog defaults: %v", err)
	}
}

func TestArticleSummariesReceivePlansWithoutChangingNavigation(t *testing.T) {
	source := []byte(`<main><h1>Writing</h1><article><a href="/posts/example/">Example article</a><div class="article-link__summary prose line-clamp-3">Asynchronous computation across services and application boundaries.</div></article><article><a href="/posts/other/">Other article</a><div class="article-link__summary prose"><p>Another paragraph about asynchronous services.</p></div></article></main>`)
	output, changed, err := transformHTML(source)
	if err != nil || !changed {
		t.Fatalf("summary planning not applied: %v", err)
	}
	root := parseDocument(t, output)
	paragraphs := elements(root, "p")
	if len(paragraphs) != 2 {
		t.Fatalf("summary paragraph structure changed: %d", len(paragraphs))
	}
	for _, paragraph := range paragraphs {
		if !hasAttribute(paragraph, "data-line-plans") || !hasClass(paragraph.Parent, "typography-summary") || attribute(paragraph.Parent, "data-plan-generated") != "true" {
			t.Fatal("summary is missing its responsive plan and pinned-font scope")
		}
	}
	if normalizeSource(root) != normalizeSource(parseDocument(t, source)) {
		t.Fatal("summary planning changed the source words")
	}
	if !strings.Contains(string(output), `<h1>Writing</h1>`) || !strings.Contains(string(output), `<a href="/posts/example/">Example article</a>`) {
		t.Fatal("summary planning changed a heading or article link")
	}
	second, changed, err := transformHTML(output)
	if err != nil || changed || !bytes.Equal(second, output) {
		t.Fatalf("summary processing is not idempotent: %v", err)
	}
}
