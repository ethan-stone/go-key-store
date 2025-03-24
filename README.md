# Overview

This repo is an attempt to many of the concepts related to distributed systems.

# Todo

- [x] Super basic key value store server.
- [x] Ping between nodes in a cluster via manual configuration.
- [x] Organize code a bit.
- [x] Assign hash slots to nodes manually for now.
- [x] Route requests to correct node for hash slot.
- [x] Consider making a "KeyValueStoreFactory" that can get either a "LocalKeyValueStore" or a "RemoteKeyValueStore". Local means the data is stored on this node, while Remote means it's stored on a different node. \*Update definitely need to do this. The local store needs to be decoupled from RPC, because the RPC server side is going to need to reference the local store. Otherwise we'd have a circular module dependency.
- [ ] Gossip-based membership and health check.
  - [x] Update to use a config file per node. For now each config file should have all other nodes.
  - [x] Add seed nodes to config file of nodes.
  - [x] Update seed node config files. To have knowledge of just other seed nodes.
  - [ ] Implement gossip with seed nodes. When a new node starts, it reaches out to a random seed node. The seed node adds the new node to it's membership list, and returns the membership list with it's configuration to the new node. Now the new node has knowledge of all other nodes.
- [ ] Automatic assigning of hash slots.
- [ ] How to gracefully handle nodes going down?
- [ ] Better error handling for internal errors vs. a key just not being found. Right now any error is handled as a not found in the http api.

# Building GRPC Code

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/rpc/node_rpc.proto
```

# Raft

- Website: https://raft.github.io/
- Paper: https://raft.github.io/raft.pdf
- Animation: https://thesecretlivesofdata.com/raft/
