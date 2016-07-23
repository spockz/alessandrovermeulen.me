--- 
layout: post
title: Installing vacuum-cairo with the help of HomeBrew
tags: 
- Haskell
- ghc
- cabal
comments: true
---
I tried installing vacuum-cairo on OS X 10.6 but it failed. With the help of <a
href="http://mxcl.github.com/homebrew/" title="HomeBrew">HomeBrew</a> I
installed the dependencies by doing the following:

Be sure to run `brew update` first.

Then install the dependencies:

``` bash
brew install pkg-config
brew install glib
brew install librsvg
chmod u+w /usr/local/Cellar/gdk-pixbuf/2.23.5/lib/gdk-pixbuf-2.0/2.10.0/loaders.cache
brew install librsvg
brew install pango
```

Note the installing of librsvg twice, this is in accordance with https://github.com/mxcl/homebrew/issues/4970. If you don't get an error after running the first install of librsvg you might have a new update brew of librsvg.

And finally install vacuum-cairo:

``` bash
cabal install vacuum-cairo
```
