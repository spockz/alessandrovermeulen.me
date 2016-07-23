--- 
layout: post
title: Caching hackage
tags: 
- Haskell
- hackage
comments: true
---
{{:toc}}

On several occasions I noticed that when performing a `cabal update` that the
index was being downloaded at the rate of plus min 300 KB/s. Finally I got around to do
something about this. I've set up a caching server located in Utrecht, The
Netherlands. It is a caching proxy for the hackage repository. If you want to
use it, add the following to your `~/.cabal/config` file. (Or equivalent on
Windows.)

``` bash
remote-repo: cache:http://spockz.nl:12080/packages/archive
```

Be sure to comment out the already existing remote-repo. Otherwise, cabal will
download both indexes and merge them, and we don't want this.

## The funny bit
Apparently this only helps if your machine is fast enough to process the index
(untarring and all extra administration cabal performs).

Plainly getting the file from the cache:

``` bash
> wget http://spockz.nl:12080/packages/archive/00-index.tar.gz
Saving to: 00-index.tar.gz
100%[======================================>] 3.394.821   10,9M/s   in 0,3s
2011-11-13 00:12:35 (10,9 MB/s) - 00-index.tar.gz saved [3394821/3394821]
```

And running cabal update with my cache as source:

``` bash
cabal update -v3  7,50s user 0,21s system 99% cpu 7,736 total
```

And then finally, with the original repository:

``` bash
cabal update -v3  7,57s user 0,25s system 28% cpu 27,372 total
```

So here we see that the user time is roughly the same but you spent almost three
times more seconds waiting for your coffee to get cold. Any further speed
improvements for cabal update will probably require optimalisation of the code.

## The caching server
I'm using <a href="https://www.varnish-cache.org/" title="Varnish">Varnish</a>
to cache the request to hackage. And here is my config file. Please shoot if you
see any improvements.

```

# This is a basic VCL configuration file for varnish.  See the vcl(7)
# man page for details on VCL syntax and semantics.

backend backend_0 {
	.host = "hackage.haskell.org";
	.port = "80";
	.connect_timeout = 0.4s;
	.first_byte_timeout = 300s;
	.between_bytes_timeout = 60s;
}


acl purge {
  "localhost";
	"hackage.haskell.org";
}

sub vcl_recv {
    set req.grace = 120s;
    set req.backend = backend_0;

		set req.http.host = "hackage.haskell.org";

    if (req.request == "PURGE") {
        if (!client.ip ~ purge) {
            error 405 "Not allowed.";
        }
        return(lookup);
    }

    if (req.request != "GET" &amp;&amp;
        req.request != "HEAD" &amp;&amp;
        req.request != "PUT" &amp;&amp;
        req.request != "POST" &amp;&amp;
        req.request != "TRACE" &amp;&amp;
        req.request != "OPTIONS" &amp;&amp;
        req.request != "DELETE") {
        /* Non-RFC2616 or CONNECT which is weird. */
        return(pipe);
    }

    if (req.request != "GET" &amp;&amp; req.request != "HEAD") {
        /* We only deal with GET and HEAD by default */
        return(pass);
    }

    if (req.http.If-None-Match) {
        return(pass);
    }

    if (req.url ~ "createObject") {
        return(pass);
    }

    remove req.http.Accept-Encoding;

    return(lookup);
}

sub vcl_pipe {
    # This is not necessary if you do not do any request rewriting.
    set req.http.connection = "close";

}

sub vcl_hit {

    if (req.request == "PURGE") {
        purge_url(req.url);
        error 200 "Purged";
    }

    if (!obj.cacheable) {
        return(pass);
    }
}

sub vcl_miss {

    if (req.request == "PURGE") {
        error 404 "Not in cache";
    }

}

sub vcl_fetch {
    set beresp.grace = 120s;

    if (!beresp.cacheable) {
        return(pass);
    }
    if (beresp.http.Set-Cookie) {
        return(pass);
    }
    if (beresp.http.Cache-Control ~ "(private|no-cache|no-store)") {
        return(pass);
    }
    if (beresp.http.Authorization &amp;&amp; !beresp.http.Cache-Control ~ "public") {
        return(pass);
    }
    
}

sub vcl_deliver {

}
```
