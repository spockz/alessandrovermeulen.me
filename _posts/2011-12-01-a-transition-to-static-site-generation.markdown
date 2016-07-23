---
layout: post
title: "A transition to static site generation"
date: 2011-12-14 22:00
comments: true
categories: 
published: true
---
Today I've launched my new blog. It is based on [Octopress](http://octopress.org/docs)
and works by statically generating the pages and then syncing them with the
server.

If you are for example on OS X Lion and installed XCode 4.2 and you run into
weird errors like a missing gcc-4.2, and Homebrew throws errors like this:

> Error: The linking step did not complete successfully
> The formula built, but is not symlinked into /usr/local

Please install the gcc package from this nice fellow:
[osx-gcc-installer](https://github.com/kennethreitz/osx-gcc-installer)

And if you are getting nagged by `rb-fsevent`. Change

```ruby Gemfile
  gem 'rb-fsevent'
```

to 

```ruby Gemfile
  gem 'rb-fsevent', "0.9.0.pre4"
```

**Update**
The comments have been exported with the Wordpress plugin to Disqus. I'm
currently looking at how to highlight code within Disqus comments.