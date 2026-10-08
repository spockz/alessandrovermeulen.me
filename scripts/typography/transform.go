// Apply the blog default and explicit typography scopes through the same transform in preview and release.
package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

var urlPattern = regexp.MustCompile(`(?i)(?:https?://|www\.)[^\s<>]+`)

func transformHTML(source []byte) ([]byte, bool, error) {
	if !containsFoldASCII(source, []byte("data-hyphenate-lang")) && !containsFoldASCII(source, []byte("article-typography")) && !containsFoldASCII(source, []byte("article-link__summary")) {
		return source, false, nil
	}
	doc, err := html.Parse(bytes.NewReader(source))
	if err != nil {
		return nil, false, fmt.Errorf("parse HTML: %w", err)
	}
	changed, err := markBlogContent(doc)
	if err != nil {
		return nil, false, err
	}
	changed = markSummaryContent(doc) || changed
	marked, err := documentHyphenator(doc)
	if err != nil {
		return nil, false, err
	}
	if marked != nil {
		walkDocument(doc, false, false, false, marked, &changed)
	}
	planned, err := planDocument(doc)
	if err != nil {
		return nil, false, err
	}
	changed = changed || planned
	if !changed {
		return source, false, nil
	}
	var output bytes.Buffer
	if err := html.Render(&output, doc); err != nil {
		return nil, false, fmt.Errorf("render transformed HTML: %w", err)
	}
	return output.Bytes(), true, nil
}

