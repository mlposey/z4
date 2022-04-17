# z4
z4 is a distributed database for managing tasks.

> This project is in an experimental state and not yet ready for production.

## Contents
* [Architecture](#architecture)
* [Deployment Model](#deployment-model)
* [Running Locally with Docker Compose](#running-locally-with-docker-compose)
* [Cluster Administration](#cluster-administration)
* [Configuration](#configuration)

### Architecture
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
failover when peers become unreachable.

### Deployment Model
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

### Running Locally with Docker Compose
A [docker-compose file](docker/docker-compose.yaml) allows you to test a three-node cluster locally.

Run `make compose_up` to build and start the cluster.  
Run `make compose_down` to stop and destroy the cluster.

The Compose environment will initially make container `peer1` the leader. The other peers
can be added to the cluster using the Admin gRPC service directly or through
the [z4t](cmd/z4t) tool. Peers can be stopped and started using docker commands.  

Storage is persisted when restarting individual containers but erased when using the
`make compose_down` command.

#### Example Usage
1. Start the cluster
   ```
   > make compose_up
   ```
2. Build z4t tool
   ```
   > cd cmd/z4t
   > go build
   ```
3. Inspect the cluster
   ```
   > ./z4t -t localhost:6355 info
   {
     "server_id": "peer1",
     "leader_address": "192.168.112.4:6356",
     "members": [
       {
         "id": "peer1",
         "address": "peer1:6356"
       }
     ]
   }
   ```
4. Connect cluster members
   ```
   > ./z4t -t localhost:6355 -p peer2:6356 -id peer2 add-peer
   peer added
   > ./z4t -t localhost:6355 -p peer3:6356 -id peer3 add-peer
   peer added
   ```
5. Inspect the cluster
   ```
   > ./z4t -t localhost:6355 info
   {
     "server_id": "peer1",
     "leader_address": "192.168.112.4:6356",
     "members": [
       {
         "id": "peer1",
         "address": "peer1:6356"
       },
       {
         "id": "peer2",
         "address": "peer2:6356"
       },
       {
         "id": "peer3",
         "address": "peer3:6356"
       }
     ]
   }
   ```
6. Add some tasks using the Collection gRPC service  
   Requests *should* be sent to the leader, but any follower with access to the leader
   can accept requests and forward them to the leader.
7. Terminate the leader
   ```
   > cd ../..
   > docker-compose -f docker/docker-compose.yaml stop peer1
   ```
8. Add more tasks using either peer2 or peer3
9. Restart the old leader (and observe that they become a follower this time)
   ```
   > docker-compose -f docker/docker-compose.yaml start peer1
   ```
10. Add more tasks using any peer
11. Tear down the cluster
   ```
   > make compose_down
   ```
### Cluster Administration
z4 provides a gRPC service for managing clusters. This repository
ships with [a tool](cmd/z4t) for interacting with that service.
#### Example Usage
##### Get cluster info
`z4t -t localhost:6355 info`

The `info` command returns information that the target node has about the overall cluster.

##### Add peers to the cluster
`z4t -t localhost:6355 -p localhost:6456 -id peer1 add-peer`

The `add-peer` command adds a peer to the cluster. The peer address must point to
the peer's raft port rather than the port of the admin gRPC service. The address
pointed to by the `t` flag should be that of the cluster leader.

##### Remove peers from the cluster
`z4t -t localhost:6355 -id peer1 remove-peer`

The `remove-peer` command removes a node from the cluster. The address
pointed to by the `t` flag should be that of the cluster leader.

### Configuration
#### Environment Variables
|Variable|Description|Default|
|--------|-----------|-------|
|Z4_DEBUG_LOGGING_ENABLED|Enables or disables debug-level logging|false|
|Z4_DB_DATA_DIR|The directory where task data is stored|z4data|
|Z4_PEER_DATA_DIR|The directory where cluster membership data is stored|z4peer|
|Z4_SERVICE_PORT|The port containing the gRPC services|6355|
|Z4_PEER_PORT|The port containing the internal cluster membership service|6356|
|Z4_METRICS_PORT|The port containing the Prometheus metrics service|2112|
|Z4_PEER_ID|The unique ID of the cluster member. Must be stable across restarts||
|Z4_PEER_ADVERTISE_ADDR|The host:port of the peer that other members use for internal operations|127.0.0.1:6356|
|Z4_BOOTSTRAP_CLUSTER|Determines whether the peer should declare itself the leader to kickstart the cluster|false|
