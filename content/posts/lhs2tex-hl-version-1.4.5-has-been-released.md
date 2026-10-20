---
# WHY: Preserve this archived article's metadata and exact historical URL.
{
  "title": "lhs2TeX-hl version 1.4.5 has been released!",
  "date": "2011-10-15T00:00:00+02:00",
  "slug": "lhs2tex-hl-version-1.4.5-has-been-released",
  "draft": false,
  "tags": [
    "haskell",
    "lhs2tex-hl"
  ],
  "url": "/2011/10/15/lhs2tex-hl-version-1.4.5-has-been-released/",
  "aliases": [
    "/posts/lhs2tex-hl-version-1.4.5-has-been-released/"
  ]
}
---

I've just released a new version of <a href="https://github.com/spockz/lhs2texhl" title="lhs2TeX HighLighter ">lhs2TeX-hl</a>: 1.4.5 and it includes some new features:

<ol>
	<li>Added support for recursively traversing includes of .lhs files</li>
	<li>The program now doesn't fail completely when haskell-src-exts fails to parse
a file.
An error is reported and the program continues.
:) A fmt file is
still generated.</li>
	<li>functions that aren't functions but constants are now given the tag `constant'</li>
	<li>Removed a faulty command from the list.</li>
	<li>Cleaned up some code. (Probably introduced other ugly code) </li>
	<li>Binary operators are now typeset better. (I hope :))</li>
</ol>

Please let me know what you think.