func markSummaryContent(root *html.Node) bool {
	changed := false
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if hasClass(node, "article-link__summary") && !hasAttribute(node, "data-plan-lines") {
			for i := range node.Attr {
				if node.Attr[i].Key == "class" {
					node.Attr[i].Val += " typography-prose typography-summary"
				}
			}
			node.Attr = append(node.Attr,
				html.Attribute{Key: "data-hyphenate-lang", Val: "en-GB"},
				html.Attribute{Key: "data-plan-lines", Val: "responsive"})
			if !hasDirectBlockChild(node) {
				paragraph := &html.Node{Type: html.ElementNode, Data: "p"}
				for node.FirstChild != nil {
					child := node.FirstChild
					node.RemoveChild(child)
					paragraph.AppendChild(child)
				}
				node.AppendChild(paragraph)
			}
			changed = true
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(root)
	return changed
}

func markBlogContent(root *html.Node) (bool, error) {
	marked := false
	var scopes []*html.Node
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "meta" && attribute(node, "name") == "article-typography" && attribute(node, "content") == "planned" {
				marked = true
			}
			if hasClass(node, "article-content") {
				scopes = append(scopes, node)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(root)
	if !marked {
		return false, nil
	}
	if len(scopes) != 1 {
		return false, fmt.Errorf("blog typography marker requires one article-content scope, found %d", len(scopes))
	}
	scope := scopes[0]
	if hasAttribute(scope, "data-plan-lines") {
		return false, nil
	}
	for i := range scope.Attr {
		if scope.Attr[i].Key == "class" {
			scope.Attr[i].Val += " typography-prose"
		}
	}
	scope.Attr = append(scope.Attr,
		html.Attribute{Key: "data-hyphenate-lang", Val: "en-GB"},
		html.Attribute{Key: "data-plan-lines", Val: "responsive"})
	return true, nil
}

func containsFoldASCII(source, target []byte) bool {
	for start := 0; start+len(target) <= len(source); start++ {
		for index, expected := range target {
			actual := source[start+index]
			if actual >= 'A' && actual <= 'Z' {
				actual += 'a' - 'A'
			}
			if actual != expected {
				break
			}
			if index == len(target)-1 {
				return true
			}
		}
	}
	return false
}

func documentHyphenator(root *html.Node) (*hyphenator, error) {
	var language string
	var find func(*html.Node) error
	find = func(node *html.Node) error {
		if node.Type == html.ElementNode && hasAttribute(node, "data-hyphenate-lang") {
			language = attribute(node, "data-hyphenate-lang")
			if !strings.EqualFold(language, "en-GB") {
				return fmt.Errorf("unsupported hyphenation language %q (supported: en-GB)", language)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := find(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := find(root); err != nil {
		return nil, err
	}
	if language == "" {
		return nil, nil
	}
	return newHyphenator(language)
}

func walkDocument(node *html.Node, marked, eligible, protected bool, h *hyphenator, changed *bool) {
	if node.Type == html.ElementNode {
		if hasAttribute(node, "data-plan-generated") {
			return
		}
		if hasAttribute(node, "data-hyphenate-lang") {
			marked = true
		}
		if hasClass(node, "not-prose") || hasClass(node, "typography-pretty") {
			return
		}
		name := strings.ToLower(node.Data)
		protected = protected || isProtectedElement(name) || isHeading(name)
		eligible = eligible || isProseBlockCandidate(node) || hasClass(node, "planned-prose-block")
	}
	if node.Type == html.TextNode && marked && eligible && !protected && h != nil {
		text, didChange := hyphenateText(node.Data, h)
		if didChange {
			node.Data = text
			*changed = true
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walkDocument(child, marked, eligible, protected, h, changed)
	}
}

func isHeading(name string) bool {
	return len(name) == 2 && name[0] == 'h' && name[1] >= '1' && name[1] <= '6'
}

func hasAttribute(node *html.Node, key string) bool {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, key) {
			return true
		}
	}
	return false
}

func hasClass(node *html.Node, class string) bool {
	for _, value := range strings.Fields(attribute(node, "class")) {
		if value == class {
			return true
		}
	}
	return false
}

func attribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, key) {
			return attr.Val
		}
	}
	return ""
}

func isProtectedElement(name string) bool {
	switch name {
	case "code", "pre", "script", "style", "textarea", "kbd", "samp", "svg", "math":
		return true
	default:
		return false
	}
}

func hyphenateText(input string, h *hyphenator) (string, bool) {
	runes := []rune(input)
	if len(runes) == 0 {
		return input, false
	}
	protected := make([]bool, len(runes))
	for _, match := range urlPattern.FindAllStringIndex(input, -1) {
		start := len([]rune(input[:match[0]]))
		end := start + len([]rune(input[match[0]:match[1]]))
		for i := start; i < end && i < len(protected); i++ {
			protected[i] = true
		}
	}
	var output strings.Builder
	changed := false
	for i := 0; i < len(runes); {
		if !unicode.IsLetter(runes[i]) && !unicode.IsMark(runes[i]) {
			output.WriteRune(runes[i])
			i++
			continue
		}
		start := i
		for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsMark(runes[i])) {
			i++
		}
		word := string(runes[start:i])
		if shouldHyphenate(runes, start, i, protected) {
			withBreaks, didChange := hyphenateWord(word, h)
			output.WriteString(withBreaks)
			changed = changed || didChange
		} else {
			output.WriteString(word)
		}
	}
	return output.String(), changed
}

func shouldHyphenate(runes []rune, start, end int, protected []bool) bool {
	word := runes[start:end]
	if len(word) < 5 {
		return false
	}
	for i := start; i < end; i++ {
		if protected[i] || runes[i] == '\u00ad' {
			return false
		}
	}
	if start > 0 && isWordConnector(runes, start-1, -1) || end < len(runes) && isWordConnector(runes, end, 1) {
		return false
	}
	seenUpper := false
	for index, r := range word {
		if !unicode.IsLetter(r) && !unicode.IsMark(r) {
			return false
		}
		if unicode.IsUpper(r) {
			if index != 0 || seenUpper {
				return false
			}
			seenUpper = true
		}
	}
	return true
}

func isWordConnector(runes []rune, index, direction int) bool {
	r := runes[index]
	switch r {
	case '-', '\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u0027', '\u2019', '@', '_', '\u00ad':
		return true
	case '.':
		far := index + direction
		return far >= 0 && far < len(runes) && (unicode.IsLetter(runes[far]) || unicode.IsDigit(runes[far]))
	case '/', ':':
		far := index + direction
		return far >= 0 && far < len(runes) && !unicode.IsSpace(runes[far]) && (unicode.IsLetter(runes[far]) || unicode.IsDigit(runes[far]))
	default:
		return false
	}
}

func hyphenateWord(word string, h *hyphenator) (string, bool) {
	runes := []rune(word)
	breaks := h.breakPoints(word)
	var output strings.Builder
	last := 0
	for _, position := range breaks {
		if position < 2 || position > len(runes)-3 || position <= last || unicode.IsMark(runes[position]) {
			continue
		}
		output.WriteString(string(runes[last:position]))
		output.WriteRune('\u00ad')
		last = position
	}
	if last == 0 {
		return word, false
	}
	output.WriteString(string(runes[last:]))
	return output.String(), true
}
