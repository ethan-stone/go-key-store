# Overview

This document goes over how a cluster is set up with N number of nodes.

# Process

1. Start N number of nodes.
2. Use command line util where the grpc addresses of the nodes are provided.
3. The util creates cluster configurations for each node.
4. The util asks the user if the configuration is ok or not.
5. The configuration is written to each node.
