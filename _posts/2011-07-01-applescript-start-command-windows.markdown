--- 
layout: post
title: Applescript script to emulate the behaviour of the `start' command in the windows CMD prompt
tags: 
- os x
- applescript
comments: true
---
People have asked me how to launch a program through the terminal in the
background but with it's own terminal screen similar to the way the "start"
command in the Windows CMD prompt works. Well here it is. Add the following two
files (start.applescript) and start to your path, and make them executable
(`chmod +x {start.applescript,start}`).

One point of notice is to <strong>always</strong> encapsulate strings with
spaces that should be one parameter in quotes.

{% gist 1059096 %}