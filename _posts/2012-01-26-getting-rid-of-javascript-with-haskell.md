---
layout: post
title: "Getting rid of programming JavaScript with Haskell"
date: 2012-01-26 20:29
comments: true
tags: 
  - Haskell
  - JavaScript
  - Web
---

For my Experimentation Project at Utrecht University I ported the "JCU"
application to Haskell. The JCU application is used to give Dutch High school
students the opportunity to taste Prolog.

The project uses the Utrecht Haskell Compiler and its JavaScript backend. The UHC
translates Haskell to Core and then translates this Core language to JavaScript.
For more information on this see the blog of the creator of the 
[UHC JavaScript backend](http://utrechthaskellcompiler.wordpress.com/2010/10/18/haskell-to-javascript-backend/).

Please read my [report](/downloads/report-on-getting-rid-of-js.pdf) on this
project. The project is hosted on GitHub in the following repositories:

* [JCU](https://github.com/spockz/JCU),
* [uhc-jscript](https://github.com/spockz/uhc-jscript),
* [NanoProlog](https://github.com/spockz/NanoProlog), and
* [UU-TC](http://hackage.haskell.org/package/uu-tc).

<a href="http://www.flickr.com/photos/spockz/6767517435/" title="Incomplete proof in the HS JCU App by Alessandro Vermeulen, on Flickr"><img src="http://farm8.staticflickr.com/7025/6767517435_b075d1c686_z.jpg" width="640" height="414" alt="Incomplete proof in the HS JCU App"></a>
<a href="http://www.flickr.com/photos/spockz/6767517241/" title="Complete proof in the HS JCU App by Alessandro Vermeulen, on Flickr"><img src="http://farm8.staticflickr.com/7155/6767517241_eacd8ec0ed_z.jpg" width="640" height="414" alt="Complete proof in the HS JCU App"></a>

**update 28-01-2012**:
The keyword `jscript` in the UHC has been changed to `js` in order to avoid 
association with [Microsoft's JScript](http://en.wikipedia.org/wiki/JScript).
Also `new Object` syntax is now available in the `foreign import` directives.

``` haskell
foreign import js "new Object()"
  newObject :: IO (JSPtr a)
```