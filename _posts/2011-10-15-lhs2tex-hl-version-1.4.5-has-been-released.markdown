--- 
layout: post
title: lhs2TeX-hl version 1.4.5 has been released!
tags: 
- Haskell
- lhs2tex-hl
comments: true
---
I've just released a new version of <a href="http://alessandrovermeulen.me/projects/lhs2texhl/" title="lhs2TeX HighLighter ">lhs2TeX-hl</a>: 1.4.5 and it includes some new features:

<ol>
	<li>Added support for recursively traversing includes of .lhs files</li>
	<li>The program now doesn't fail completely when haskell-src-exts fails to parse
    a file. An error is reported and the program continues. :) A fmt file is
    still generated.</li>
	<li>functions that aren't functions but constants are now given the tag `constant'</li>
	<li>Removed a faulty command from the list.</li>
	<li>Cleaned up some code. (Probably introduced other ugly code) </li>
	<li>Binary operators are now typeset better. (I hope :))</li>
</ol>

Please let me know what you think.

