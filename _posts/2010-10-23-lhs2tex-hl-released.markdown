--- 
layout: post
title: lhs2TeX-hl released
tags: 
- Haskell
- lhs2tex
- syntax highlighting
- colour
- lhs2tex-hl
- open source
comments: true
---
I'm proud to announce the first release of my lhs2TeX-hl tool. For us who fancy using colours in our presentations or papers this should now go a whole lot easier. Go to the <a href="http://alessandrovermeulen.me/projects/lhs2texhl/">lhs2TeX-hl homepage</a>!

Install it like this:

[bash]
&gt; cabal install lhs2TeX-hl
[/bash]

lhs2TeX-hl is run before you run lhs2TeX and you supply it with the input file and the output file. A typical execution would look like this:

[bash]
&gt; lhs2TeX-hl -o MyPaper.fmt MyPaper.lhs
&gt; lhs2TeX -o MyPaper.Tex MyPaper.lhs
&gt; pdflatex MyPaper.tex
[/bash]

It's important to note here that you should've added the following line to your MyPaper.lhs file:
<pre>%include MyPaper.fmt</pre>
And that you have the following commands defined in your latex file:
<ul>
	<li> \lhsCHkeyword</li>
	<li>\lhsCHprelude</li>
	<li>\lhsCHtype</li>
	<li>\lhsCHlitNumber</li>
	<li>\lhsCHconstructor</li>
	<li>\lhsCHfunction</li>
	<li>\lhsCHinfixoperator</li>
</ul>
For example, like this:
<pre>\definecolor{datatype}{RGB}{42,0,217}
\definecolor{class}{RGB}{197,11,16}
\definecolor{fieldname}{RGB}{0,0,162}
\definecolor{prelude}{RGB}{64,80,117}
\definecolor{numeral}{RGB}{0,0,205}
\definecolor{infixoperator}{RGB}{42,0,217}
\definecolor{constructor}{RGB}{0,161,0}
\definecolor{keyword}{RGB}{229,120,0}
\definecolor{special1}{RGB}{159,138,0}

\newcommand{\lhsCHfunction}[1]{\color{infixoperator}{{#1}}}
\newcommand{\lhsCHinfixoperator}[1]{\color{infixoperator}{{#1}}}
\newcommand{\lhsCHprelude}[1]{\color{prelude}{\mathbf{#1}}}
\newcommand{\lhsCHkeyword}[1]{\color{keyword}{\textbf{#1}}}
\newcommand{\lhsCHconstructor}[1]{\color{constructor}{\textbf{#1}}}
\newcommand{\lhsCHlitNumber}[1]{\color{numeral}{{#1}}}
\newcommand{\lhsCHtype}[1]{\color{datatype}{{#1}}}</pre>
And then you might end up with something that looks like this:

[caption id="attachment_306" align="aligncenter" width="300" caption="Literate Highlighter Code v0.1.0.2"]<a href="http://alessandrovermeulen.me/wp-content/2010/10/Screen-shot-2010-10-23-at-18.41.00.png"><img class="size-medium wp-image-306" title="Screen shot 2010-10-23 at 18.41.00" src="http://alessandrovermeulen.me/wp-content/2010/10/Screen-shot-2010-10-23-at-18.41.00-300x255.png" alt="Literate Highlighter Code v0.1.0.2" width="300" height="255" /></a>[/caption]
