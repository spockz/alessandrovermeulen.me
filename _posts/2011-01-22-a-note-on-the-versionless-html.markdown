--- 
layout: post
title: A note on the versionless HTML
tags: 
- web
- html5
- xhtml
- css
comments: true
---
Yesterday WhatWG announced to drop the versioning of HTML. HTML is to be a <a
href="http://blog.whatwg.org/html-is-the-new-html5">living standard</a>. This
comes at a very inconvenient time I fear.

Over the past few years web standards have, albeit slowly, developed into some
mature state: Most browsers support the current (x)HTML, CSS and JS standards to
the letter and can do more. However, due to the fact that these standards have
evolved so slowly new developments are not included in these standards. HTML5
(<a href="http://dev.w3.org/html5/spec/Overview.html">W3C</a> and <a
href="http://www.whatwg.org/specs/web-apps/current-work/multipage/">WhatWG</a>)
solves a lot of these problems, including features like `<canvas>` and
`<video>`.

With WhatWG dropping the version number of HTML I foresee a descent back to the
Dark Ages where every browser supported a different standard and supported the
same standards differently, one browser would support feature X and not Y while
the other supports Y but not X. This is a nightmare for developers: not being
able to rely on the uniformity of the browsers of your visitors. This created
necessities as writing specific CSS and JS code for specific browsers to `fix'
certain behaviour. These fixes mostly had to be done for Internet Explorer as it
had at the time by far the largest install base.

Now that we have finally arrived at an era where all browsers support the same
set of rather advanced techniques we decide to drop versions. In some cases it
is not a disaster to not have a version number. Look at CSS, the changes that
have been made to the standard were mostly additions. New operators and new
properties and values have been added. These changes required a better parser,
and the recognition of the new fields, but never did a semantic
<strong>change</strong>. This, together with the tradition that CSS engines
ignore the statements they don't understand has led to developers being able to
include new techniques in CSS files and thus implicitly setting the version of
the CSS in the file. As nothing changed older and newer browsers behaved the
same at the old statements and only the newer browsers acted on the new
statements. So up to now the absence of explicit CSS versioning didn't cause any
problems for CSS.

However, with the absence of versions in HTML5 we have a problem. Not only does
the HTML5 standard describe the structure of HTML but it also describes the JS
that the browser should support. These standards are in contrast to CSS subject
to the changing of their semantics. For example, a JavaScript function can get a
different interface: different parameters and/or return values, but also its
semantic proper might change.

This changing of semantics without version numbers causes a new nightmare for
developers, as they have no way to be sure that their site will look the same
across different browsers as they all may support a slightly different
incarnation of the standard and on top of that the developer cannot be sure his
site will look the same in the same browser a month from now, or even the next
day, as the browser can switch to support a newer instance of the standard. As
different versions of the same browser may have different engines that support
different incarnations of the standard you now also have to worry about the
differences between browser versions that your visitors use.

One might notice the presence of a lot of conditions in the above text and think
"What is the fuss, it might all not happen.". The main argument I would like to
make here is this: a `living' standard introduces a lot of uncertainties. At a
time where certainties are just starting to come back, giving developers peace
of mind, going back to living with uncertainties is certainly a descent into
Dark Ages in my opinion.

<h3>A solution?</h3>
There is of course a reason why they choose to declare this standard as
`living'. It denotes the fact that the web is evolving on a faster pace than we
are used to, it is quite dynamic. Having to wait several years before you can
use a very nice new feature on your website is not nice, as you would like to
use it as soon as it becomes defined. Therefore, instead of declaring the
standard completely dynamic and deprived of versions, I suggest something like
annual milestones. This way one can still compare browsers on their
capabilities, new features can be expected to be available rather quickly and
most importantly you provide a frame of reference for the developer.
