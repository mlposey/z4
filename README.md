# z4
z4 is a distributed database for managing tasks.

## Architecture
The z4 architecture is focused on providing
* Durability
  * Writes are persisted onto a storage medium and then replicated to other
    peers. This ensures data persists across application restarts and even
    if the storage used by a peer fails.
* Availability
  * The system is not highly available, but it can support automated failover
    if configured with a sufficient number of peers.

A key part of achieving the above goals is a reliance on the Raft consensus
algorithm. Raft enables the replication of data as well as the automated
failover when peers die.

## Deployment Model
A collection of z4 application instances forms a cluster. Each member of the cluster is called
a peer.

When choosing the number of peers for a cluster, one must consider quorum needs.
A cluster needs `(N/2)+1` peers to be available to reach quorum. If it cannot
reach quorum, the cluster will be become unavailable. This encourages the following
cluster configurations

|Cluster Size (N)|Tolerated Peer Failures|
|------------|-----------------------|
|3|1|
|5|2|
|7|3|
|..|..|

## Cluster Administration
z4 provides a gRPC service for managing clusters. This repository
ships with [a tool](cmd/z4t) for interacting with that service.
### Example Usage
#### Get cluster info
`z4t -t localhost:6355 info`

The `info` command returns information that the target node has about the overall cluster.

#### Add peers to the cluster
`z4t -t localhost:6355 -p localhost:6456 -id peer1 add-peer`

The `add-peer` command adds a peer to the cluster. The peer address must point to
the peer's raft port rather than the port of the admin gRPC service. The target peer
must be the leader.

#### Remove peers from the cluster
`z4t -t localhost:6355 -id peer1 remove-peer`

The `remove-peer` command removes a node from the cluster. The target peer
must be the leader.
