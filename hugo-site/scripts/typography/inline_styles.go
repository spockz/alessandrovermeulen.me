// Model the site's prose CSS in the builder and publish the same metrics for browser validation.
package main

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type textFormat struct {
	Family           string  `json:"family"`
	Size             float64 `json:"size"`
	Weight           string  `json:"weight"`
	Style            string  `json:"style"`
	Transform        string  `json:"transform"`
	LetterSpacing    float64 `json:"letterSpacing"`
	PaddingLeft      float64 `json:"paddingLeft"`
	PaddingRight     float64 `json:"paddingRight"`
	BorderLeftWidth  float64 `json:"borderLeftWidth"`
	BorderRightWidth float64 `json:"borderRightWidth"`
	MarginLeft       float64 `json:"marginLeft"`
	MarginRight      float64 `json:"marginRight"`
}

func baseFormat() textFormat {
	return textFormat{Family: "article sans", Size: 16, Weight: "400", Style: "normal", Transform: "none"}
}

func (f textFormat) textOnly() textFormat {
	f.PaddingLeft, f.PaddingRight, f.BorderLeftWidth, f.BorderRightWidth, f.MarginLeft, f.MarginRight = 0, 0, 0, 0, 0, 0
	return f
}

func (f textFormat) edges() float64 {
	return f.PaddingLeft + f.PaddingRight + f.BorderLeftWidth + f.BorderRightWidth + f.MarginLeft + f.MarginRight
}

func inlineElement(tag string) bool {
	switch tag {
	case "a", "strong", "b", "em", "i", "small", "sup", "sub", "span", "code", "kbd", "samp", "mark", "del", "s", "u", "abbr", "cite", "q", "img", "br":
		return true
	}
	return false
}

func formatNode(node *html.Node, inherited textFormat) (textFormat, bool) {
	f := inherited.textOnly()
	switch node.Data {
	case "strong", "b", "th", "dt":
		f.Weight = "700"
	case "em", "i", "cite", "blockquote":
		f.Style = "italic"
	case "small":
		f.Size *= 0.8
	case "sup", "sub":
		f.Size *= 0.75
	case "figcaption":
		f.Size *= 0.875
		f.Style = "normal"
		f.Weight = "400"
	case "table":
		f.Size *= 0.875
	case "code", "kbd", "samp":
		f.Family = "article mono"
		f.Size *= 0.875
		f.Style = "normal"
		f.Weight = "400"
		f.PaddingLeft = 5
		f.PaddingRight = 5
	}
	if hasClass(node, "footnotes") {
		f.Size *= 0.875
	}
	for _, declaration := range strings.Split(attribute(node, "style"), ";") {
		key, value, exists := strings.Cut(declaration, ":")
		if exists && strings.ToLower(strings.TrimSpace(key)) == "font-size" {
			parsed, ok := cssLength(strings.ToLower(strings.TrimSpace(value)), inherited.Size)
			if !ok || parsed <= 0 {
				return f, false
			}
			f.Size = parsed
		}
	}
	for _, declaration := range strings.Split(attribute(node, "style"), ";") {
		key, value, exists := strings.Cut(declaration, ":")
		if !exists {
			continue
		}
		key, value = strings.ToLower(strings.TrimSpace(key)), strings.ToLower(strings.TrimSpace(value))
		switch key {
		case "font-size":
			continue
		case "font-weight":
			if value == "bold" {
				value = "700"
			}
			if value == "normal" {
				value = "400"
			}
			if value != "400" && value != "700" {
				return f, false
			}
			f.Weight = value
		case "font-style":
			if value != "normal" && value != "italic" {
				return f, false
			}
			f.Style = value
		case "font-family":
			value = strings.Trim(strings.Split(value, ",")[0], "\"'")
			if value != "article sans" && value != "article mono" {
				return f, false
			}
			f.Family = value
		case "text-transform":
			if value != "none" && value != "uppercase" && value != "lowercase" && value != "capitalize" {
				return f, false
			}
			f.Transform = value
		case "letter-spacing":
			if value == "normal" {
				f.LetterSpacing = 0
				continue
			}
			parsed, ok := cssLength(value, f.Size)
			if !ok {
				return f, false
			}
			f.LetterSpacing = parsed
		case "padding-inline", "margin-inline":
			parts := strings.Fields(value)
			if len(parts) < 1 || len(parts) > 2 {
				return f, false
			}
			left, okL := cssLength(parts[0], f.Size)
			right := left
			okR := okL
			if len(parts) == 2 {
				right, okR = cssLength(parts[1], f.Size)
			}
			if !okL || !okR {
				return f, false
			}
			if key == "padding-inline" {
				if left < 0 || right < 0 {
					return f, false
				}
				f.PaddingLeft, f.PaddingRight = left, right
			} else {
				f.MarginLeft, f.MarginRight = left, right
			}
		case "border", "border-left", "border-right":
			width := 0.0
			for _, part := range strings.Fields(value) {
				if parsed, ok := cssLength(part, f.Size); ok {
					width = parsed
					break
				}
			}
			if width < 0 {
				return f, false
			}
			if strings.Contains(value, "none") || strings.Contains(value, "hidden") {
				width = 0
			}
			if key != "border-right" {
				f.BorderLeftWidth = width
			}
			if key != "border-left" {
				f.BorderRightWidth = width
			}
		case "padding", "margin":
			parts := strings.Fields(value)
			if len(parts) < 1 || len(parts) > 4 {
				return f, false
			}
			left, right := parts[0], parts[0]
			if len(parts) >= 2 {
				left, right = parts[1], parts[1]
			}
			if len(parts) == 4 {
				left = parts[3]
			}
			l, okL := cssLength(left, f.Size)
			r, okR := cssLength(right, f.Size)
			if !okL || !okR {
				return f, false
			}
			if key == "padding" {
				if l < 0 || r < 0 {
					return f, false
				}
				f.PaddingLeft, f.PaddingRight = l, r
			} else {
				f.MarginLeft, f.MarginRight = l, r
			}
		case "padding-left", "padding-right", "margin-left", "margin-right", "border-left-width", "border-right-width":
			parsed, ok := cssLength(value, f.Size)
			if !ok || (parsed < 0 && !strings.HasPrefix(key, "margin")) {
				return f, false
			}
			switch key {
			case "padding-left":
				f.PaddingLeft = parsed
			case "padding-right":
				f.PaddingRight = parsed
			case "margin-left":
				f.MarginLeft = parsed
			case "margin-right":
				f.MarginRight = parsed
			case "border-left-width":
				f.BorderLeftWidth = parsed
			case "border-right-width":
				f.BorderRightWidth = parsed
			}
		case "font", "font-stretch", "font-feature-settings", "font-variation-settings", "word-spacing", "white-space":
			return f, false
		}
	}
	return f, true
}

