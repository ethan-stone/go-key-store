# Overview

This repo is an attempt to many of the concepts related to distributed systems.

# Todo

- [ ] Super basic key value store server.
- [ ] Ping between nodes in a cluster via manual configuration.
- [ ] Data replication between nodes in a cluster. Replicate data on all nodes at first.

# Building GRPC Code

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/rpc/node_rpc.proto
```
