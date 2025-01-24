package dndb

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"thirdai_platform/search/ndb"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/raft"
)

var (
	ErrNotLeader = errors.New("operation can only be applied by leader")
)

type DNdb struct {
	sync.RWMutex

	ndb ndb.NeuralDB

	raft            *raft.Raft
	lastUpdateIndex atomic.Uint64

	replicaId     string
	bindAddr      string
	localNdbStore string
}

type RaftConfig struct {
	ReplicaId     string
	BindAddr      string
	SnapshotStore raft.SnapshotStore
	LogStore      raft.LogStore
	StableStore   raft.StableStore
	Bootstrap     bool
}

func New(ndbPath string, localNdbStore string, config RaftConfig) (*DNdb, error) {
	if config.ReplicaId == "" {
		config.ReplicaId = config.BindAddr
	}

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(config.ReplicaId)

	addr, err := net.ResolveTCPAddr("tcp", config.BindAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve bind addr: %w", err)
	}

	transport, err := raft.NewTCPTransport(config.BindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("unable to create transport: %w", err)
	}

	dndb := &DNdb{
		replicaId:     config.ReplicaId,
		bindAddr:      config.BindAddr,
		localNdbStore: localNdbStore,
	}

	ra, err := raft.NewRaft(raftConfig, (*distributedNdbFSM)(dndb), config.LogStore, config.StableStore, config.SnapshotStore, transport)
	if err != nil {
		return nil, fmt.Errorf("error constructing raft instance: %w", err)
	}

	dndb.raft = ra

	if config.Bootstrap {
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
			return nil, fmt.Errorf("error boostrapping raft cluster: %w", err)
		}
	}

	ndbCopyPath := filepath.Join(localNdbStore, uuid.NewString())
	if err := os.CopyFS(ndbCopyPath, os.DirFS(ndbPath)); err != nil {
		return nil, fmt.Errorf("error making replica of ndb: %w", err)
	}

	dndb.ndb, err = ndb.New(ndbCopyPath)
	if err != nil {
		return nil, fmt.Errorf("error opening ndb: %w", err)
	}

	return dndb, nil
}

func (dndb *DNdb) ReplicaID() string {
	return dndb.replicaId
}

func (dndb *DNdb) Addr() string {
	return dndb.bindAddr
}

func (dndb *DNdb) IsLeader() bool {
	return dndb.raft.State() == raft.Leader
}

func (dndb *DNdb) AddReplica(replicaId, addr string) error {
	if !dndb.IsLeader() {
		return ErrNotLeader
	}

	future := dndb.raft.AddVoter(raft.ServerID(replicaId), raft.ServerAddress(addr), 0, 0)
	if err := future.Error(); err != nil {
		return fmt.Errorf("error adding new replica: %w", err)
	}

	return nil
}

type UpdateResult struct {
	Index uint64
}

func (dndb *DNdb) Insert(document, docId string, chunks []string, metadata []map[string]interface{}) (UpdateResult, error) {
	// This is a little bit of a leaky abstraction but performing this check here is
	// so that we don't have to wait for raft to apply the insert to find out if the
	// args are valid.
	if err := ndb.CheckInsertArgs(document, docId, chunks, metadata); err != nil {
		return UpdateResult{}, err
	}

	op := UpdateOp{
		Insert: &InsertOp{
			Document: document, DocId: docId, Chunks: chunks, Metadata: metadata,
		},
	}

	return dndb.applyUpdate(op)
}

func (dndb *DNdb) Upvote(query string, label uint64) (UpdateResult, error) {
	op := UpdateOp{
		Upvote: &UpvoteOp{
			Query: query, Label: label,
		},
	}

	return dndb.applyUpdate(op)
}

func (dndb *DNdb) Associate(source, target string) (UpdateResult, error) {
	op := UpdateOp{
		Associate: &AssociateOp{
			Source: source, Target: target,
		},
	}

	return dndb.applyUpdate(op)
}

func (dndb *DNdb) Delete(docId string, keepLatestVersion bool) (UpdateResult, error) {
	op := UpdateOp{
		Delete: &DeleteOp{
			DocId: docId,
		},
	}

	return dndb.applyUpdate(op)
}

func (dndb *DNdb) applyUpdate(op UpdateOp) (UpdateResult, error) {
	if !dndb.IsLeader() {
		return UpdateResult{}, ErrNotLeader
	}

	serializedLog, err := op.Serialize()
	if err != nil {
		return UpdateResult{}, err
	}

	future := dndb.raft.Apply(serializedLog, 0)

	if err := future.Error(); err != nil {
		return UpdateResult{}, fmt.Errorf("error applying update: %w", err)
	}

	res := future.Response()
	if err, ok := res.(error); ok {
		return UpdateResult{}, err
	}

	return UpdateResult{Index: future.Index()}, nil
}

