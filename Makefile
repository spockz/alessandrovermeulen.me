# Keep development and validation reproducible with the pinned Hugo release and project-local caches.
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := build
HUGO_VERSION := 0.166.0
export HUGO_CACHEDIR := $(CURDIR)/.cache/hugo
export GOMODCACHE := $(CURDIR)/.cache/go-mod
export GOCACHE := $(CURDIR)/.cache/go-build
TYPOGRAPHY := .cache/tools/typography
PREVIEW_BIND ?= 127.0.0.1
PREVIEW_BASE_URL ?= http://localhost:1313/
BROWSER_BASE_URL ?= $(PREVIEW_BASE_URL)
TYPOGRAPHY_SOURCES := $(wildcard scripts/typography/*.go scripts/typography/go.mod scripts/typography/go.sum scripts/typography/*.txt scripts/typography/*.md scripts/typography/linebreak/*.go scripts/typography/font/*.ttf)

.PHONY: check-hugo badges badges-check site-check typography-check build serve check check-browser release format lint vale

check-hugo:
	@version="$$(hugo version)" || exit "$$?"; \
	python3 -c 'import re, sys; expected = "v$(HUGO_VERSION)"; fields = sys.argv[1].split(); actual = fields[1] if len(fields) > 1 else "(missing version)"; release = re.sub(r"-[0-9a-f]{7,40}$$", "", actual.split("+")[0]); sys.exit(0 if release == expected else "Expected Hugo " + expected + ", got " + actual)' "$$version"

$(TYPOGRAPHY): $(TYPOGRAPHY_SOURCES)
	mkdir -p ".cache/tools"
	cd scripts/typography && go build -o "$(CURDIR)/$(TYPOGRAPHY)" .

badges:
	python3 scripts/generate_project_badges.py

badges-check:
	python3 scripts/generate_project_badges.py --check

typography-check:
	cd scripts/typography && go test ./... && go vet ./...
	node --test scripts/planned-layout.test.cjs

site-check:
	python3 -m unittest discover -s scripts -p '*_test.py'

format:
	python3 scripts/sembr.py -w content/posts/*.md

lint: vale
	python3 scripts/sembr.py --check content/posts/*.md

build: check-hugo badges-check $(TYPOGRAPHY)
	hugo --minify --gc --panicOnWarning --cleanDestinationDir
	"$(TYPOGRAPHY)" process .cache/build

serve: check-hugo badges-check $(TYPOGRAPHY)
	"$(TYPOGRAPHY)" serve --bind "$(PREVIEW_BIND)" --baseURL "$(PREVIEW_BASE_URL)"

vale:
	vale content/posts

check: build site-check typography-check vale
	python3 scripts/check_site.py .cache/build

check-browser: $(TYPOGRAPHY)
	node scripts/check_typography_browser.cjs "$(BROWSER_BASE_URL)"
	TYPOGRAPHY_BROWSER=webkit node scripts/check_selection_browser.cjs "$(BROWSER_BASE_URL)"
	node scripts/check_layout_performance.cjs "$(BROWSER_BASE_URL)"

release: check-hugo badges-check $(TYPOGRAPHY)
	hugo --minify --gc --panicOnWarning --cleanDestinationDir --destination public
	"$(TYPOGRAPHY)" process public
	python3 scripts/check_site.py public
