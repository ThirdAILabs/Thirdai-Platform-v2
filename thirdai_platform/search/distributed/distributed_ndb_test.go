package distributed_test

import (
	"testing"
	"thirdai_platform/search/distributed"
	"thirdai_platform/search/ndb"
	"time"

	"github.com/hashicorp/raft"
)

func createConfig(id, addr string, bootstrap bool) distributed.RaftConfig {
	return distributed.RaftConfig{
		ReplicaId:     id,
		BindAddr:      addr,
		SnapshotStore: raft.NewInmemSnapshotStore(),
		LogStore:      raft.NewInmemStore(),
		StableStore:   raft.NewInmemStore(),
		Bootstrap:     bootstrap,
	}
}

func waitForLeader(t *testing.T, dndb *distributed.DistributedNdb, timeout time.Duration) {
	ticker := time.Tick(time.Second)
	cancel := time.After(timeout)

	for {
		select {
		case <-ticker:
			if dndb.IsLeader() {
				return
			}
		case <-cancel:
			t.Fatal("timeout reached before node is elected leader")
		}
	}
}

func TestBasicClusterSetup(t *testing.T) {
	baseNdbPath := t.TempDir()
	ndb, err := ndb.New(baseNdbPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := ndb.Insert("doc1", "id1", []string{"a b", "x y"}, nil, nil); err != nil {
		ndb.Free()
		t.Fatal(err)
	}

	ndb.Free()

	node1, err := distributed.New(baseNdbPath, t.TempDir(), createConfig("node1", "localhost:3001", true))
	if err != nil {
		t.Fatal(err)
	}

	node2, err := distributed.New(baseNdbPath, t.TempDir(), createConfig("node2", "localhost:3002", false))
	if err != nil {
		t.Fatal(err)
	}

	waitForLeader(t, node1, 10*time.Second)

	if err := node1.AddReplica(node2.ReplicaID(), node2.Addr()); err != nil {
		t.Fatal(err)
	}

	node3, err := distributed.New(baseNdbPath, t.TempDir(), createConfig("node3", "localhost:3003", false))
	if err != nil {
		t.Fatal(err)
	}

	if err := node1.AddReplica(node3.ReplicaID(), node3.Addr()); err != nil {
		t.Fatal(err)
	}

	if err := node1.Insert("doc2", "id2", []string{"a b c", "x y z"}, nil); err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	{
		results, err := node2.Query("a b c", 2, nil)
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 2 || results[0].Text != "a b c" {
			t.Fatalf("incorrect results: %v", results)
		}
	}

	{
		results, err := node3.Query("a b c", 2, nil)
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 2 || results[0].Text != "a b c" {
			t.Fatalf("incorrect results: %v", results)
		}
	}
}
