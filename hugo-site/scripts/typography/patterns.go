// This file embeds the selected language patterns so builds do not depend on network access.
package main

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/speedata/hyphenation"
)

//go:embed hyph-en-gb.pat.txt
var britishEnglishPatterns string

//go:embed hyph-en-gb.hyp.txt
var britishEnglishExceptions string

type hyphenator struct {
	patterns   *hyphenation.Lang
	exceptions map[string][]int
}

var britishEnglish struct {
	sync.Once
	value *hyphenator
	err   error
}

func newHyphenator(language string) (*hyphenator, error) {
	if !strings.EqualFold(language, "en-GB") {
		return nil, fmt.Errorf("unsupported hyphenation language %q (supported: en-GB)", language)
	}
	britishEnglish.Do(loadBritishEnglish)
	return britishEnglish.value, britishEnglish.err
}

func loadBritishEnglish() {
	patterns, err := hyphenation.New(strings.NewReader(britishEnglishPatterns))
	if err != nil {
		britishEnglish.err = fmt.Errorf("load en-GB hyphenation patterns: %w", err)
		return
	}
	exceptions := make(map[string][]int)
	for _, exception := range strings.Fields(britishEnglishExceptions) {
		var word []rune
		var breaks []int
		for _, r := range exception {
			if r == '-' {
				breaks = append(breaks, len(word))
			} else {
				word = append(word, r)
			}
		}
		exceptions[strings.ToLower(string(word))] = breaks
	}
	britishEnglish.value = &hyphenator{patterns: patterns, exceptions: exceptions}
}

func (h *hyphenator) breakPoints(word string) []int {
	if positions, ok := h.exceptions[strings.ToLower(word)]; ok {
		return positions
	}
	return h.patterns.Hyphenate(word)
}
