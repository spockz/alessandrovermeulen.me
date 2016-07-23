---
layout: post
title: "Orchestration support announced on DockerCon"
date: 2014-12-04 21:46
comments: true
categories: Programming
tags:
  - Docker
---
# Orchestration

The philosophy behind docker is that in order to be solved, a large problem has to be divided into its root problems. One can then proceed by solving every one of these problems step by step. Additionally all elements of the solution need to communicate through a common app.

Docker has always been a tool with a single purpose: the creating, transport, and running of images. Until today there where several issues with docker that make using it somewhat trying at times. It lacked in capabilities for orchestration which is categorized by:

1. Installation of a docker host from scratch;
2. Clustering of multiple docker hosts to spread resource utilization over the cluster;
3. Managing inter-container dependencies at runtime.

Today this changed as Docker Inc. announced a new set of tools.

## Provisioning: Machine

Machine provides a one step installer for creating a new docker host on your local machine, a publicly hosted cloud, or a private cloud. It will automatically provision a new machine and set the environment variables such that any following docker command runs on the newly created host. This is very similar to what boot2docker provides.

There are several engines for provisioning in different platforms such as:

* VirtualBox
* VMWare
* AWS
* Microsoft Hypervisor
* etc. (todo link)

More information can be found at [github](https://github.com/docker/machine).

## Clustering: Swarm

Ideally you want to control a cluster of docker hosts with the same interface as you control a single host. In other words the interface needs to be transparent or standardized. With swarm you can.

All existing commands on docker work with the swarm as well. Just point your docker binary to the swarm proxy and you are controlling the cluster instead of one single machine. Swarm is location/data center aware and also incorporates resource management. The default strategy is to use as little hosts as possible. The strategy places several lighter containers on the same node in order to reserve other nodes for heavier containers.

The main features are:

1. Resource management
2. Scheduling honoring constraints
3. Health checks on the cluster and nodes
4. Supporting the entire docker interface

Additionally Mesos can be used to provide the scheduling. Docker Inc. also announced that Mesos will be a first class citizen in Docker. The goal is to be able to run docker containers along side other Mesos jobs in the Mesos cluster.

More information also on [github](https://github.com/docker/swarm).

It appears that swarm is not supported yet by machine, sadly.

### Managing inter-container dependencies: Composer

Setting up applications that require multiple containers to function correctly is difficult. Keeping them running is even harder. Docker proposes the `Docker Composer`.

Traditionally it ran on one single machine and, until today, orchestration needed to be done manually or through external tools.


# Docker Hub

Docker Inc. also announces an enterprise version of the Docker Hub. It is able to run wherever the enterprise needs it to run and comes with safe 1-click upgrades. Enterprises are adopting containers as development is up to 30 times faster with halve the error rate.

Some fun facts:

* 100000 contributors to docker hub
* 157 TB of data transmitted each month
* 50 TB of data stored

The timeline for 2015:

1. Increase performance of pulls
2. Increase transparancy by adding and improving on status pages
3. Engage in partnership with Microsoft. Most notably this will result being able to run Linux on Microsoft Azure.

