--- 
layout: post
title: How To setup AFP on Debian Lenny
tags: []

comments: true
---
After some time of struggling with MacFuse (+MacFusion) and SSHFS I set out to get something that works nice, is integrated in OS X nicely and above all, is not sslooow. AFP which is all of this is available for Debian but without support of the new encrypted password mechanism.

To this extent I searched and found a nice tutorial: <a href="http://routerjockey.com/2009/08/28/setting-up-apple-filing-protocol-and-bonjour-under-debian/">Setting up Apple Filing Protocol and Bonjour under Debian</a>.

Apparently is was there when I tried this the last time and I missed it. But now it works. This works seamlessly with OS X 10.6. I now even have a nice heavy duty server icon in the Finder for my Debian machine. This is because of the avahi service.

This how to also explains how to setup your TimeMachine to backup to the Debian server.

<a href="http://alessandrovermeulen.me/wp-content/2010/09/Screen-shot-2010-09-10-at-12.24.28.png"><img class="alignnone size-medium wp-image-265" title="Screenshot of my Debian machine in the Finder" src="http://alessandrovermeulen.me/wp-content/2010/09/Screen-shot-2010-09-10-at-12.24.28-300x225.png" alt="Screenshot of my Debian machine in the Finder" width="300" height="225" /></a>
