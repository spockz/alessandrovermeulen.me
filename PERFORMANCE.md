<!-- Record reproducible loading measurements and their limits beside the configuration that controls them. -->

# Loading performance

Transfer sizes updated on 9 October 2026 using Hugo 0.166.0 extended and Blowfish v3.8.0. These are local lab measurements, not measurements of the deployed site.

## Transfer sizes

`make check` measures each minified file with gzip level 9 and sums the HTML, referenced CSS/JavaScript, favicon, and local dependencies referenced by CSS. It counts the regular, bold, italic, bold italic and code font faces used by article text, inferred from HTML styles and generated font contracts. This is a transfer estimate rather than browser network instrumentation. Inline SVG data in CSS is included in the stylesheet's size. HTTP headers are excluded; the estimates require compression at the host. WOFF2 is already compressed; each gzip estimate is 23 bytes larger than its raw file.

| Page                           | Total gzip bytes |   HTML | Shared CSS | Shared JavaScript | Layout guard | Article fonts | Favicon |
| ------------------------------ | ---------------: | -----: | ---------: | ----------------: | -----------: | ------------: | ------: |
| Home                           |           70,442 | 24,255 |     22,997 |             2,405 |        5,428 |        15,083 |     274 |
| Projects                       |           64,591 |  8,921 |     22,997 |             2,405 |            0 |        29,994 |     274 |
| Writing                        |           94,996 | 48,809 |     22,997 |             2,405 |        5,428 |        15,083 |     274 |
| Shallow/deep embedding article |          137,213 | 50,049 |     22,997 |             2,405 |        5,428 |        56,060 |     274 |
| FinagleCon comparison          |          130,303 | 57,734 |     22,997 |             2,405 |        5,428 |        41,465 |     274 |

Article pages use a local 15,060-byte regular prose WOFF2 and load matching bold (14,888 bytes), italic (14,572 bytes) or bold italic (14,544 bytes) faces when used. Inline code in typography scopes requests an 11,448-byte monospace WOFF2. Article summaries also use the regular prose font; the home profile and navigation retain system fonts. There are two shared JavaScript files for appearance, navigation, and code copying. Blog posts and pages containing article summaries additionally load a 5,428-byte gzipped geometry guard. The optimizer runs only in Go during build/preview; there is no browser optimizer or Go WebAssembly runtime. CSS and JavaScript filenames include content fingerprints, and font filenames are versioned, so they can be cached across article visits. Each representative page's local transfer estimate excludes the two deferred external CI badges on Projects, which are listed separately in the report.

The Projects page uses Blowfish's feature grid and card styling and loads the regular and bold article fonts. Each project has a distinct inline SVG emblem generated from the shared local style catalog. Language pills beside titles and unconfigured CI labels are local HTML. The two configured workflows use GitHub's live status SVGs with lazy loading and reserved space. Builds make no GitHub requests, and the labels add no JavaScript. WebKit checks cover emblem sizing, distinct symbols, clipping, overflow, and title pills, and CI status aligned right in the source-link row at phone, tablet, and desktop widths in light and dark modes.

The automated budgets for these five representative pages are under 100,000 gzip bytes of HTML/CSS/JavaScript and under 200,000 gzip bytes in total. Font and stylesheet image dependencies are checked and counted. Image-bearing archive articles have different totals; these numbers do not cover every article.

Responsive plans for plain paragraphs added approximately 35 KB compressed versus the original single-width prototype. Adding lists and paragraphs containing code added approximately 21 KB more, including the 11.4 KB font. Mixed formatting adds approximately 17 KB to FinagleCon and 46 KB to the shallow/deep article compared with that earlier implementation, mostly matching font faces and newly eligible text plans. The build covers every integer content width from 240 through 576 pixels, plus table prose down to 80 pixels, sharing line metrics and repeated sequences rather than duplicating article text. The browser selects stored sequences and measures spacing; resizing runs no line-breaking search.

Default typography covers all 34 published posts, generating 390 eligible prose plans. Browser checks of every post activated 388 blocks at a 1280-pixel desktop viewport and 383 at a 390-pixel phone viewport; the remaining planned blocks used native wrapping when the selected lines could not fit the spacing constraints. Every post had active planned text, and none had horizontal page overflow. Quotes, footnotes, captions, emphasis and styled text now participate. IDs, links and code-block text remained identical across widths, and no code block contained planned markup. The Go tests also verify unchanged code-block source. Evidence is saved in `.cache/review/mixed-layout-browser.json`.

