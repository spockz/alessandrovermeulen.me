// Partition compound prose into styleable inline blocks while preserving source DOM nodes for fallback and clipboard use.
package main

import (
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func prepareProseBlocks(node *html.Node) bool {
	changed := false
	var visit func(*html.Node)
	visit = func(current *html.Node) {
		if hasClass(current, "not-prose") || (hasClass(current, "typography-prose") && !hasAttribute(current, "data-plan-lines")) {
			return
		}
		if current.Type == html.ElementNode && (current.Data == "pre" || current.Data == "code") {
			return
		}
		for child := current.FirstChild; child != nil; {
			next := child.NextSibling
			visit(child)
			child = next
		}
		if current.Type != html.ElementNode || !isProseBlockCandidate(current) || hasClass(current, "planned-prose-block") {
			return
		}
		if current.Data == "blockquote" && !hasDirectBlockChild(current) {
			return
		}
		if !hasDirectBreakOrBlock(current) || hasDirectProseWrapper(current) {
			return
		}
		changed = wrapInlineGroups(current) || changed
	}
	visit(node)
	return changed
}

func isProseBlockCandidate(node *html.Node) bool {
	switch node.Data {
	case "p", "li", "figcaption", "td", "th", "dt", "dd", "blockquote":
		return true
	default:
		return false
	}
}

func hasDirectBreakOrBlock(node *html.Node) bool {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && (child.Data == "br" || isBlockChild(child.Data)) {
			return true
		}
	}
	return false
}

func hasDirectBlockChild(node *html.Node) bool {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && isBlockChild(child.Data) {
			return true
		}
	}
	return false
}

func hasDirectProseWrapper(node *html.Node) bool {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && hasClass(child, "planned-prose-block") {
			return true
		}
	}
	return false
}

func isBlockChild(name string) bool {
	switch name {
	case "address", "article", "aside", "blockquote", "div", "dl", "fieldset", "figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6", "header", "hr", "main", "nav", "ol", "p", "pre", "section", "table", "ul":
		return true
	default:
		return false
	}
}

func wrapInlineGroups(parent *html.Node) bool {
	changed := false
	for child := parent.FirstChild; child != nil; {
		if isProseBoundary(child) {
			child = child.NextSibling
			continue
		}
		start := child
		for child != nil && !isProseBoundary(child) {
			child = child.NextSibling
		}
		if !groupHasWords(start, child) {
			continue
		}
		wrapper := &html.Node{Type: html.ElementNode, Data: "span", DataAtom: atom.Span, Attr: []html.Attribute{{Key: "class", Val: "planned-prose-block"}}}
		parent.InsertBefore(wrapper, start)
		for current := start; current != child; {
			next := current.NextSibling
			parent.RemoveChild(current)
			wrapper.AppendChild(current)
			current = next
		}
		changed = true
	}
	return changed
}

func isProseBoundary(node *html.Node) bool {
	return node.Type == html.ElementNode && (node.Data == "br" || isBlockChild(node.Data) || hasClass(node, "planned-prose-block"))
}

func groupHasWords(start, end *html.Node) bool {
	var visit func(*html.Node) bool
	visit = func(node *html.Node) bool {
		if node.Type == html.TextNode {
			for _, r := range node.Data {
				if !unicode.IsSpace(r) {
					return true
				}
			}
			return false
		}
		if node.Type == html.ElementNode && (node.Data == "pre" || node.Data == "code") {
			return false
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if visit(child) {
				return true
			}
		}
		return false
	}
	for current := start; current != end; current = current.NextSibling {
		if visit(current) {
			return true
		}
	}
	return false
}
