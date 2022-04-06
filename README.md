# z4
z4 is a distributed database for managing tasks.

## Cluster Administration
z4 instances provide a gRPC service for managing the cluster. This repository
ships with [a tool](cmd/z4t) for interacting with that service.
### Example Usage
#### Get cluster info
`z4t -t localhost:6355 info`

The `info` command returns information that the target node has about the overall cluster.

#### Add peers to the cluster
`z4t -t localhost:6355 -p localhost:6456 -id peer1 add-peer`

The `add-peer` command adds a node to the cluster. The peer address must point to
the peer's raft port rather than the port of the admin gRPC service. The target node
must be the leader.

#### Remove peers from the cluster
`z4t -t localhost:6355 -id peer1 remove-peer`

The `remove-peer` command removes a node from the cluster. The target node
must be the leader.
