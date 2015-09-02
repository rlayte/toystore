# toystore

Toystore is an implementation of the [Dynamo](http://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf) distributed database. Toystore relies on an adapted [SWIM](https://www.cs.cornell.edu/~asdas/research/dsn02-swim.pdf) gossip protocol to keep track of membership. 

![Visual Representation](http://www.charlesetc.com/images/toystore.png)

## Build Status

Nodes can join an existing cluster and start saving and distributing keys. When a node joins, data is transferred over to them. In order to insure that this works correctly, we're working on a sister project called [Teardown](http://github.com/rlayte/teardown) which will be a testing library for distributed systems. Dynamo's version of hinted-handoff which deals with temporarily failing nodes is still being worked on.

## Overview

Dynamo works on top of the membership established by SWIM. Dynamo relies on an internal structure of a ring of members. The members are placed in the ring by a consistent hash of a unique identifier. Likewise, each incoming key is placed in the ring by the consistent hash and is therefore assigned to a node. 

Each dynamo node exposes two methods to a client: Get(string)(string,bool) and Put(string, string)bool. Right now, we're only using strings as values which means marshaling has to be done in an external setting. It's possible we'll include this by default.


### SWIM

SWIM is a gossip protocol that aims to reduce network traffic while still moving membership data around relatively quickly. The SWIM protocol has two main features: It piggybacks new joins and failures across the heartbeat ping, instead of flooding the system or a similar distribution strategy. Secondly, when a node (A) in SWIM can't communicate with another node (B), it will ask a third node (C) to ping (B). This makes SWIM very good at tolerating partial network partitions. 

For Toystore, we implemented a SWIM-like protocol at [charlesetc/dive](http://github.com/Charlesetc/dive). Dive uses the first piggybacking strategy that SWIM suggests, but doesn't implement the probing strategy because Dynamo includes its own strategy for dealing with these errors. 

## Differences

### Data Versioning

Dynamo calls for Data Versioning, which becomes important when sending data across during a join. To accomplish this, vector clocks are used to keep track of relative time. This makes a mostly reliable solution, which the Dynamo paper argues is better than timestamps because there cannot be a guarantee on the accuracy of clocks between machines. The only problem with vector clocks is when there is a conflict between data at the same version level, which must be passed to a client for resolution.

We have not yet implemented these, but it might be worthwhile in the future.

### Permanent Failures

If a node permanently fails, there should be a rebalancing of the keyspace so that nearby nodes aren't taking the failed node's load for no reason. This doesn't happen with Toystore yet.

###  Virtual Nodes

Dynamo also calls for Virtual nodes to be used instead of a single one per peer. This means that peers can be configured to have a larger load or a smaller load depending on the actual hardware of the system. While useful, it's not strictly necessary for the Dynamo protocol so we have yet to make an implementation.

## Todo

* Hinted hand-off during failures
* Better testing with teardown.

## Run the server

    $ go run api/http.go <port>

Or run a cluster with

    $ supervisord

[rlayte/toystore-admin](http://github.com/rlayte/toystore-admin) serves as some example code for using the Toystore library.

## Testing

The messaging protocol Dive is very well tested. It's a little harder to test a larger system like Toystore. We're trying to make something generalizable with [Teardown](http://github.com/rlayte/teardown) to really ensure reliability, but right now running `go test` will launch a suite of integration and unit tests.
