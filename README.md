# toystore

Toystore is an implementation of the [Dynamo](http://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf) distributed database, which relies on an adapted [SWIM](https://www.cs.cornell.edu/~asdas/research/dsn02-swim.pdf) gossip protocol to keep track of cluster membership. 

[Full API documentation](https://godoc.org/github.com/rlayte/toystore)

### Differences from the paper

The purpose of this project was primarily educational so we took some short cuts where it made sense.

#### Data Versioning

Dynamo uses vector clocks to handle data conflicts between nodes. This has the advantage of allowing the client to decide how to merge these conflicts in the case of a partition, but we opted for the more straight-forward last-write-wins approach. This has drawbacks, but because we're testing with simple key/value pairs rather than updating fields within encoded data we decided it was good enough. 

#### Permanent Failures

Dynamo classifies two types of failures: transient and permanent. Transisent failures are the most common and can be cause by network partitions, node crashes, and slow running processes. We're handling these temporary failures using the hinted handoff approach described in the paper. However, Dynamo handles failures that have persisted for a longer time (e.g. more than 24 hours) differently by removing them from the cluster and rebalancing the key range. We decided not to implement this for now.

#### Virtual Nodes

Dynamo has the concept of virtual nodes, which allow a single physical host to store multiple key ranges. This has two benefits: spread the nodes over the hash ring more effectively and allow for non-homogenous hosts (e.g. some servers can store more keys than others). As we're testing this locally and therefor don't have different hardware to consider we decided to not implement this feature.

## Setup

We assume you have loopback addresses on `127.0.0.2:127.0.0.24`. If you're running OSX this won't be the case so you'll need add these addresses or use a VM.

### Run the server

We've provided an example app in the api package. Run it with:

    $ # Start the seed node
    $ go run api/http.go 127.0.0.2
    $ # Start other nodes
    $ go run api/http.go 127.0.0.{n}

### Testing

    $ go test

## Admin

If you prefer to use a browser based tool you can run an example admin interface using [rlayte/toystore-admin](https://github.com/rlayte/toystore-admin)

![Visual Representation](http://www.charlesetc.com/images/toystore.png)
