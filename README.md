# toystore

Toystore is an implementation of the [Dynamo](http://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf) distributed database. Toystore relies on an adapted [SWIM](https://www.cs.cornell.edu/~asdas/research/dsn02-swim.pdf) gossip protocol to keep track of membership. 

![Visual Representation](http://www.charlesetc.com/images/toystore.png)

## Build Status

We have the starting section working right now. Given a seed, additional nodes can join and start receiving keys. There's a branch dedicated to then copying over the keyspace. Lastly, we haven't fully implemented the reaction to a failure. In order to insure that this works correctly, we're working on a sister project called [Teardown](http://github.com/rlayte/teardown) which will be a testing library for distributed systems.

## SWIM

SWIM is a gossip protocol that aims to reduce network traffic while still moving membership data around relatively quickly. The SWIM protocol has two main features: It piggybacks new joins and failures across the heartbeat ping, instead of flooding the system or a similar distribution strategy. Secondly, when a node (A) in SWIM can't communicate with another node (B), it will ask a third node (C) to ping (B). This makes SWIM very good at tolerating partial network partitions. 

For Toystore, we implemented a SWIM-like protocol at [charlesetc/dive](http://github.com/Charlesetc/dive). Dive uses the first piggybacking strategy that SWIM suggests, but doesn't implement the probing strategy because Dynamo includes its own strategy for dealing with these errors. 

## Differences

-- Data Versioning
-- Permanent Failures
-- Virtual Nodes

Dynamo works on top of the membership established by SWIM. Dynamo relies on an internal structure of a ring of members. The members are placed in the ring by a consistent hash of a unique identifier. Likewise, each incoming key is placed in the ring by the consistent hash and is therefore assigned to a node. 

Each dynamo node exposes two methods to a client: Get(string)(string,bool) and Put(string, string)bool. Right now, we're only using strings as values which means marshaling has to be done in an external setting. It's possible we'll include this by default.

When a node is asked to either Get or Put a key, it first checks to see if it's responsible for a key. A node is responsible for a key if there are no members in its membership ring less than itself and greater than the key. If there is such a member, it will forward the Get request to the node that it thinks is the closest to the key. That node then checks to make sure if it's responsible for the key according to its membership list and proceeds accordingly. This provides amortized constant network traffic, whereas some other protocols jump along until they find the right node.

When another node joins the cluster, it automatically takes on responsibility for its keyspace, and this part of the keyspace is copied over from an adjacent node. When the node that is responsible for a key is no longer available, the node trying to read will look at the next closest node to the key. The data is set to be replicated backwards N times, so the next closest node will always have the data unless there's huge outage across multiple nodes all at once.

In the toystore implementation, data is currently replicated N times so that if a node fails temporarily, there's not a data outage. There's a mechanism for dealing with writes during a failure scenario called Hinted Hand-off that has not yet been implemented.

## Todo

* Finish copying keyspace on join
* Hinted hand-off during failures
* Better testing with teardown.

## Run the server

    $ go run api/http.go <port>

Or run a cluster with

    $ supervisord

[rlayte/toystore-admin](http://github.com/rlayte/toystore-admin) serves as some example code for using the Toystore library.

## Testing

The messaging protocol Dive is very well tested. It's substantially harder to test a larger system like Toystore. We've been testing it manually so far, but we're working on a testing library, [Teardown](http://github.com/rlayte/teardown) to really ensure reliability. 
