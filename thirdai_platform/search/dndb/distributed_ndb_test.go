package dndb_test

import (
	"fmt"
	"testing"
	"thirdai_platform/search/dndb"
	"thirdai_platform/search/ndb"
	"time"

	"github.com/hashicorp/raft"
)

func createConfig(id, addr string, bootstrap bool) dndb.RaftConfig {
	return dndb.RaftConfig{
		ReplicaId:     id,
		BindAddr:      addr,
		SnapshotStore: raft.NewInmemSnapshotStore(),
		LogStore:      raft.NewInmemStore(),
		StableStore:   raft.NewInmemStore(),
		Bootstrap:     bootstrap,
	}
}

type testCluster []*dndb.DNdb

func waitForLeader(t *testing.T, timeout time.Duration, cluster testCluster) *dndb.DNdb {
	ticker := time.Tick(time.Second)
	cancel := time.After(timeout)

	for {
		select {
		case <-ticker:
			for _, node := range cluster {
				if node.IsLeader() {
					return node
				}
			}
		case <-cancel:
			t.Fatal("timeout reached before node is elected leader")
		}
	}
}

func waitForUpdate(t *testing.T, timeout time.Duration, node *dndb.DNdb, index uint64) {
	ticker := time.Tick(time.Second)
	cancel := time.After(timeout)

	for {
		select {
		case <-ticker:
			if node.LastUpdateIndex() >= index {
				return
			}
		case <-cancel:
			t.Fatal("timeout reached before node is elected leader")
		}
	}
}

func createCluster(t *testing.T, nNodes int, ndbPath string) []*dndb.DNdb {
	leader, err := dndb.New(ndbPath, t.TempDir(), createConfig("node0", "localhost:3000", true))
	if err != nil {
		t.Fatal(err)
	}

	cluster := testCluster{leader}

	waitForLeader(t, 10*time.Second, cluster)

	for i := 1; i < nNodes; i++ {
		id := fmt.Sprintf("node%d", i)
		addr := fmt.Sprintf("localhost:%d", 3000+i)
		follower, err := dndb.New(ndbPath, t.TempDir(), createConfig(id, addr, false))
		if err != nil {
			t.Fatalf("error creating follower: %v", err)
		}

		if err := leader.AddReplica(follower.ReplicaID(), follower.Addr()); err != nil {
			t.Fatalf("error adding follower to cluster: %v", err)
		}

		cluster = append(cluster, follower)
	}

	return cluster
}

func TestBasicReplication(t *testing.T) {
	ndbPath := t.TempDir()
	ndb, err := ndb.New(ndbPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := ndb.Insert("doc1", "id1", []string{"a b", "x y"}, nil, nil); err != nil {
		ndb.Free()
		t.Fatal(err)
	}

	ndb.Free()

	cluster := createCluster(t, 3, ndbPath)

	leader := waitForLeader(t, 10*time.Second, cluster)

	update, err := leader.Insert("doc2", "id2", []string{"a b c", "x y z"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, node := range cluster {
		waitForUpdate(t, 10*time.Second, node, update.Index)

		results, err := node.Query("a b c", 2, nil)
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 2 || results[0].Text != "a b c" {
			t.Fatalf("incorrect results: %v", results)
		}
	}
}

func TestAddNewReplica(t *testing.T) {

}

func TestRemoveReplica(t *testing.T) {

}

func TestRemoveLeader(t *testing.T) {

}
