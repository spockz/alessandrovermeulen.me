--- 
layout: post
title: Setting up multiple applications with CodeIgniter
tags: 
- codeigniter
comments: true
---
In addition to my previous post where I discussed the <a title="Setting up
CodeIgniter Basics"
href="http://alessandrovermeulen.me/2009/05/22/setting-up-codeigniter/">initial
setup of your application</a> with CI (CodeIgniter), I'll use this post to
provide a method of setting up your environment for multiple applications.
<!--more-->
Now, in the previous post I was really sloppy about the locations of both the
CodeIgniter system files and the application files. The same guidelines still
apply, both the system folder and application folder should reside outside of
the web document root to ensure safety.

First, we want to store our applications and CodeIgniter releases in separate
folders. So for example store your CodeIgniter system folder in
<em>/var/codeigniter/</em>. This could be done by doing a svn export (or
checkout) with the following command-line statement:

``` bash
$ svn export http://dev.ellislab.com/svn/CodeIgniter/tags/v1.7.1/system/ /var/codeigniter/release-1.7.1
```

Or you can download the <a
href="http://www.codeigniter.com/download_files/CodeIgniter_1.7.1.zip">1.7.1
release</a> from the CodeIgniter website and copy the contents of the system
folder to <em>/var/codeigniter/release-1.7.1</em>.

Now for each application you want to build you can do a export/checkout to our
applications directory. As we are going to store our applications outside the
document root you could pick <em>/var/applications/</em> as root directory for
your CI applications. Now if you want to make a blog you just have to execute
this command:

``` bash
$ svn export http://dev.ellislab.com/svn/CodeIgniter/tags/v1.7.1/system/application/ /var/applications/blog[/bash] or [bash light="true"]$ cp /var/codeigniter/release-1.7.1/application /var/applications/blog
```

I prefer the SVN export as this will ensure that you always have a fresh copy.
Otherwise, should you find yourself without an internet connection you can use
the latter.

Now just as in the previous post, create the directory in the document root
(whether you want to access it through a subfolder or directly through a
sub-domain/domain is entirely up to you and is outside the scope of this post.)

``` bash
$ mkdir /var/www/blog
``` 

Now get the index.php file from svn or from the zip file you might have
downloaded earlier:

``` bash
svn export http://dev.ellislab.com/svn/CodeIgniter/tags/v1.7.1/index.php /var/applications/blog/
```

The only thing left is to alter the index.php so the <em>$system_folder</em> and
the <em>$application_folder</em> point in the right direction:

``` php 
//...
$system_folder = '/var/codeigniter/release-1.7.1';
//...
$application_folder = '/var/applications/blog';
```

Now you can add your web stuff like CSS, JS to your /var/www/blog. Of course
this follows the same principle as creating the non-www and www directories
mentioned in the previous post.

Now for each application you want to add, retake the above steps and replace
'blog' with something relevant. Good steaming!
