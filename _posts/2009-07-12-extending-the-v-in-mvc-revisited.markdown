--- 
layout: post
title: Extending the V in MVC - revisited
tags: 
- mvc
- codeigniter
- partials
- partial
- rss
- php
- layout
comments: true
---
This post is an revised version of 'Extending the V in MVC on the web' of November 9, 2008.

<hr />

When I first started working with the MVC method it was with my own framework.
After noticing that maintaining my own framework would be very time-comsuming I
embarked on the search for a free, fast and widely supported PHP web framework.
The one I found that time was '<a
href="http://www.codeigniter.com">CodeIgniter</a>'.

After working on different projects I noticed that most websites present the
same data to the visitor in several different formats. Think of RSS/Atom for
blog posts for use in an RSS reader together with the (x)html versions.

While essentially providing the same data to the visitor the programmer of the
website has to put sizable amount of work to add little extra functionality in
the form of an RSS feed. Repeating a lod of code for input validation, data
retrieval from the database and creating extra routes into the application. As
you can imagine this could grow into a tedious job.

I'll be describing techniques known and in use today, but reinvented each time
again as a (new) developer starts working on an web application. Here we'll be
giving them name and be describing them.

<!--more-->
<h3>Theory</h3>
What are layouts? What are partials? The answers to these questions will be
discussed in the following sections.
<h4>Layouts</h4>
What is a layout? We will be using the following definition of the word layout:
"A layout is the representation of data in a specific format". So the layout of
a page can be either, xhtml, RSS, PDF, PPT or any other data format.

Suppose you would find yourself in the situation where you need to write a blog.
You could start with building the models that describe users, posts, comments
and maybe even backtracks. Then you start writing the controllers and views and
let someone else do the design for you. :)

So, the designer has made some nice designs and you are converting those to
(x)html, CSS and images for use in your views. In the same time you write your
controllers, one for every aspect of your blog. Consider the controller that
shows a list of your most recent posts.

Blogs are essentially meant for publishing information, the messages the poster
posts, to the world. People want to see a nice readable web page when they look
at your blog. But it is hard to keep track of all the blogs one visits for new
posts. That's why people invented the syndicates. And look now, here you are,
having two URLs providing the same information, the latest posts, and the only
difference is that the information is that they present the data in another
format. So why should the URLs be different? - Other than the extension that
denotes in which format the page should be presented? - Right, they shouldn't.
Nor should you have to write any administrative code. The only code you should
have to write is the code responsible for presenting the given data in a
specific layout.

With just a small amount of initial trouble, which we'll see in the next
chapter, it will save you a lot of time and administrative rut. Sounds good,
doesn't it?

<h4>Partials</h4>
Another aspect of this article is 'partials'. Programming today is all about
re-using code and rapid development and low-maintenance. OOP gives you the
ability to re-use your previously written code you worked so hard on in a
sensible matter. It also can help in reducing the draught of maintenance. The
MVC paradigm is extending this and helps by further aiding rapid development.

In the aspect of re-using code, partials are very important. Where OOP , Models
and Controllers can save you time by re-using program code, partials can help
you save time by offering a new way to re-use your html (or any other layout)
code. Thus not only decreasing the maintenance needed for your application but
simultaneously increasing consistency throughout your application without
causing you any headaches.

Where does the name 'partial' come from? The origins of this name are quite
simple. A 'partial' is a small re-usable snippet of View code and is a View
itself.

Partials can be used for many things. For example, the header on all your blog
pages could be the same, apart from variables as titles and needed stylesheets
and javascript includes. As we have learned during the age of the static
webpages, maintaining this would be an hell of a pain. So let's not do that
anymore and let's use partials instead. Keep the code in one place and just use
the partial if you need the header. Do no retype or copy any code.

Of course, headers and footers are not the only things that we can store in
partials. The menu, the login form, the code for a blog post and the lot can be
placed in partials. Actually almost anything that will be used more than once
should be placed in a partial.

<h3>An implementation in CodeIgniter</h3>
This part shows one way to implement layouts and partials in CodeIgniter. We
assume you have at least version 1.6.3 of CodeIgniter and that you have a fair
idea of what we are going to do. Proceeding is completely at your own expense
and I am not accountable for any issues that may rise do to these adjustments

<h4>Adjustments</h4>
<h5>The Controller class</h5>
The first thing we have to do is extending the main controller class provided by
CodeIgniter. This enables us to do whatever we want without changing anything in
the CodeIgniter code base. Place this function in the extending controller:

[file lang="php"]snippets/extendingthevinmvc/controller.php[/file]
<h5>The loader</h5>
Now put these functions inside class extending the CI Loader class:

[file lang="php"]snippets/extendingthevinmvc/loader.php[/file]
<h4>Usage</h4>
<h5>Views</h5>
Now that we have extended the functionality of CodeIgniter with our own we can
start to use layouts and partials. Consider the following url:

<blockquote>http://someurl/blog/category/5</blockquote>

It will give you a nice and cheesy overview formatted in xhtml. So let's take a
look at the view file.

[file lang="php" preserveformat="on"]snippets/extendingthevinmvc/short/xhtml_blog_category.php[/file]

And the code for the short summary looks like this:

[file lang="php"]snippets/extendingthevinmvc/short/xhtml_blog__shortsummary.php[/file]

Now consider this url:
<blockquote>http://someurl/blog/category.rss/5</blockquote>
This url would give you the overview of the posts in category 5. Just as with
the url above. But this time the information will be formatted in RSS. Notice
that the only thing we had to do was to change the url. Create these two files:

<strong>APPPATH/views/rss/blog/summary.php</strong>

[file lang="php"]snippets/extendingthevinmvc/short/rss_blog_summary.php[/file]

<strong>APPPATH/views/rss/blog/_shortsummary.php</strong>

[file lang="php"]snippets/extendingthevinmvc/short/rss_blog__shortsummary.php[/file]

Wow. Now we only had to write a few lines (15) for adding RSS feeds to our site.
This is really all there is to it. Now go out there and have fun.
