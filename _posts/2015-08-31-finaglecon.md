---
layout: post
title: "FinagleCon"
date: 2015-08-31 17:57
comments: true
categories:
  - Programming
  - Software
tags:
  - Scala
  - Finagle
---

FinagleCon was held at TwitterHQ in San Francisco. It is refreshing to see a
nice working atmosphere with free food and drinks. Now for the contents.

Twitter's RPC framework, Finagle, has been in production since [August
2010](http://vkostyukov.ru/slides/finch-101/?full#5) and has over 140
contributors. In addition to Twitter, it has been adopted by many large
companies such as SoundCloud. Initially written in Java with FP constructs
(monads, maps, etc.) all over, it was soon after rewritten in Scala.

Finagle is based on three core concepts: Simplicity, Composability, and
Separation of Concerns. These concepts are shown through three primitive
building blocks:
[`Future`](https://twitter.github.io/util/docs/index.html#com.twitter.util.Future),
[`Service`](https://twitter.github.io/finagle/docs/index.html#com.twitter.finagle.Service),
and
[`Filter`](https://twitter.github.io/finagle/docs/index.html#com.twitter.finagle.Filter).

* `Future`s provide an easy interface to create asynchronous computation and
  to model sequential or asynchronous data-flows.
* `Service`s are functions that return futures, used to abstract away, possibly
  remote, service calls.
* `Filter`s are essentially decorators and are meant to contain modular blocks
  of re-usable, non-business logic. Example usages are [LoggingFilter](https://twitter.github.io/finagle/docs/index.html#com.twitter.finagle.filter.LoggingFilter) and
  [RetryingFilter](https://twitter.github.io/finagle/docs/index.html#com.twitter.finagle.service.RetryingFilter).

The use of Futures makes it easy to test asynchronous computations. Services and
filters both can be created separately, each containing their specialized logic.
This modularity makes it easy to test and reason about them separately. Services
and filters are easily composed, just like functions do, which makes it
convenient to test chains. Services and filters are meant to separate behaviour
from domain logic.

As amazing as Finagle is, there are some things one should be aware of. To
create a really resilient application with Finagle one has to be an expert in
its internals. Many configuration parameters influence each other, e.g. queue
size and time-outs. With a properly tuned setup Finagle is properly fast and
resilient (the defaults are good as well, mind you). As most data centres are
heterogenous in their setup, faster machines are added to the pool, and other
conditions change, one has to keep attention to the tuning continuously in order
to maintain optimal performance.

Some general advice, watch out for traffic amplification due to retries, keep
your timeouts low so retry is useful, but not as low that you introduce spurious
timeouts.

For extra points, keep hammering your application until it breaks, find out why
it breaks, fix it, and repeat.

## The future

In addition to this heads up we were also given a nice insight in the upcoming
things for Finagle.

In order to make more informed decision, we will get a new Failure type which
contains more information instead of 'just' a `Throwable`. In this new
`Failure`, an added field indicates whether it is safe to `retry`.

There are several issues with the current way of fine-tuning Finagle, as
mentioned, you need to be an expert to use all the configuration parameters
properly. Next to this the configuration is static and doesn't take into account
changing environments and behaviour of downstream services. Because the tuning
of the parameters is tightly coupled with the implementation of Finagle it is
also hard to change the implementation significantly without significant
re-tuning.

In order to battle the last two points, Finagle will introduce Service Level
Objectives (SLO). The SLO is a higher-level goal that Finagle should strive to
reach instead of low-level hardcoded parameters. What these SLO will be exactly
is not yet known.

## The community

The Finagle team will synchronize the internal Finagle repository with the
Github repository every Monday. They will strive to publish a snapshot version
of the change as well.

For someone looking to write his own protocol to connect to his service,
`finagle-serial` is a nice project to start with. It is small enough to grasp
within a day but big enough to be non-trivial.

It was found that the
[ParGCCardsPerStrideChunk](http://blog.ragozin.info/2012/03/secret-hotspot-option-improving-gc.html)
garbage collection option, available from 7u40, can halve GC times on large
heaps. It is recommended to try this parameter. Tuning seems to be hard to do
and is generally done by copying a 'known good set' of parameters.

[Scrooge](http://twitter.github.io/scrooge/) is a good utility to use for Thrift
and Scala as it is aware of Scala features such as Traits and Objects and can
generate relevant transformations for them.

When you want to connect to multiple data-centres from a single data-centre one
can use
[LatencyCompensation](https://twitter.github.io/finagle/docs/index.html#com.twitter.finagle.client.LatencyCompensation$)
to include latency times.
