---
# WHY: Preserve this archived article's metadata and exact historical URL.
{
  "title": "How To setup AFP on Debian Lenny",
  "date": "2010-09-10T00:00:00+02:00",
  "slug": "how-to-setup-afp-on-debian-lenny",
  "draft": false,
  "tags": [],
  "url": "/2010/09/10/how-to-setup-afp-on-debian-lenny/",
  "aliases": [
    "/posts/how-to-setup-afp-on-debian-lenny/"
  ]
}
---

After some time of struggling with MacFuse (+MacFusion) and SSHFS I set out to get something that works nice, is integrated in OS X nicely and above all, is not sslooow.
AFP which is all of this is available for Debian but without support of the new encrypted password mechanism.

To this extent I searched and found a nice tutorial: <a href="http://routerjockey.com/2009/08/28/setting-up-apple-filing-protocol-and-bonjour-under-debian/">Setting up Apple Filing Protocol and Bonjour under Debian</a>.

Apparently is was there when I tried this the last time and I missed it.
But now it works.
This works seamlessly with OS X 10.6.
I now even have a nice heavy duty server icon in the Finder for my Debian machine.
This is because of the avahi service.

This how to also explains how to setup your TimeMachine to backup to the Debian server.

<p><small>Archived image unavailable: Screenshot of my Debian machine in the Finder.</small></p>
