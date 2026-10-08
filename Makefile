# Keep development and validation reproducible with the pinned Hugo release and project-local caches.
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
HUGO_VERSION := 0.166.0
export HUGO_CACHEDIR := $(CURDIR)/.cache/hugo
export GOMODCACHE := $(CURDIR)/.cache/go-mod
export GOCACHE := $(CURDIR)/.cache/go-build
TYPOGRAPHY := .cache/tools/typography
PREVIEW_BIND ?= 127.0.0.1
PREVIEW_BASE_URL ?= http://localhost:1313/
BROWSER_BASE_URL ?= $(PREVIEW_BASE_URL)
TYPOGRAPHY_SOURCES := $(wildcard scripts/typography/*.go scripts/typography/go.mod scripts/typography/go.sum scripts/typography/*.txt scripts/typography/*.md scripts/typography/linebreak/*.go scripts/typography/font/*.ttf)

.PHONY: check-hugo site-check typography-check build serve check check-browser release

check-hugo:
	@hugo version | python3 -c 'import sys; expected = "v$(HUGO_VERSION)"; actual = sys.stdin.read().split()[1].split("+")[0]; sys.exit(0 if actual == expected else "Expected Hugo " + expected + ", got " + actual)'

$(TYPOGRAPHY): $(TYPOGRAPHY_SOURCES)
	mkdir -p ".cache/tools"
	cd scripts/typography && go build -o "$(CURDIR)/$(TYPOGRAPHY)" .

typography-check:
	cd scripts/typography && go test ./... && go vet ./...
	node --test scripts/planned-layout.test.cjs

site-check:
	python3 -m unittest discover -s scripts -p '*_test.py'

build: check-hugo $(TYPOGRAPHY)
	hugo --minify --gc --panicOnWarning --cleanDestinationDir
	"$(TYPOGRAPHY)" process .cache/build

serve: check-hugo $(TYPOGRAPHY)
	"$(TYPOGRAPHY)" serve --bind "$(PREVIEW_BIND)" --baseURL "$(PREVIEW_BASE_URL)"

check: build site-check typography-check
	python3 scripts/check_site.py .cache/build

check-browser:
	node scripts/check_typography_browser.cjs "$(BROWSER_BASE_URL)"
	TYPOGRAPHY_BROWSER=webkit node scripts/check_selection_browser.cjs "$(BROWSER_BASE_URL)"

release: check-hugo $(TYPOGRAPHY)
	hugo --minify --gc --panicOnWarning --cleanDestinationDir --destination public
	"$(TYPOGRAPHY)" process public
	python3 scripts/check_site.py public
