# Overview

This repo is an attempt to many of the concepts related to distributed systems.

# Todo

- [x] Super basic key value store server.
- [x] Ping between nodes in a cluster via manual configuration.
- [x] Organize code a bit.
- [x] Assign hash slots to nodes manually for now.
- [ ] Route requests to correct node for hash slot.
- [ ] Consider making a "KeyValueStoreFactory" that can get either a "LocalKeyValueStore" or a "RemoteKeyValueStore". Local means the data is stored on this node, while Remote means it's stored on a different node. \*Update definitely need to do this. The local store needs to be decoupled from RPC, because the RPC server side is going to need to reference the local store. Otherwise we'd have a circular module dependency.
- [ ] How to gracefully handle nodes going down?

# Building GRPC Code

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/rpc/node_rpc.proto
```

# Raft

- Website: https://raft.github.io/
- Paper: https://raft.github.io/raft.pdf
- Animation: https://thesecretlivesofdata.com/raft/
