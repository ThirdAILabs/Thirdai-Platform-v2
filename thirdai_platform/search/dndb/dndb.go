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

type DNDB struct {
	sync.RWMutex

	ndb ndb.NeuralDB

	raft            *raft.Raft
	lastUpdateIndex atomic.Uint64

	replicaId     string
	bindAddr      string
	localNdbStore string

	logger *slog.Logger
}

type RaftConfig struct {
	ReplicaId     string
	BindAddr      string
	Transport     raft.Transport
	SnapshotStore raft.SnapshotStore
	LogStore      raft.LogStore
	StableStore   raft.StableStore
	Bootstrap     bool
}

func CreateTcpTransport(bindAddr string) (raft.Transport, error) {
	addr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve bind addr: %w", err)
	}

	const maxConnPoolSize = 10
	transport, err := raft.NewTCPTransport(bindAddr, addr, maxConnPoolSize, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("unable to create transport: %w", err)
	}

	return transport, nil
}

func New(ndbPath string, localNdbStore string, config RaftConfig) (*DNDB, error) {
	logger := slog.With("replica_id", config.ReplicaId, "addr", config.BindAddr)

	logger.Info("[DNDB]: initializing", "ndb_path", ndbPath, "local_ndb_store", localNdbStore, "boostrap", config.Bootstrap)
	if config.ReplicaId == "" {
		return nil, errors.New("replica id cannot be empty")
	}

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(config.ReplicaId)

	dndb := &DNDB{
		replicaId:     config.ReplicaId,
		bindAddr:      config.BindAddr,
		localNdbStore: localNdbStore,
		logger:        logger,
	}

	ra, err := raft.NewRaft(raftConfig, (*distributedNdbFSM)(dndb), config.LogStore, config.StableStore, config.SnapshotStore, config.Transport)
	if err != nil {
		logger.Error("[DNDB]: error constructing raft server", "error", err)
		return nil, fmt.Errorf("error constructing raft instance: %w", err)
	}

	dndb.raft = ra

	if config.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: config.Transport.LocalAddr(),
				},
			},
		}
		err := ra.BootstrapCluster(configuration).Error()
		if err != nil {
			logger.Error("[DNDB]: boostrapping raft cluster failed", "error", err)
			return nil, fmt.Errorf("error bootstrapping raft cluster: %w", err)
		}
	}

	ndbCopyPath := filepath.Join(localNdbStore, uuid.NewString())
	if err := os.CopyFS(ndbCopyPath, os.DirFS(ndbPath)); err != nil {
		logger.Error("[DNDB]: error creating copy of ndb on disk", "error", err)
		return nil, fmt.Errorf("error making replica of ndb: %w", err)
	}

	dndb.ndb, err = ndb.New(ndbCopyPath)
	if err != nil {
		logger.Error("[DNDB]: error opening ndb", "error", err)
		return nil, fmt.Errorf("error opening ndb: %w", err)
	}

	logger.Info("[DNDB]: completed initialization")

	return dndb, nil
}

func (dndb *DNDB) Shutdown() error {
	defer dndb.ndb.Free()

	dndb.logger.Info("[DNDB]: shutting down")

	if err := dndb.raft.Shutdown().Error(); err != nil {
		dndb.logger.Error("error calling raft shutdown", "error", err)
		return fmt.Errorf("error calling raft shutdown: %w", err)
	}

	dndb.logger.Info("[DNDB]: shutdown complete")

	return nil
}

func (dndb *DNDB) ReplicaID() string {
	return dndb.replicaId
}

func (dndb *DNDB) Addr() string {
	return dndb.bindAddr
}

func (dndb *DNDB) IsLeader() bool {
	return dndb.raft.State() == raft.Leader
}

func (dndb *DNDB) AddReplica(replicaId, addr string) error {
	dndb.logger.Info("[DNDB]: adding replica", "new_replica_Id", replicaId, "new_addr", addr)
	if !dndb.IsLeader() {
		dndb.logger.Error("[DNDB]: cannot add replica, node is not leader", "new_replica_Id", replicaId, "new_addr", addr)
		return ErrNotLeader
	}

	// The final args are not needed, 0 means that they are ignored.
	future := dndb.raft.AddVoter(raft.ServerID(replicaId), raft.ServerAddress(addr), 0, 0)
	if err := future.Error(); err != nil {
		dndb.logger.Error("[DNDB]: error adding replica", "new_replica_Id", replicaId, "new_addr", addr, "error", err)
		return fmt.Errorf("error adding new replica: %w", err)
	}

	dndb.logger.Info("[DNDB]: replica added", "new_replica_Id", replicaId, "new_addr", addr)

	return nil
}

