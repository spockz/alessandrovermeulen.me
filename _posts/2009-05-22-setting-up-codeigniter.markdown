--- 
layout: post
title: Setting up CodeIgniter - Basics
tags: 
- codeigniter
comments: true
---
In this post I'll show how to set up CodeIgniter in a way that your code and configuration (passwords!) are safe. It will involve moving the "<em>system</em>" and "<em>application</em>" outside the (public) document root.

Seperating both `<em>system</em>' and `<em>application</em>' has obvious advantages for maintainance and for reusibility. Using a seperate `<em>www</em>' directory enables you to publish all your application specific JS/CSS and other public files.

<!--more-->

The first thing we need to do is to create an checkout of the latest CodeIgniter release (currently 1.7.1)
``` bash
$ svn checkout http://dev.ellislab.com/svn/CodeIgniter/tags/v1.7.1/ somepathoutsidedocumentroot
```

now  move the application folder outside the system folder:

``` bash
$ cd  somepathoutsidedocumentroot
$ mv system/application ./application
```
Now move the system and application folder to some folder outside of your public
document root. Let's call this directory `<em>non-www</em>'. The documentroot
directory will be called `<em>www</em>'.

In the root of your checkout you will find an <em>index.php</em>. Copy/move this
file to <em>www</em>. Open it and look for the following lines:

``` php
$system_folder = "system";
$application_folder = "application";
```

Change them to the following:
``` php 
$system_folder = "../non-www/system";
$application  = "../non-www/application";
```

Now you are up and running. Probably you want something like nice url's. For
this you can use the following .htaccess

```
RewriteEngine on
RewriteRule ([^/]*)\/\?([^/?]*)$ $1/$2 [R]
```

```
RewriteCond $1 !^(index\.php|test\.php|export\.php|robots\.txt|css|img|js|favicon\.ico|doc|data|user_guide|js)
RewriteRule ^(.*)$ ./index.php/$1 [L]
```

Now, go and create some new shiny sites.
