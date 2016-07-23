--- 
layout: post
title: UUAG TextMate Bundle released
tags: 
- Haskell
- uuagc
- textmate
- attribute grammars
comments: true
---
I'm proud to announce the release of the <a title="GitHub repository of the
UUAGC TextMate Bundle"
href="https://github.com/spockz/UUAttributeGrammar-TextMate-Bundle">TextMate
bundle</a> for <a title="Utrecht University Attribute Grammar System"
href="http://www.cs.uu.nl/wiki/HUT/AttributeGrammarSystem">UUAG</a>. It is
currently very simple and is based on the <a title="Haskell Bundle for TextMate"
href="http://weblog.jamisbuck.org/2005/11/1/haskell-bundle-for-textmate">Haskell
TextMate bundle</a> by Jamis Buck.

The only quirk is that you need to add `--|` after every sem block, otherwise
TextMate will consider everything afterward Haskell code. This token is used to
indicate the end of the Haskell code that officially starts after the `=`. This
has the effect that every `|` before every Constructor is coloured blue. If you
want to avoid this, insert `--|` after every definition.

```
sem Foo
  | Bar lhs.cafe = ...
  | Hotel lhs.desk = ...
--|
```

Happy coding! (<a
href="https://github.com/spockz/UUAttributeGrammar-TextMate-Bundle">GitHub repository</a>, develop: <a
href="https://github.com/spockz/UUAttributeGrammar-TextMate-Bundle/tarball/master">UUAG.tmbundle.tar.gz</a>)