func (dndb *DNDB) RemoveReplica(replicaId string) error {
	dndb.logger.Info("dndb: removing replica", "removed_replica_id", replicaId)
	if !dndb.IsLeader() {
		dndb.logger.Error("[DNDB]: cannot remove replica, node is not leader", "removed_replica_id", replicaId)
		return ErrNotLeader
	}

	future := dndb.raft.RemoveServer(raft.ServerID(replicaId), 0, 0)
	if err := future.Error(); err != nil {
		dndb.logger.Error("[DNDB]: error removing replica", "removed_replica_id", replicaId, "error", err)
		return fmt.Errorf("error removing replica: %w", err)
	}

	dndb.logger.Info("[DNDB]: replica removed", "removed_replica_id", replicaId)

	return nil
}

func (dndb *DNDB) ForceSnapshot() error {
	dndb.logger.Info("[DNDB]: creating snapshot")
	if !dndb.IsLeader() {
		dndb.logger.Error("[DNDB]: cannot create snapshot, node is not leader")
		return ErrNotLeader
	}

	future := dndb.raft.Snapshot()
	if err := future.Error(); err != nil {
		dndb.logger.Error("[DNDB]: error creating snapshot")
		return fmt.Errorf("error creating snapshot: %w", err)
	}

	dndb.logger.Info("[DNDB]: snapshot created")

	return nil
}

type UpdateResult struct {
	Index uint64
}

func (dndb *DNDB) Insert(document, docId string, chunks []string, metadata []map[string]interface{}) (UpdateResult, error) {
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

func (dndb *DNDB) Upvote(query string, label uint64) (UpdateResult, error) {
	op := UpdateOp{
		Upvote: &UpvoteOp{
			Query: query, Label: label,
		},
	}

	return dndb.applyUpdate(op)
}

func (dndb *DNDB) Associate(source, target string, strength uint32) (UpdateResult, error) {
	if strength == 0 {
		strength = ndb.DefaultAssociateStrength
	}
	op := UpdateOp{
		Associate: &AssociateOp{
			Source: source, Target: target, Strength: strength,
		},
	}

	return dndb.applyUpdate(op)
}

func (dndb *DNDB) Delete(docId string, keepLatestVersion bool) (UpdateResult, error) {
	op := UpdateOp{
		Delete: &DeleteOp{
			DocId:      docId,
			KeepLatest: keepLatestVersion,
		},
	}

	return dndb.applyUpdate(op)
}

func (dndb *DNDB) applyUpdate(op UpdateOp) (UpdateResult, error) {
	opType := op.Op()
	dndb.logger.Info("[DNDB]: applying update", "op", opType)

	if !dndb.IsLeader() {
		dndb.logger.Error("[DNDB]: cannot apply update, node is not leader", "op", opType)
		return UpdateResult{}, ErrNotLeader
	}

	serializedLog, err := op.Serialize()
	if err != nil {
		dndb.logger.Error("[DNDB]: error serializing update op", "op", opType, "error", err)
		return UpdateResult{}, err
	}

	future := dndb.raft.Apply(serializedLog, 0)

	if err := future.Error(); err != nil {
		dndb.logger.Error("[DNDB]: raft apply failed", "op", opType, "error", err)
		return UpdateResult{}, fmt.Errorf("error applying update: %w", err)
	}

	res := future.Response()
	if err, ok := res.(error); ok {
		dndb.logger.Error("[DNDB]: update returned error", "op", opType, "error", err)
		return UpdateResult{}, err
	}

	dndb.logger.Info("[DNDB]: update committed", "op", opType, "index", future.Index())

	return UpdateResult{Index: future.Index()}, nil
}

func (dndb *DNDB) LastUpdateIndex() uint64 {
	return dndb.lastUpdateIndex.Load()
}

func (dndb *DNDB) Query(query string, topk int, constraints ndb.Constraints) ([]ndb.Chunk, error) {
	dndb.RLock() // Prevent snapshots while reading from ndb
	defer dndb.RUnlock()

	return dndb.ndb.Query(query, topk, constraints)
}

func (dndb *DNDB) Sources() ([]ndb.Source, error) {
	dndb.RLock() // Prevent snapshots while reading from ndb
	defer dndb.RUnlock()

	return dndb.ndb.Sources()
}

// The FSM methods need to be public to be called by raft, but defining them on
// a non exported type ensures that they cannot be called outside of this package.
type distributedNdbFSM DNDB

func (dndb *distributedNdbFSM) Apply(raftLog *raft.Log) interface{} {
	dndb.logger.Info("[DNDB]: applying update to fsm", "index", raftLog.Index)

	op, err := DeserializeOp(raftLog.Data)
	if err != nil {
		dndb.logger.Error("[DNDB]: error deserializing raft log", "index", raftLog.Index, "error", err)
		return err
	}

	dndb.RLock() // Prevent snapshots while applying entries
	defer dndb.RUnlock()

	if op.Insert != nil {
		err := dndb.ndb.Insert(op.Insert.Document, op.Insert.DocId, op.Insert.Chunks, op.Insert.Metadata, nil)
		if err != nil {
			dndb.logger.Error("[DNDB]: ndb insert failed", "index", raftLog.Index, "error", err)
			return fmt.Errorf("ndb insert failed: %w", err)
		}
	}

	if op.Delete != nil {
		err := dndb.ndb.Delete(op.Delete.DocId, op.Delete.KeepLatest)
		if err != nil {
			dndb.logger.Error("[DNDB]: ndb delete failed", "index", raftLog.Index, "error", err)
			return fmt.Errorf("ndb delete failed: %w", err)
		}
	}

	if op.Upvote != nil {
		err := dndb.ndb.Finetune([]string{op.Upvote.Query}, []uint64{op.Upvote.Label})
		if err != nil {
			dndb.logger.Error("[DNDB]: ndb upvote failed", "index", raftLog.Index, "error", err)
			return fmt.Errorf("ndb upvote failed: %w", err)
		}
	}

	if op.Associate != nil {
		err := dndb.ndb.Associate([]string{op.Associate.Source}, []string{op.Associate.Target}, op.Associate.Strength)
		if err != nil {
			dndb.logger.Error("[DNDB]: ndb associate failed", "index", raftLog.Index, "error", err)
			return fmt.Errorf("ndb associate failed: %w", err)
		}
	}

	dndb.lastUpdateIndex.Store(raftLog.Index)

	dndb.logger.Info("[DNDB]: update applied to fsm", "index", raftLog.Index)

	return nil
}

func (dndb *distributedNdbFSM) Snapshot() (raft.FSMSnapshot, error) {
	dndb.Lock()
	defer dndb.Unlock()

	dndb.logger.Info("[DNDB]: fsm creating snapshot")

	snapshotPath := filepath.Join(dndb.localNdbStore, uuid.NewString())

	if err := dndb.ndb.Save(snapshotPath); err != nil {
		dndb.logger.Error("[DNDB]: fsm create snapshot failed", "error", err)
		return nil, fmt.Errorf("ndb save failed: %w", err)
	}

	dndb.logger.Info("[DNDB]: fsm snapshot created")

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

	dndb.logger.Info("[DNDB]: restoring from snapshot")

	snapshotPath := filepath.Join(dndb.localNdbStore, uuid.NewString())

	reader := tar.NewReader(snapshotReader)

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			dndb.logger.Info("[DNDB]: error reading snapshot", "error", err)
			return fmt.Errorf("error reading snapshot: %w", err)
		}

		path := filepath.Join(snapshotPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0666); err != nil {
				dndb.logger.Info("[DNDB]: error creating subdirectory for snapshot", "error", err)
				return fmt.Errorf("error creating subdirectory '%s' from snapshot: %w", path, err)
			}
		case tar.TypeReg:
			if err := saveFile(path, reader); err != nil {
				dndb.logger.Info("[DNDB]: error saving file in snapshot", "error", err)
				return err
			}
		}
	}

	snapshotNdb, err := ndb.New(snapshotPath)
	if err != nil {
		dndb.logger.Error("[DNDB]: error loading ndb from snapshot", "path", snapshotPath, "error", err)
		return fmt.Errorf("error loading ndb from snapshot: %w", err)
	}

	dndb.Lock()
	defer dndb.Unlock()

	dndb.ndb.Free()
	dndb.ndb = snapshotNdb

	dndb.logger.Info("[DNDB]: restore from snapshot completed")

	return nil
}

