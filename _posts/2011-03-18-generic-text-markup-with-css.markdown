---
layout: post
title: Generic Text Markup with CSS
tags:
- xhtml
- css
- LaTeX
- open source
comments: true
---
Most of you are familiar with CSS reset scripts to reset CSS behaviour to
something uniform across several browsers. I have been searching for a good CSS
stylesheet that would do the same for the formatting of my text. That is, to
give the text on a site a markup, suitable for reading. As a great fan of LaTeX
this markup should follow the default markup of LaTeX as closely as possible.

As I could not find such a stylesheet I build my own. It probably is far from
finished, although I tried to cover the most common cases in this script.

In time there will be a git <a href="http://github.com">repository</a> that will
contain test cases and updated versions of this stylesheet.

I tested this script in the presence of the <a title="CSS Tools: Reset CSS"
href="http://meyerweb.com/eric/tools/css/reset/">Meyer CSS reset</a>, but it
should also work without any CSS Reset present, or together with other CSS reset
stylesheets.

Please let me know (either by mail or a reply) if you have any remarks!
{% gist 876971 %}