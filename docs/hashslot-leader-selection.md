# Overview

This doc goes over how leader selection for a range of hashslots work.

The hashslots are divided amongst all nodes in the cluster. Every node is a leader (in raft terminology) for a range of hashslots.