Additional mixed-prose fixtures verified padded and bordered inline boxes, a fixed-width inline image, explicit breaks inside emphasis, four levels of lists, table cells and definition lists. Table cells selected their own widths after padding; a cell with an indivisible code box fell back at a width where no valid plan existed. Copying the shallow/deep article preserved its footnote reference and formatting with no soft hyphens, generated breaks or plan metadata. Clipboard evidence is in `.cache/review/mixed-prose-clipboard.json`, and a quote screenshot is in `.cache/review/mixed-prose-akka-desktop.jpg`. Temporary verification pages were removed from the site's static files.

Earlier FinagleCon verification on 7 October covered ten viewport widths: 320, 375, 390, 430, 639, 640, 768, 1023, 1024, and 1280 pixels. All 16 eligible paragraphs activated at columns of 327, 342, 382, 508, 528, and 576 pixels; all three list items activated at their corresponding 32-pixel narrower content widths. At a 272-pixel column, 15 paragraphs and all three 240-pixel list items activated. One paragraph retained native wrapping because its chosen plan required excessive spacing. Every nonfinal planned line matched its content width within 0.01 pixels after signed spacing, including code-box padding. There was no horizontal overflow, and the article retained its 11 links and ten code boxes across all widths.

An additional temporary CSS check gave a paragraph 12/8-pixel padding and 2-pixel borders on both sides. Its 576-pixel border box correctly selected a 552-pixel plan. The temporary styling was removed after verification.

Selecting 20-pixel text disabled plans, returning to 16 pixels re-enabled them, and switching between native and planned modes worked after resizing. Browser copy verification at the 342-pixel column found all 19 complete source blocks with no generated breaks or soft hyphens; copied HTML also omitted generated markup and layout classes. No browser warnings or errors were observed. Measurements and screenshots are saved in `.cache/review/list-plans.json` and `.cache/review/finaglecon-list-planned-*.jpg`.

WebKit verification on 7 October reproduced the iPhone report at a 402-pixel viewport and 354-pixel article column. Native `text-wrap: pretty` and justification were overriding planned typography through CSS specificity, causing composed lines to wrap again. The planned and print selectors now take precedence. Fifteen WebKit checks across three articles and five viewport widths passed without extra visual lines or page overflow. Links and code blocks survived resizing unchanged, and printing used native wrapping. FinagleCon kept 16 active blocks at the reported phone width; three blocks retained native wrapping under the existing font-metric guard. This is browser-engine verification, not a physical-device test. Evidence is in `.cache/review/webkit-typography.json`.

On 8 October the site's stretching allowance changed from one quarter of the font size to half the actual natural space advance in each text run. The numerical optimizer and its default hyphen penalties remain unchanged. A controlled search regression selects a hyphenated path with 1 pixel of added spacing over a fitting whole-word path requiring 8 pixels, with incumbent pruning both enabled and disabled. The complete HTML-transform regression uses FinagleCon's Futures paragraph at 354 pixels: two selected word splits reduce the largest spacing addition from 4.63 to 2.78 pixels. The regression was also run with the old spacing allowance and with hyphen candidates removed; both mutations failed for excessive spacing and missing selected splits. Evidence is in `.cache/review/hyphen-regression-*.log`.

Across FinagleCon's generated phone plans, using 354-pixel paragraphs and 322-pixel list items, selected splits increased from 3 to 12 and median added spacing fell from 2.20 to 1.08 pixels. Lines adding more than 3 pixels per space fell from 25 to 12. These are generated-plan metrics, including blocks that may retain native wrapping under the browser's font guard; they are not a guarantee that every line stays below a fixed spacing cap. The comparison is saved in `.cache/review/finagle-phone-hyphen-summary.json`.

The updated plans passed `make check` and all 15 WebKit checks, retaining native fallback, print wrapping, links and code-block source. More distinct line sequences add approximately 4 KB compressed HTML to FinagleCon and 3 KB to the shallow/deep article; browser JavaScript and fonts were unchanged by that spacing change. Space measurements are cached once per text format and added through a prefix table, so the existing search still runs only during builds and preview responses.

