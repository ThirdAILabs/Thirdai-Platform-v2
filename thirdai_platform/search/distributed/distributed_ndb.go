package distributed

import (
	"encoding/gob"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"thirdai_platform/search/ndb"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/raft"
)

type DistributedNdb struct {
	sync.RWMutex

	ndb ndb.NeuralDB

	raft *raft.Raft

	snapshotDir string
}

func New(ndbPath, snapshotDir, bindAddr string, bootstrap bool) (*DistributedNdb, error) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(bindAddr)

	addr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(bindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	snapshots, err := raft.NewFileSnapshotStore("TODO", 2, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("file snapshot store: %s", err)
	}

	loadedNdb, err := ndb.New(ndbPath) // TODO: cleanup ndb on error?
	if err != nil {
		return nil, err
	}

	dndb := &DistributedNdb{
		ndb:         loadedNdb,
		snapshotDir: snapshotDir,
	}

	ra, err := raft.NewRaft(raftConfig, dndb, raft.NewInmemStore(), raft.NewInmemStore(), snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("new raft: %s", err)
	}

	dndb.raft = ra

	if bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		err := ra.BootstrapCluster(configuration).Error()
		if err != nil {
			return nil, err
		}
	}

	return dndb, nil
}

func (dndb *DistributedNdb) Apply(raftLog *raft.Log) interface{} {
	dndb.RLock() // Prevent snapshots while applying entries
	defer dndb.RUnlock()

	log, err := DeserializeLog(raftLog.Data)
	if err != nil {
		slog.Error("error deserializing raft log", "error", err)
		return err
	}

	if log.Insert != nil {
		err := dndb.ndb.Insert(log.Insert.Document, log.Insert.DocId, log.Insert.Chunks, log.Insert.Metadata, nil)
		if err != nil {
			slog.Error("ndb insert failed", "error", err)
			return fmt.Errorf("ndb insert failed: %w", err)
		}
	}

	if log.Delete != nil {
		err := dndb.ndb.Delete(log.Delete.DocId, log.Delete.KeepLatest)
		if err != nil {
			slog.Error("ndb delete failed", "error", err)
			return fmt.Errorf("ndb delete failed: %w", err)
		}
	}

	if log.Upvote != nil {
		err := dndb.ndb.Finetune([]string{log.Upvote.Query}, []uint64{log.Upvote.Label})
		if err != nil {
			slog.Error("ndb upvote failed", "error", err)
			return fmt.Errorf("ndb upvote failed: %w", err)
		}
	}

	if log.Associate != nil {
		err := dndb.ndb.Associate([]string{log.Associate.Source}, []string{log.Associate.Target})
		if err != nil {
			slog.Error("ndb associate failed", "error", err)
			return fmt.Errorf("ndb associate failed: %w", err)
		}
	}

	return nil
}

func (dndb *DistributedNdb) Snapshot() (raft.FSMSnapshot, error) {
	dndb.Lock()
	defer dndb.Unlock()

	snapshotPath := filepath.Join(dndb.snapshotDir, uuid.NewString())

	if err := dndb.ndb.Save(snapshotPath); err != nil {
		return nil, fmt.Errorf("ndb save failed: %w", err)
	}

	return &ndbSnapshot{path: snapshotPath}, nil
}

func (dndb *DistributedNdb) Restore(snapshotReader io.ReadCloser) error {
	defer snapshotReader.Close()

	var snapshot ndbSnapshot
	if err := gob.NewDecoder(snapshotReader).Decode(&snapshot); err != nil {
		slog.Error("error reading snapshot data", "error", err)
		return fmt.Errorf("error reading snapshot data: %w", err)
	}

	snapshotNdb, err := ndb.New(snapshot.path)
	if err != nil {
		slog.Error("error loading ndb from snapshot", "path", snapshot.path, "error", err)
		return fmt.Errorf("error loading ndb from snapshot: %w", err)
	}

	dndb.Lock()
	defer dndb.Unlock()

	dndb.ndb.Free()
	dndb.ndb = snapshotNdb

	return nil
}

type ndbSnapshot struct {
	path string
}

func (snapshot *ndbSnapshot) Persist(sink raft.SnapshotSink) error {
	if err := gob.NewEncoder(sink).Encode(snapshot); err != nil {
		slog.Error("error saving snapshot", "error", err)
		sink.Cancel()
		return fmt.Errorf("error saving snapshot: %w", err)
	}

	return sink.Close()
}

func (snapshot *ndbSnapshot) Release() {
	if err := os.RemoveAll(snapshot.path); err != nil {
		slog.Error("error deleting snapshot", "path", snapshot.path, "error", err)
	}
}
