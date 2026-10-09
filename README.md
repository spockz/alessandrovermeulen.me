<!-- Keep development and deployment instructions beside the active Hugo sources. -->

# Alessandro Vermeulen's website

This repository contains the Hugo site using Blowfish v3.8.0. The site lives at the repository root on `hugo-blowfish`. The old Jekyll site and previous Cactus Plus site remain in Git history and on the `master` and `hugo` branches.

The design uses a text profile, recent writing, tags, syntax highlighting, light/dark appearance, and a small mobile menu. Article prose and article-list summaries use pinned local regular, bold, italic and bold italic faces, with a second local font for inline code inside typography scopes; navigation retains system fonts. Search, image zoom, reading progress, and analytics are disabled. There is no browser framework or production Node dependency.

## Development

Required tools: Hugo **0.166.0 extended**, Go **1.26 or later**, Python **3.11 or later**, Node **20 or later** for JavaScript checks, and Make. The theme and typography helper dependencies are pinned in their respective `go.mod` and `go.sum` files; the first build downloads those modules.

The Hugo release check accepts the pinned version with commit hashes and build metadata, as reported by Netlify and Homebrew binaries. Other releases and prereleases are rejected.

```sh
make serve
```

The preview is available at [localhost:1313](http://localhost:1313/). Published archive posts retain their exact dated URLs from the live site, including slugs that differ from their filenames. New posts automatically use `/posts/<title>/`, with the title converted to a URL segment. Publication dates in filenames do not enter these URLs; dates in titles remain. Writing remains at `/posts/`.

To share the preview with devices on your local network, bind it to this computer's LAN address and use that address in the browser-facing URL:

```sh
make serve PREVIEW_BIND=192.168.1.141 PREVIEW_BASE_URL=http://192.168.1.141:1313/
```

Replace the example address with the computer's current LAN address. Other devices can open the specified URL while the server is running and the computer is awake. The internal Hugo server stays on localhost; the shared preview retains article typography and live reload.

```sh
make check
```

This performs a minified build with warnings treated as errors, applies soft hyphens and line plans to every blog post and About, runs Hugo version, permalink, and download integrity tests, the Go helper's tests and static analysis and the browser guard's unit tests, then checks the captured live article URLs and compatibility aliases, default typography on all posts and About, draft exclusions, the sitemap and robots.txt, the complete Atom feed and its discovery link in every page head, local links and anchors, original download hashes, and page-weight budgets including the article fonts. Results are saved to `.cache/review/site.json`.

With the preview running, optional WebKit regression checks verify rendered typography at phone and desktop widths, native text selection and copying, resizing, and printing:

```sh
make check-browser
```

This uses `PREVIEW_BASE_URL` by default; override `BROWSER_BASE_URL` to check another running preview. It requires Playwright with its WebKit browser; see [PERFORMANCE.md](PERFORMANCE.md#repeat-the-browser-measurement) for module/cache options.

`make` and `make build` write review output to `.cache/build/`. They clean obsolete build files. `make release` writes the deployment output to `public/`. Module, build, and review caches stay inside `.cache/` and are ignored by Git.

## Content and configuration

- `content/posts/`: the imported archive, two retained drafts, and new writing.
- `content/about.md`: the current biography, with planned prose layout.
- `content/projects/index.md`: curated project cards, grouped under AI-adjacent, Service mesh-adjacent and Utilities headings. Edit the descriptions and shortcode order here; contributions, forks and roadmap features are labelled explicitly. Contribution cards describe the specific work and link to published upstream documentation. The page uses Blowfish's static `feature-grid`/`feature` cards, with a small local layout to use the available site width. It makes no GitHub API requests during builds or visits.
- `config/_default/languages.en.toml`: site description, author headline, and social links.
- `config/_default/params.toml`: appearance and theme features.
- `config/_default/menus.en.toml`: navigation.
- `static/`: original downloads, recovered images, and favicon.
- `layouts/rss.xml`: the original `/atom.xml` subscription endpoint. Every rendered page advertises this shared blog feed in its head; redirect pages are excluded.

The migration procedure, source-to-post mappings, and missing archive assets are recorded in [MIGRATION.md](MIGRATION.md) and `migration.json`.

The one-time importer and original sources are preserved at Git commit `1c746d6`. They are no longer part of the active checkout. The retained migration manifest records the original filenames, exact public URLs, recovered references and SHA-256 hashes of the two retained reports. The old CV download has been retired; About links to LinkedIn for current professional information. Builds and checks require only the current Hugo sources.

New posts can use `hugo new content posts/example.md`. The archetype creates a draft and removes a leading `YYYY-MM-DD-` from its suggested title; set `draft = false` to publish it. The checker verifies the imported archive and also includes new published posts in its feed and sitemap checks.

Hugo excludes posts with a future publication date, even when `draft = false`. Use the correct UTC offset for the intended publication time: Amsterdam uses `+02:00` during summer time and `+01:00` during winter time. `hugo list future` identifies scheduled posts.

Posts need no `url` or `slug` field. The shared `/posts/:slug/` permalink rule uses the post's title when `slug` is absent: `2026-10-08-a-fresh-leaf.md` with title “A fresh leaf” becomes `/posts/a-fresh-leaf/`. An explicit `slug` overrides the title; an explicit `url` overrides the entire pattern, preserving the imported archive's routes. Changing a title changes its generated URL, so use an optional `slug` when renaming a published article whose URL should stay fixed.

## Project emblems

Every project card uses an original SVG emblem from the same family. The reusable specification is [data/project-badges.toml](data/project-badges.toml): shared frame and stroke settings, each project's visual concept, and its SVG shapes. [scripts/generate_project_badges.py](scripts/generate_project_badges.py) generates the local theme icons in `assets/icons/project-*.svg`. It uses Python's standard library and makes no network requests.

Keep new symbols simple and geometric, centred within the 20–80 region of a 96-by-96 canvas. Use rounded three-unit strokes, open space, and one visual metaphor for the project's purpose. The family has a softly tinted rounded frame and no embedded text, shadows, or vendor logos. All colour uses `currentColor`, so the emblems follow Blowfish's light and dark palette. Presentation size and card integration are in `assets/css/custom.css`; cards retain the theme's feature grid.

To add or update an emblem, edit the shared specification and run:

```sh
make badges
make badges-check
```

Add a `[badges.<project-id>]` entry with a `concept` and `shapes` list, then use `icon="project-<project-id>"` on its feature card in `content/projects/index.md`. Supported shapes are paths, circles, rectangles, lines, polylines, and polygons. Commit the specification and generated SVGs together. The generator owns the `project-*.svg` prefix and removes generated icons whose definitions have been removed.

Build, preview, and release commands reject missing or stale generated badges. `make check` also verifies that every project card has exactly one distinct emblem from the catalog. `make check-browser` checks emblem sizing, distinct symbols, clipping, and card overflow at phone, tablet, and desktop widths in light and dark modes. The SVGs are inserted into the cards by Blowfish's existing icon partial, requiring no additional image requests or JavaScript.

## Paragraph typography

Planned typography is the default for every blog post, including new posts, and for the article descriptions on Writing, its pagination, home-page recent posts, and tag/category article lists. No front matter or shortcode is required for posts. About opts into the same layout with `typography: planned` in its front matter; other standalone prose pages can do the same. A head marker identifies these pages, and the shared preview/build transform scopes typography to their article content without copying the theme's single-page template. The transform also marks Blowfish's existing summary containers and preserves their article links and three-line clamp. Descriptions use the same pinned font and a maximum 576-pixel text column. Navigation, metadata, headings, and project cards retain native wrapping.

Single-column article pages center their title, prose, and article footer while keeping the 576-pixel prose limit. The table of contents stays above the centered column on tablets and phones; at desktop widths, articles with a table of contents retain the theme's two-column layout. Project cards use the full site width.

Use a shortcode to override a section's mode or to add typography outside blog posts:

```text
{{% typography mode="planned" lang="en-GB" %}}
Your paragraphs, with **formatting**, links, and `inline code`.
{{% /typography %}}
```

`mode="pretty"` asks the browser to choose better lines without word splitting. `mode="hyphenate"` allows build-time soft hyphens with ordinary wrapping. `mode="both"` combines them. These native modes justify prose blocks. Browser support determines how much `pretty` improves line choices; it is a hint, not a guarantee of TeX paragraph composition.

`mode="planned"` additionally composes article prose at build time, using the Go translation of Encarta's Typst-derived numerical optimizer and pure-Go HarfBuzz shaping. Mixed text runs use matching pinned fonts and their own sizes. Selected lines use signed `word-spacing`, allowing both tightening and stretching; final lines stay ragged unless they need tightening to fit. The numerical core preserves Encarta's practical search, pruning and retirement choices rather than introducing expensive exhaustive checks. Its pinned fixtures, Apache-2.0 license and provenance are in [scripts/typography/linebreak/](scripts/typography/linebreak/UPSTREAM.md).

Whole-word and discretionary hyphen breaks compete in the same search. The site's preferred stretching allowance is half the natural space advance measured in each run's font and size. This makes wide spaces more costly than the previous allowance of one quarter of the font size, so useful word splits are selected sooner. It is a scoring preference, not a hard cap: long identifiers and indivisible inline boxes can still force wider gaps. The numerical kernel and its hyphen penalties remain unchanged. Space advances are cached by text format and summed through a prefix table during the build.

Plans cover **every integer content width from 240 to 576 CSS pixels**, using each block's configured font size. Table cells and their child prose additionally cover widths down to 80 pixels. The theme caps the article at 576 pixels but makes narrower columns fluid. Breakpoints can also reduce available width: a 390-pixel viewport has a 342-pixel column, while the table-of-contents layout at 1024 pixels has a 508-pixel column. Selection uses each block's measured content width after subtracting its padding and borders. Every level of list indentation is reflected in the item's border-box width; its own padding reduces the text area further. FinagleCon list text is 32 pixels narrower than its article column. Deep items below the supported width use native wrapping.

Each paragraph stores its source text once, shared line metrics, shared sequences of line IDs, and a width-to-sequence mapping. Font shaping is cached across the width sweep during the build. The browser selects the plan for the integer immediately below its measured width, then adjusts signed spacing to the actual fractional width. The difference is less than one pixel and never requires extra shrinking. It rebuilds line spans only when the chosen sequence changes, retaining the original links. If resizing requires a new sequence while text is selected, that paragraph uses native wrapping until the selection clears; its text nodes are not replaced.

Planned line breaks and displayed hyphens are empty inline markers rendered with CSS pseudo-elements. They create no source newlines or hyphen characters, so native word and paragraph selection crosses the composed lines. Source spaces and soft hyphens at chosen breaks remain selectable at zero visual width. Original explicit breaks are preserved for native wrapping, printing and copying. The copy handler strips discretionary hyphens and plan markup from both whole paragraphs and individual words.

A small deferred script enables `.planned-layout` only after each required font face loads, declared block and inline styling matches, and actual browser measurements of every chosen line agree within 0.4 pixels. Resize, font-loading and style changes revalidate the plans without running the optimizer. Larger text, font changes, missing JavaScript or failed font loading retain continuous, justified text with soft hyphens. Widths outside the covered range and individual blocks whose chosen lines cannot fit with the allowed spacing also use native wrapping. Other eligible blocks remain planned. Printing uses native wrapping too.

Inline code uses **Article Mono at 0.875em**, with 5 pixels of padding on each side. Both font advances and padding count toward the line width. Code boxes stay whole in planned lines, including spaces and hyphens inside them. Their spaces remain fixed; only prose spaces stretch or shrink. The browser measures the code boxes with their actual font and padding and disables plans if they differ from the build contract.

The supported article elements retain their source formatting and identities:

| Article element                                              | Treatment                                                                                                                            |
| ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------ |
| Bold, italic and combined emphasis                           | Shape with the matching local face inside the same paragraph.                                                                        |
| Small text, superscripts, subscripts and footnote references | Shape at their own size; preserve reference IDs and links.                                                                           |
| Styled spans                                                 | Include supported font size, case transformation, letter spacing, horizontal padding, borders and margins. Edged boxes remain whole. |
| Blockquotes                                                  | Compose italic prose at its measured indented width.                                                                                 |
| Captions and footnote bodies                                 | Compose at 14 pixels. Footnote return links read “back” because the pinned font lacks the theme's return-arrow glyph.                |
| Compound and nested lists                                    | Compose direct parent prose and child paragraphs independently, at each level's available width.                                     |
| Explicit line breaks                                         | Preserve original breaks; end their lines ragged. Empty nested break sequences retain native wrapping.                               |
| Inline images                                                | Treat images with explicit matching widths as indivisible boxes; preserve the image node.                                            |
| Table cells and definition lists                             | Compose text at each cell or definition's width. Tables use fixed columns and 14-pixel text.                                         |
| Headings, code blocks, standalone images and rendered math   | Keep native block layout. Code blocks retain their original text, theme monospace font and scrolling.                                |

Unknown font or box styling, unsupported characters, unsupported inline elements, and blocks exceeding 5,000 characters retain native layout. Browser measurement rejects unmodeled generated content or CSS geometry rather than applying incorrect breaks. Emergency wrapping keeps long native strings within the article column. FinagleCon has **16 eligible paragraphs and three list items**. All activate at 342/310 and 576/544 pixel content widths; at 272/240 pixels one paragraph retains native wrapping. Links and anchors remain single nodes across chosen breaks, and the copy handler restores complete source text and formatting without generated breaks, discretionary hyphens, or layout metadata.

`mode="compare"` adds four keyboard-accessible choices over the same article, initially selecting **Planned**. Its text-size selector demonstrates the guard's fallback. The three native choices still work through CSS when JavaScript is unavailable.

Article Sans is a renamed Latin subset of **Source Sans 3 3.052**, chosen to retain the theme's sans-serif appearance. Its regular, bold, italic and bold italic WOFF2 files are **15,060**, **14,888**, **14,572** and **14,544 bytes** respectively. Article Mono is a renamed Latin subset of **Source Code Pro Regular 2.042**, with an **11,448-byte WOFF2**. The browser loads only faces used on a page. Fonts are served locally without `local()` substitution. Each TTF/WOFF2 pair is generated together and retains matching metrics and shaping tables. See [font provenance](scripts/typography/font/PROVENANCE.md) for licensing, hashes, character coverage and the occasional regeneration procedure. Font generation is not a normal build dependency.

The Go helper in `scripts/typography/` embeds British English TeX patterns and exceptions, inserts U+00AD into eligible HTML text nodes, and keeps code, URLs, attributes, headings, and identifiers intact. Only `en-GB` is currently supported. Pattern provenance and licensing are recorded alongside the data.

Use the Make commands for both preview and release. `make serve` runs Hugo internally on port 1315 and exposes the transformed preview at port 1313; Hugo's live reload is forwarded. `make build` and `make release` apply the same transform to generated HTML. A direct `hugo` command does not insert soft hyphens or line plans. Feed content retains the original words without generated soft hyphens or planned breaks.

## Release

Create the deployment artifact:

```sh
make release
```

It builds and checks `public/`. This works on macOS and Linux. Historical live URLs serve the articles directly; the temporary `/posts/<slug>/` paths and differing dated filename paths forward to them. Internal article links, the sitemap, Atom entries and canonical page links use the historical live URLs. See [MIGRATION.md](MIGRATION.md#url-behavior) for the captured URL contract and alias behavior.

The repository-root [`netlify.toml`](netlify.toml) configures Netlify to build from the repository root, run `make release`, and publish `public`. It pins Hugo, Go and Python to the versions used for the local release build. Do not replace the build command with plain `hugo`: that omits the typography transform and release checks. Commit the source files and configuration; `public/` remains ignored and is rebuilt by Netlify.

For the existing Netlify project, check **Project configuration > Developer settings > Continuous deployment** for its linked repository and production branch. Keep `master` and `hugo` as archives of the previous sites. The active site lives on `hugo-blowfish` and has no dependency on their files or theme submodules.

Push `hugo-blowfish`, then enable a branch deployment of it in the existing Netlify project. Once the preview passes review, select `hugo-blowfish` as Netlify's production branch. The existing project can retain its domain settings. Branch selection remains a Netlify project setting, separate from this build configuration.

Release output retains `https://alessandrovermeulen.me/` as its base URL in every deployment context. Preview pages use relative navigation and assets, while canonical links, feed entries and HTML alias redirects point to the production domain. Check historical article paths directly when reviewing a preview; following a compatibility alias opens the production site.

Performance results and the repeatable browser measurement are documented in [PERFORMANCE.md](PERFORMANCE.md).