func cssLength(value string, size float64) (float64, bool) {
	factor := 1.0
	switch {
	case strings.HasSuffix(value, "rem"):
		value = strings.TrimSuffix(value, "rem")
		factor = 16
	case strings.HasSuffix(value, "em"):
		value = strings.TrimSuffix(value, "em")
		factor = size
	case strings.HasSuffix(value, "px"):
		value = strings.TrimSuffix(value, "px")
	case strings.HasSuffix(value, "%"):
		value = strings.TrimSuffix(value, "%")
		factor = size / 100
	}
	number, err := strconv.ParseFloat(value, 64)
	result := number * factor
	return result, err == nil && !math.IsNaN(result) && !math.IsInf(result, 0)
}

func blockFormat(node *html.Node) (textFormat, bool) {
	var ancestors []*html.Node
	for current := node; current != nil; current = current.Parent {
		ancestors = append(ancestors, current)
		if hasClass(current, "typography-prose") {
			break
		}
	}
	f := baseFormat()
	for i := len(ancestors) - 1; i >= 0; i-- {
		var ok bool
		f, ok = formatNode(ancestors[i], f)
		if !ok {
			return f, false
		}
	}
	return f.textOnly(), true
}

func annotateFormat(node *html.Node, f textFormat) error {
	encoded, err := json.Marshal(f)
	if err != nil {
		return err
	}
	setAttribute(node, "data-plan-inline", string(encoded))
	if f.edges() != 0 || f.Family == "article mono" || node.Data == "sup" || node.Data == "sub" {
		setAttribute(node, "data-plan-atom", "true")
	}
	return nil
}

func setAttribute(node *html.Node, key, value string) {
	for i := range node.Attr {
		if node.Attr[i].Key == key {
			node.Attr[i].Val = value
			return
		}
	}
	node.Attr = append(node.Attr, html.Attribute{Key: key, Val: value})
}

func transformedText(text string, f textFormat) string {
	switch f.Transform {
	case "uppercase":
		return strings.ToUpper(text)
	case "lowercase":
		return strings.ToLower(text)
	case "capitalize":
		wordStart := true
		return strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) {
				if wordStart {
					wordStart = false
					return unicode.ToUpper(r)
				}
			} else {
				wordStart = unicode.IsSpace(r)
			}
			return r
		}, text)
	}
	return text
}
