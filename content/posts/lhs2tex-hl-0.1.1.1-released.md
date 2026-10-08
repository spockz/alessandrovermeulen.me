---
# WHY: Preserve this archived article's metadata and exact historical URL.
{
  "title": "lhs2TeX-hl 0.1.1.1 released",
  "date": "2010-11-01T00:00:00+01:00",
  "slug": "lhs2tex-hl-0.1.1.1-released",
  "draft": false,
  "tags": [
    "haskell",
    "lhs2tex",
    "lhs2tex-hl",
    "open source"
  ],
  "url": "/2010/11/01/lhs2tex-hl-0.1.1.1-released/",
  "aliases": [
    "/posts/lhs2tex-hl-0.1.1.1-released/"
  ]
}
---

After the <a
href="/2010/10/23/lhs2tex-hl-released/">first
release</a> of my tool on hackage <a
href="http://hackage.haskell.org/package/lhs2TeX-hl-0.1.1.1">this release</a>
actually is a working package. The previous one didn't install out of the box.

The most important `changes ` are:

1. Added a filter for rewrite rules that aren't "lhs2TeX safe", e.g. `format () = ...`;
1. I hardcoded the default formatting rules for several of lhs2TeX defaults,
   including but not limited to `->`, `<-` and `=>`. This is not a desirable
   solution but it suits my purposes.

You can do literal (numeral, character and string) formatting easily with
lhs2TeX itself with the following directives:

``` latex
%subst char a    	= "\color{char}\text{\tt ''" a "''}"
%subst string a  	= "\color{string}\text{\tt \char34 " a "\char34}"
%subst numeral a =  "\color{numeral}{ " a " }"
```
