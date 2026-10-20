---
# WHY: Preserve this archived article's metadata and exact historical URL.
{
  "title": "UUAG TextMate Bundle released",
  "date": "2011-03-10T00:00:00+01:00",
  "slug": "uuag-textmate-bundle-released",
  "draft": false,
  "tags": [
    "haskell",
    "uuagc",
    "textmate",
    "attribute-grammars"
  ],
  "url": "/2011/03/10/uuag-textmate-bundle-released/",
  "aliases": [
    "/posts/uuag-textmate-bundle-released/"
  ]
}
---

I'm proud to announce the release of the <a title="GitHub repository of the
UUAGC TextMate Bundle"
href="https://github.com/spockz/UUAttributeGrammar-TextMate-Bundle">TextMate
bundle</a> for <a title="Utrecht University Attribute Grammar System"
href="http://www.cs.uu.nl/wiki/HUT/AttributeGrammarSystem">UUAG</a>.
It is
currently very simple and is based on the <a title="Haskell Bundle for TextMate"
href="http://weblog.jamisbuck.org/2005/11/1/haskell-bundle-for-textmate">Haskell
TextMate bundle</a> by Jamis Buck.

The only quirk is that you need to add `--|` after every sem block, otherwise
TextMate will consider everything afterward Haskell code.
This token is used to
indicate the end of the Haskell code that officially starts after the `=`.
This
has the effect that every `|` before every Constructor is coloured blue.
If you
want to avoid this, insert `--|` after every definition.

```
sem Foo
  | Bar lhs.cafe = ...
  | Hotel lhs.desk = ...
--|
```

Happy coding!
(<a
href="https://github.com/spockz/UUAttributeGrammar-TextMate-Bundle">GitHub repository</a>, develop: <a
href="https://github.com/spockz/UUAttributeGrammar-TextMate-Bundle/tarball/master">UUAG.tmbundle.tar.gz</a>)
