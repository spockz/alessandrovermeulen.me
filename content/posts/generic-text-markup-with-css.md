---
# WHY: Preserve this archived article's metadata and exact historical URL.
{
  "title": "Generic Text Markup with CSS",
  "date": "2011-03-18T00:00:00+01:00",
  "slug": "generic-text-markup-with-css",
  "draft": false,
  "tags": [
    "xhtml",
    "css",
    "latex",
    "open-source"
  ],
  "url": "/2011/03/18/generic-text-markup-with-css/",
  "aliases": [
    "/posts/generic-text-markup-with-css/"
  ]
}
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
`gistfile1.css`

```css
/*** Generic Text Markup ***/
/*** Version: 0.2
     Date: 18-3-2011
     Author: Alessandro Vermeulen <alessandro@alessandrovermeulen.me>
     Licence:
       Copyright (c) 2011 Alessandro Vermeulen <alessandro@alessandrovermeulen.me>

       Permission is hereby granted, free of charge, to any person obtaining a copy
       of this software and associated documentation files (the "Software"), to deal
       in the Software without restriction, including without limitation the rights
       to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
       copies of the Software, and to permit persons to whom the Software is
       furnished to do so, subject to the following conditions:

       The above copyright notice and this permission notice shall be included in
       all copies or substantial portions of the Software.

       THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
       IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
       FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
       AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
       LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
       OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
       THE SOFTWARE.
***/
h1, h2, h3, h4, h5, h6 {
  font-weight: bold;
}

p {
  text-align: justify;
}

a {
  text-decoration:none;
}

p a {
  font-weight: bold;
}

a:hover {
  text-decoration: underline;
}

h1 {
  font-size:14pt;
}

h2 {
  font-size: 13pt;
}

h3 {
  font-size: 12pt;
}

h4, h5, h6 {
  font-size: 11pt;
}

/* Give head space to text items */
* + p, * + ul, * + ol, * + table {  
  margin-top: 10px;
}

/* Give headers space from their preceeding text */
* + h2, * + h3, * + h4, * + h5, * + h6 {
  margin-top: 10px;
}

/* Align elements tightly to their headers, overriding previous statement
   of give elements head space. */
h1 + *, h2 + *, h3 + *, h4 + *, h5 + *, h6 + * {
  margin-top:0;
}

ul {
  list-style-type: disc;
  list-style-position: inside;
}

ol {
  list-style-type: decimal;
  list-style-position: inside;
}

table {
  margin: 0 auto;
  width: 90%;
}
/*** End Generic Text Markup ***/
```

`gistfile2.html`

```html
<h1>A level one (1) heading</h1>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed imperdiet, lacus id congue molestie, ipsum nibh aliquet erat, a eleifend justo quam a nisi. Nunc molestie rutrum vulputate. Sed vehicula rhoncus velit, nec lacinia nibh tristique eu. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Sed malesuada est et arcu lobortis viverra. Sed facilisis nulla ut diam adipiscing a dictum nunc feugiat. Proin vestibulum, sem sed tincidunt mattis, nulla massa aliquam enim, ac laoreet ipsum massa quis dui. Cras id metus nisi. Praesent congue blandit sapien et lacinia. Suspendisse varius risus ac magna pellentesque et euismod turpis laoreet. Morbi et tortor lorem. Etiam ac turpis nibh, ac iaculis diam. </p>

<h2>A level two (2) heading</h2>
<p>Een stukje tekst. Dat lang genoeg is om meerdere regels te beslaan. Als
het het goed is staat de titel hierboven een eindje van de vorige paragraaf
af.</p>

<p>Hierna volgt er een lijstje met wat elementen. Eens kijken wat dit
  allemaal gaat doen.</p>
  
<ul>
  <li>Item1</li>
  <li>Item2</li>
</ul>

<p>En wat doet deze paragraaf? Die hopelijk langer is dan een enkele
  regel zodat de pagina niet zo leeg lijkt.</p>
  
<h2>Kopje met meteen een lijst</h2>
<ul>
  <li>Item1</li>
  <li>Item2</li>
</ul>

<p>En dan hebben we hier een leuke paragraaf die eigenlijk helemaal
   nergens over gaat. Maar slechts hier staat om te laten zien hoe een
   <a href="#">link</a> eruit ziet binnen een tekst.</p>
      
<h2>Lists</h2>
<h3>Two lists after eachother</h3>
<ul>
  <li>List 1</li>
</ul>
<ul>
  <li>List 2</li>
</ul>

<h3>Two ol lists after each other</h3>
<ol>
  <li>List 1</li>
</ol>
<ol>
  <li>List 2</li>
</ol>

<h3>First ol, then ul</h3>

<ol>
  <li>Foo</li>
</ol>
<ul>
  <li>Bar</li>
</ul>

<h3>First ul, then ol</h3>

<ul>
  <li>Bar</li>
</ul>
<ol>
  <li>Foo</li>
</ol>

<h2>Tables</h2>

<table>
  <tr>
    <td>A</td>
    <td>B</td>
  </tr>
  <tr>
    <td>C</td>
    <td>D</td>
  </tr>
</table>

<p>This is the paragraph after a table...</p>
```
