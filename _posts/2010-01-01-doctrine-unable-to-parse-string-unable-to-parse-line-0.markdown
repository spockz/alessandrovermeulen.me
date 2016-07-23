--- 
layout: post
title: "Doctrine: Unable to parse string: Unable to parse line 0 ("
tags: 
- doctrine
comments: true
---
Normally this error would mean that you have some weird syntax error somewhere in your yml files. However, if you are sure everything is up to spec be sure to check whether there are any `hidden' files in the folder. Such files start with a dot and are also read and parsed by the command line generator. 

This could potentially fail, in case you get the aforementioned error. Be sure to remove these files before proceeding. To check whether there are any, go to the directory and retrieve it's listing.

[bash]
cd /some/path/to/your/yml/files
ls -lha
[/bash]

If you see any files starting with a dot and ending in .yml be sure to check them for any valuable content and throw them away.

Thanks to <a href="http://blog.rajatpandit.com/2009/08/23/doctrine_parser_exception-fail-some-error-message/">rajat pandit</a> for hinting to this solution.
