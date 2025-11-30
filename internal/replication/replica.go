package replication

import "github.com/ethan-stone/go-key-store/internal/wal"

type Replica struct {
	walEntryCh   chan *wal.WalEntry
	replayDoneCh chan struct{}
}

type ReplicaConfig struct {
	WalEntryCh   chan *wal.WalEntry
	ReplayDoneCh chan struct{}
}

func NewReplica(config *ReplicaConfig) *Replica {
	return &Replica{
		walEntryCh:   config.WalEntryCh,
		replayDoneCh: config.ReplayDoneCh,
	}
}

// type ReplayJob struct {
// 	snapshotSequenceNumber    uint64 // The sequence number of the last WAL entry that the replica has in it's latest snapshot.
// 	lastAppliedSequenceNumber uint64 // The sequence number of the last WAL entry that the replica has applied to it's local store.
// 	cutoffSequenceNumber      uint64 // The sequence number of the last WAL entry that we will send during the replay. Essentially this is the last sequence number in the log at the time the replay job is started.s
// }

// func (replica *Replica) Replay(job *ReplayJob) {
// 	go func() {
// 		defer close(replica.replayDoneCh)
// 	}()
// }