func (dndb *DNdb) LastUpdateIndex() uint64 {
	return dndb.lastUpdateIndex.Load()
}

func (dndb *DNdb) Query(query string, topk int, constraints ndb.Constraints) ([]ndb.Chunk, error) {
	dndb.RLock() // Prevent snapshots while reading from ndb
	defer dndb.RUnlock()

	return dndb.ndb.Query(query, topk, constraints)
}

func (dndb *DNdb) Sources() ([]ndb.Source, error) {
	dndb.RLock() // Prevent snapshots while reading from ndb
	defer dndb.RUnlock()

	return dndb.ndb.Sources()
}

// The FSM methods need to be public to be called by raft, but defining them on
// a non exported type ensures that they cannot be called outside of this package.
type distributedNdbFSM DNdb

func (dndb *distributedNdbFSM) Apply(raftLog *raft.Log) interface{} {
	op, err := DeserializeOp(raftLog.Data)
	if err != nil {
		slog.Error("error deserializing raft log", "error", err)
		return err
	}

	dndb.RLock() // Prevent snapshots while applying entries
	defer dndb.RUnlock()

	if op.Insert != nil {
		err := dndb.ndb.Insert(op.Insert.Document, op.Insert.DocId, op.Insert.Chunks, op.Insert.Metadata, nil)
		if err != nil {
			slog.Error("ndb insert failed", "error", err)
			return fmt.Errorf("ndb insert failed: %w", err)
		}
	}

	if op.Delete != nil {
		err := dndb.ndb.Delete(op.Delete.DocId, op.Delete.KeepLatest)
		if err != nil {
			slog.Error("ndb delete failed", "error", err)
			return fmt.Errorf("ndb delete failed: %w", err)
		}
	}

	if op.Upvote != nil {
		err := dndb.ndb.Finetune([]string{op.Upvote.Query}, []uint64{op.Upvote.Label})
		if err != nil {
			slog.Error("ndb upvote failed", "error", err)
			return fmt.Errorf("ndb upvote failed: %w", err)
		}
	}

	if op.Associate != nil {
		err := dndb.ndb.Associate([]string{op.Associate.Source}, []string{op.Associate.Target})
		if err != nil {
			slog.Error("ndb associate failed", "error", err)
			return fmt.Errorf("ndb associate failed: %w", err)
		}
	}

	dndb.lastUpdateIndex.Store(raftLog.Index)

	return nil
}

func (dndb *distributedNdbFSM) Snapshot() (raft.FSMSnapshot, error) {
	dndb.Lock()
	defer dndb.Unlock()

	snapshotPath := filepath.Join(dndb.localNdbStore, uuid.NewString())

	if err := dndb.ndb.Save(snapshotPath); err != nil {
		return nil, fmt.Errorf("ndb save failed: %w", err)
	}

	return &ndbSnapshot{path: snapshotPath}, nil
}

func saveFile(dstPath string, src io.Reader) error {
	file, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("error creating local file from tar file '%s' from snapshot: %w", dstPath, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, src); err != nil {
		return fmt.Errorf("error writing local file from tar file '%s' from snapshot: %w", dstPath, err)
	}

	return nil
}

func (dndb *distributedNdbFSM) Restore(snapshotReader io.ReadCloser) error {
	defer snapshotReader.Close()

	snapshotPath := filepath.Join(dndb.localNdbStore, uuid.NewString())

	reader := tar.NewReader(snapshotReader)

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading snapshot: %w", err)
		}

		path := filepath.Join(snapshotPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0777); err != nil {
				return fmt.Errorf("error creating subdirectory '%s' from snapshot: %w", path, err)
			}
		case tar.TypeReg:
			if err := saveFile(path, reader); err != nil {
				return err
			}
		}
	}

	snapshotNdb, err := ndb.New(snapshotPath)
	if err != nil {
		slog.Error("error loading ndb from snapshot", "path", snapshotPath, "error", err)
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
	archive := tar.NewWriter(sink)

	if err := archive.AddFS(os.DirFS(snapshot.path)); err != nil {
		sink.Cancel()
		return err
	}

	if err := archive.Close(); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (snapshot *ndbSnapshot) Release() {
	if err := os.RemoveAll(snapshot.path); err != nil {
		slog.Error("error deleting snapshot", "path", snapshot.path, "error", err)
	}
}
