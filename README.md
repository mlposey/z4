> This project is in an experimental state and not yet ready for production.

# z4
z4 is a distributed task queue.

Main features
* Flexible APIs
  * Tasks are published and consumed using a high-throughput gRPC service.
  * Tasks can be queried using SQL.
* Durability
  * Tasks are persisted until acknowledged by consumers (at-least-once delivery).
  * Data is replicated across multiple peers in a cluster.
* Multiple delivery modes
  * Tasks can be delivered in the order they are published or scheduled for delivery
    at a specific time.
* Availability
  * Clusters automatically recover when replicas fail.
* Small footprint
  * z4 does not rely on external dependencies like queues or databases to store tasks. It manages
    its own data.
  * Resource consumption is low compared to other open-source queue implementations
    like Apache Kafka. This enables good performance on cheap hardware.

## Contents
* [Architecture](#architecture)
* [Deployment](#deployment)
  * [Running Locally with Docker Compose](#running-locally-with-docker-compose)
  * [Kubernetes Deployment with Helm](#kubernetes-deployment-with-helm)
  * [Cluster Configurations](#cluster-configurations)
* [Cluster Administration](#cluster-administration)
* [APIs](#apis)
  * [gRPC](#grpc)
  * [MySQL](#mysql)
* [Customization](#customization)
  * [Environment Variables](#environment-variables)
* [Metrics](#metrics)

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

### Deployment
#### Running Locally with Docker Compose
A [docker-compose file](deployments/docker/docker-compose.yaml) allows you to test a three-node cluster locally.

Run `make compose_up` to build and start the cluster.  
Run `make compose_down` to stop and destroy the cluster.

The Compose environment will initially make container `peer1` the leader. The other peers
can be added to the cluster using the Admin gRPC service directly or through
the [z4t](cmd/z4t) tool. Peers can be stopped and started using docker commands.  

Storage is persisted when restarting individual containers but erased when using the
`make compose_down` command.

##### Example Usage
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
   > docker-compose -f deployments/docker/docker-compose.yaml stop peer1
   ```
8. Add more tasks using either peer2 or peer3
9. Restart the old leader (and observe that they become a follower this time)
   ```
   > docker-compose -f deployments/docker/docker-compose.yaml start peer1
   ```
10. Add more tasks using any peer
11. Tear down the cluster
    ```
    > make compose_down
    ```
#### Kubernetes Deployment with Helm
A [Helm chart](deployments/charts/z4) is provided for Kubernetes deploys. It will deploy z4 as a
StatefulSet with a headless service.

#### Cluster Configurations
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

### APIs
#### gRPC
A gRPC service is exposed on the default port 6355. It suppports cluster administration as well
as task management.
* [Admin service proto](proto/admin_service.proto)
* [Queue service proto](proto/queue_service.proto)

#### MySQL
A MySQL interface is exposed on the default port 3306. It provides read-only access to queues and tasks.

There are few things to note
* There is currently no support for username and password authorization. When connecting, disable authentication.
* All tables are stored in the database `z4`.
* This interface is not optimized for transactional workloads. It is primarily meant as a tool for troubleshooting issues.
* When querying tasks, queries **require** expressions that filter the queue and date range.  
  This is okay:
  ```sql
  SELECT *
  FROM   tasks
  WHERE  queue = 'welcome_emails'
    AND  deliver_at BETWEEN '2022-04-16' AND '2022-04-17'
    AND  JSON_EXTRACT(metadata, '$.user_id') = 'newuser@example.com';
  ```
  This is not:
  ```sql
  SELECT *
  FROM   tasks
  WHERE  queue = 'welcome_emails'
    AND  JSON_EXTRACT(metadata, '$.user_id') = 'newuser@example.com';
  ```
  And neither is this:
  ```sql
  SELECT *
  FROM   tasks
  WHERE  JSON_EXTRACT(metadata, '$.user_id') = 'newuser@example.com';
  ```

### Customization
#### Environment Variables
|Variable|Description|Default|
|--------|-----------|-------|
|Z4_DEBUG_LOGGING_ENABLED|Enables or disables debug-level logging|false|
|Z4_DATA_DIR|The directory where task data is stored|z4data|
|Z4_SERVICE_PORT|The port containing the gRPC services|6355|
|Z4_PEER_PORT|The port containing the internal cluster membership service|6356|
|Z4_METRICS_PORT|The port containing the Prometheus metrics service|2112|
|Z4_SQL_PORT|The port containing the MySQL-compatible server|3306|
|Z4_PEER_ID|The unique ID of the cluster member. Must be stable across restarts||
|Z4_PEER_ADVERTISE_ADDR|The host:port of the peer that other members use for internal operations|127.0.0.1:6356|
|Z4_BOOTSTRAP_CLUSTER|Determines whether the peer should declare itself the leader to kickstart the cluster|false|

### Metrics
Prometheus metrics are exposed on a configurable port that defaults to 2112.

#### API Metrics
|Name|Type|Description|
|----|----|-----------|
|z4_create_task_request_total|Counter|The total number of requests from clients to create a task|
|z4_streamed_task_total|Counter|The total number of tasks sent to clients|


#### Cluster Metrics
|Name|Type|Description|
|----|----|-----------|
|z4_received_log_count|Counter|The total number of Raft logs sent for application to the fsm|
|z4_is_leader|Gauge|A boolean value that indicates if a peer is the cluster leader|

#### Database Metrics
|Name|Type|Description|
|----|----|-----------|
|z4_last_fsm_snapshot|Gauge|The unix time in seconds when the last fsm snapshot was taken|
