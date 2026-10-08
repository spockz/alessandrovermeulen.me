<!-- Preserve the imported archive's source provenance and public URL contract after retiring the old site. -->

# Content migration

The one-time migration is complete. The original Jekyll sources and Python importer remain recoverable from Git commit `1c746d6`, where the replacement was staged in `hugo-site/`. They are no longer required by the active checkout or its build. The current site lives at the repository root.

`migration.json` records historical source filenames, current content paths, canonical URLs, draft status, recovered or unresolved references, the source archive commit and the original downloads' SHA-256 hashes. Its old source paths describe provenance rather than files expected in the current checkout. Release validation compares the published downloads against those frozen hashes. The six source files from the four old Gist embeds are rendered as local fenced code blocks; no embed scripts remain in migrated content.

The migration contains 36 posts, one About page, and all three downloads (`cv.pdf`, `report.pdf`, and `report-on-getting-rid-of-js.pdf`). Of the posts, 34 are published and two remain drafts. Post dates without an explicit front matter date come from the filename. Explicit Jekyll dates take precedence; all dates use `Europe/Amsterdam`. Jekyll `published: false` posts remain in Hugo with `draft: true`.

## URL behavior

The 34 published posts retain the exact dated URLs from the live Hugo site, captured from [its article archive](https://alessandrovermeulen.me/post/) on 8 October 2026 in `scripts/legacy-post-urls.json`. Each imported post pins its canonical path with `url` front matter. Nine live slugs differ from the archived source filenames, so dates and filename slugs alone cannot reproduce the existing URLs. For example, FinagleCon uses `/2015/08/31/finaglecon/`, while the first folds article uses `/2009/12/17/haskell-datatypes-and-folds-part-i/`.

The temporary `/posts/<filename-slug>/` paths forward directly to those canonical URLs through Hugo's portable HTML aliases. Where a dated filename path differs from the live slug, it also forwards to the canonical article; this preserves older links found in the archived sources. These aliases use a canonical link and meta refresh, not an HTTP redirect. Drafts generate neither articles nor aliases. Category prefixes from the archived Jekyll configuration are not part of the live site's canonical URLs.

Internal article links are rewritten directly to the pinned canonical paths, including the Graphviz article's malformed relative link. Hugo includes only the canonical published URLs in `/sitemap.xml` and `/atom.xml`, and advertises the sitemap in `/robots.txt`. New posts default to `/posts/<slug>/`, without a publication date; dates describing the subject can remain in the filename or slug. Writing remains at `/posts/`. The build checks verify the captured URL contract, direct alias targets, draft exclusions and canonical sitemap/feed entries.

The historical About page remains unchanged in substance and displays a notice that it is preserved from the 2016 site and may be outdated. Its claims about current employment at ING, hobbies, and referee activity need an author review before being presented as current biography. Its photography link now points to `/category/photography/`.

The declarative programming article's FTP citation uses a raw HTML anchor to retain its original address; the theme's Markdown link renderer otherwise replaces the FTP URL with an invalid anchor. The FTP endpoint has not been checked for availability.

Links to the old `/projects/lhs2texhl/` homepage now point directly to the [surviving project repository](https://github.com/spockz/lhs2texhl), which is also linked by the original release articles.

## Media and other unresolved references

The repository contains no local `wp-content` assets. The four screenshot URLs below were checked at their original host and returned HTTP 404. The articles show a small unavailable notice in place of each broken screenshot; exact original URLs and statuses remain in `migration.json`:

- `/wp-content/2010/09/Screen-shot-2010-09-10-at-12.24.28.png`
- `/wp-content/2010/09/Screen-shot-2010-09-10-at-12.24.28-300x225.png`
- `/wp-content/2010/10/Screen-shot-2010-10-23-at-18.41.00.png`
- `/wp-content/2010/10/Screen-shot-2010-10-23-at-18.41.00-300x255.png`

An old link to `/wp-content/2009/09/README.html` also has no local copy and returned HTTP 404 at its original host. The broken hyperlink is removed and its label has an adjacent unavailable notice.

All four Flickr images were recovered into `static/images/archive/`, verified as decodable image files, and referenced locally in the posts. The two JPEGs were fetched over HTTPS from Flickr (HTTP 200, 639×413). The original hosts returned HTTP 429 for the two PNGs; their exact bytes were recovered from archived HTTPS snapshots of the original URLs (HTTP 200, 563×155 and 430×410):

- `https://farm8.staticflickr.com/7025/6767517435_b075d1c686_z.jpg`
- `https://farm8.staticflickr.com/7155/6767517241_eacd8ec0ed_z.jpg`
- `https://web.archive.org/web/20160528095748id_/http://farm8.staticflickr.com/7313/9243796534_60bb926e44_o.png`
- `https://web.archive.org/web/20160528051606id_/http://farm4.staticflickr.com/3804/9241070475_9d48236aa7_o.png`

The migrated HTML image elements use local paths, lazy loading, and intrinsic dimensions.

Raw HTML is retained for the theme's trusted-content rendering strategy. The Kramdown TOC token and trailing target attribute were removed, and the WordPress caption shortcode was converted to a figure with its caption. No gist embeds remain unresolved.