Selection and summary coverage were also updated on 8 October. Generated line breaks and hyphens are now CSS presentation, while source word spacing remains in the selectable text. WebKit word and paragraph selection, clean copying and selection preservation during resizing passed. Writing, home recent posts and tag/category article summaries receive responsive plans, retaining the theme's three-line clamp and article links. Thirty WebKit layout checks across three articles and three listing pages passed. The summary plans, font and guard add approximately 38 KB to the home page's estimated compressed transfer. Evidence is saved in `.cache/review/webkit-selection.json` and `.cache/review/webkit-typography.json`.

## Mobile rendering

The following measurements are from **6 October, before the article font and planned-layout prototype**. They have not been rerun for the current font configuration. Chromium 151 headless ran three cold-cache visits to each page at a 390×844 viewport, 1.6 Mbps download throughput, 150 ms network latency, and 4× CPU slowdown. The local Python HTTP server served the minified files **without compression**. First Contentful Paint (FCP) and Largest Contentful Paint (LCP) were equal in these runs because the main visible content is text.

| Page                           | Median FCP | Median LCP | LCP across three runs    | Layout shift (CLS) |
| ------------------------------ | ---------: | ---------: | ------------------------ | -----------------: |
| Home                           |   1,076 ms |   1,076 ms | 1,172 / 1,076 / 1,072 ms |                  0 |
| Shallow/deep embedding article |   1,204 ms |   1,204 ms | 1,344 / 1,192 / 1,204 ms |                  0 |

The checks also confirmed no failed requests, browser exceptions, third-party resource requests, or horizontal page overflow on these pages. Desktop light/dark appearance was checked at 1440×1000. Browser caching, HTTP compression, DNS, TLS, host latency, and actual device performance will affect deployed results.

Raw evidence is in the ignored `.cache/review/site.json` and `.cache/review/browser.json`; screenshots are saved in that directory.

## Repeat the browser measurement

Run `make check`, then start a second terminal in this directory:

```sh
python3 -m http.server 1314 --bind 127.0.0.1 --directory .cache/build
```

With Node 20 or later, Playwright, and its Chromium browser installed:

```sh
node scripts/measure_browser.cjs http://127.0.0.1:1314/
```

If Playwright is installed outside normal Node module resolution, set `PLAYWRIGHT_MODULE` to that package's absolute path. Set `PLAYWRIGHT_BROWSERS_PATH` when using a non-default browser cache. The script uses an isolated browser and does not control the interactive preview.

With Playwright's WebKit browser installed and `make serve` running, check the responsive typography separately:

```sh
make check-browser
```

The target runs both WebKit scripts against `PREVIEW_BASE_URL`, which defaults to `http://localhost:1313/`. Set `BROWSER_BASE_URL` to check another running preview, for example `make check-browser BROWSER_BASE_URL=http://192.168.1.141:1313/`. The target uses the same `PLAYWRIGHT_MODULE` and `PLAYWRIGHT_BROWSERS_PATH` environment options described above.

The layout script checks actual rendered line fragments at 320, 390, 402, 834, 874, 1024, and 1280 pixels on four articles, four listing pages, and About, including links, inline code, and quotes. Every route must retain active planned text at each width. Single-column article pages must be centered at their existing prose width, with their headers and article footers aligned. Writing, its pagination, tag articles, and home-page recent articles must also be centered at that width. The home profile retains its original layout. It also checks resizing, horizontal overflow, and native print wrapping. Subpixel width differences up to one CSS pixel are allowed; extra visual lines are not. The selection script checks native word and paragraph selection, clean clipboard output, resizing with an active selection, and selectable Writing descriptions. Evidence is saved to `.cache/review/webkit-typography.json` and `.cache/review/webkit-selection.json`, with phone-width screenshots in `.cache/review/webkit-iphone-fixed.png` and `.cache/review/webkit-writing-planned.png`, and a home-page screenshot in `.cache/review/webkit-home-centered.png`.

The site uses lowercase post URLs, so the same build and validation commands work on macOS and Linux.

The layout script also checks cold font loading on About and home summaries at phone and desktop widths. It holds font requests open and captures actual pixels without Playwright's usual wait for fonts: no fallback glyphs may appear during the blocking period, and loaded text must be visible at the same 16-pixel size. The regular face must use one request shared by the preload and CSS. Failed font requests and disabled JavaScript must retain readable native text. Results are saved to `.cache/review/webkit-font-loading.json`.