type ndbSnapshot struct {
	path string
}

func (snapshot *ndbSnapshot) Persist(sink raft.SnapshotSink) error {
	archive := tar.NewWriter(sink)

	slog.Info("[DNDB SNAPSHOT]: persisting snapshot", "path", snapshot.path)

	if err := archive.AddFS(os.DirFS(snapshot.path)); err != nil {
		slog.Error("[DNDB SNAPSHOT]: error adding snapshot to archive", "path", snapshot.path, "error", err)
		return errors.Join(err, sink.Cancel())
	}

	if err := archive.Close(); err != nil {
		slog.Error("[DNDB SNAPSHOT]: error closing snapshot", "path", snapshot.path, "error", err)
		return errors.Join(err, sink.Cancel())
	}

	slog.Info("[DNDB SNAPSHOT]: snapshot persisted", "path", snapshot.path)

	return sink.Close()
}

func (snapshot *ndbSnapshot) Release() {
	slog.Info("[DNDB SNAPSHOT]: releasing snapshot", "path", snapshot.path)

	if err := os.RemoveAll(snapshot.path); err != nil {
		slog.Error("[DNDB SNAPSHOT]: error deleting snapshot", "path", snapshot.path, "error", err)
	}

	slog.Info("[DNDB SNAPSHOT]: snapshot released", "path", snapshot.path)

}
