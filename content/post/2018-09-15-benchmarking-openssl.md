+++
categories = ["programming", "benchmarks"]
date = "2018-09-15T14:33:00+02:00"
description = "Maturity of TLS in Java "
tags = ["openssl", "tomcat", "finagle", "netty"]
title = "Experiences of troubleshooting and benchmarking TLS in Java using JSSE and OpenSSL – Part I"
draft = true
+++

## Preamble

Traditionally, TLS has been employed at TLS offloaders at the edge of network boundaries partly because of the high computational cost.
With the advent of an overall increase in concern for security – partly due to the advent of APT (advanced persistent threats), because TLS is just easier to integrate into your average day application, and because of reduced performance impact, TLS is now also steadily being employed more *within* network boundaries.
This means, amongst other things, that TLS is being added to microservices. The fact that it is relatively easily possible to add TLS does not mean that it is also easy to *do it right*.

One of the reasons to why it is hard to do TLS *right* is because there are so many knobs to turn and considerations to make with relation to security but the exact effects, in terms of runtime performance and security, of turning a knob is not known. The two parameters that probably best known are the size of the keys and which protocol to talk. The first because it is mentioned in pretty much every tutorial, the later because everyone encounters this moment where they specifically need to set a TLS version to talk because it doesn't work otherwise.

In this series I'll be posting about my journey to learn more about TLS and how it fits optimally in Java microservices. This first post starts with describing the setup of a benchmark suite I created and the analysis of the data.

### Why

The trigger for writing this benchmark is two-fold. Firstly, our software was showing unexpected behaviour in different environments, and secondly, I couldn't stomach that when things went wrong we were in the dark. A lot of effort goes into making TLS implementations secure and performant. I am thankful for these efforts. What I also observed is that observability,  asfds fdOne of the things I've observed is that a lot of focus is put on getting a performant solution, but when there is very little observability for when things go wrong. E.g. when CPU usage jumps up there can be several reasons:

1. We are, legitimally, receiving more load
1. Somebody is, accidentally, connecting to us with wrong settings
    * Mismatching protocol or cipher suites
    * Expired client certificates
    * Missing session resumption
    * ...
1. Maliciously trying to get our servers to overload

Matters are aggrevated because in a microservices environment everything is updating continously, including Java runtime environments. One of those upgrades (8u161) [enabled extended master secret](https://bugs.openjdk.java.net/browse/JDK-8192045) causing session resumption to fail with OpenSSL 1.0 enabled systems.

One of the reasons this could have happened is that for far to long TLS has been considered and used as a bolt-on by application developers — perhaps limited to the Java ecosystem as Go seems to have its own tls stack. Or maybe this just happened because before DevOps operations was just left to ops. Similarly, it appears as if for a large part the intent of most TLS implementations was to abstract any details other than the secret and the session id away from users(TODO, add links). With recent attacks and the introduction of TLSv1.3 this appears to have changed a bit(TODO: Insert note to tls1.3 application level considerenations and to the mitigation which requires application level mitigation).

As more and more applications are using (mutual-TLS) I hope libraries providing TLS support will mature with regards to observability as well.

## Benchmarking Tomcat

As one of the first steps in understanding more about the performance of our application I've created a benchmark suite to measure


## Further reading

* TLSv1.3
    * OpenSSL released with TLSv1.3 support
* TLSv1.2
