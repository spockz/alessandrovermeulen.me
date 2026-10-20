---
# WHY: Preserve this archived article's metadata and exact historical URL.
{
  "title": "Getting rid of programming JavaScript with Haskell",
  "date": "2012-01-26T20:29:00+01:00",
  "slug": "getting-rid-of-javascript-with-haskell",
  "draft": false,
  "tags": [
    "haskell",
    "javascript"
  ],
  "url": "/2012/01/26/getting-rid-of-programming-javascript-with-haskell/",
  "aliases": [
    "/posts/getting-rid-of-javascript-with-haskell/",
    "/2012/01/26/getting-rid-of-javascript-with-haskell/"
  ]
}
---

For my Experimentation Project at Utrecht University I ported the "JCU"
application to Haskell.
The JCU application is used to give Dutch High school
students the opportunity to taste Prolog.

The project uses the Utrecht Haskell Compiler and its JavaScript backend.
The UHC
translates Haskell to Core and then translates this Core language to JavaScript.
For more information on this see the blog of the creator of the
[UHC JavaScript backend](http://utrechthaskellcompiler.wordpress.com/2010/10/18/haskell-to-javascript-backend/).

Please read my [report](/downloads/report-on-getting-rid-of-js.pdf) on this
project.
The project is hosted on GitHub in the following repositories:

* [JCU](https://github.com/spockz/JCU),
* [uhc-jscript](https://github.com/spockz/uhc-jscript),
* [NanoProlog](https://github.com/spockz/NanoProlog), and
* [UU-TC](http://hackage.haskell.org/package/uu-tc).

<a href="http://www.flickr.com/photos/spockz/6767517435/" title="Incomplete proof in the HS JCU App by Alessandro Vermeulen, on Flickr"><img src="/images/archive/6767517435_b075d1c686_z.jpg" width="639" height="413" alt="Incomplete proof in the HS JCU App" loading="lazy"></a>
<a href="http://www.flickr.com/photos/spockz/6767517241/" title="Complete proof in the HS JCU App by Alessandro Vermeulen, on Flickr"><img src="/images/archive/6767517241_eacd8ec0ed_z.jpg" width="639" height="413" alt="Complete proof in the HS JCU App" loading="lazy"></a>

**update 28-01-2012**:
The keyword `jscript` in the UHC has been changed to `js` in order to avoid
association with [Microsoft's JScript](http://en.wikipedia.org/wiki/JScript).
Also `new Object` syntax is now available in the `foreign import` directives.

``` haskell
foreign import js "new Object()"
  newObject :: IO (JSPtr a)
```
