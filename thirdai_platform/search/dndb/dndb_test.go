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

func createSimpleNdb(t *testing.T) string {
	ndbPath := t.TempDir()
	ndb, err := ndb.New(ndbPath)
	if err != nil {
		t.Fatal(err)
	}

	defer ndb.Free()

	if err := ndb.Insert("doc1", "id1", []string{"a b", "x y"}, nil, nil); err != nil {
		t.Fatal(err)
	}

	return ndbPath
}

func TestBasicReplication(t *testing.T) {
	ndbPath := createSimpleNdb(t)

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

func createClusterAndAddReplica(t *testing.T, snapshot bool) {
	ndbPath := createSimpleNdb(t)

	cluster := createCluster(t, 3, ndbPath)

	leader := waitForLeader(t, 10*time.Second, cluster)

	var updateIdx uint64
	for i := 0; i < 10; i++ {
		update, err := leader.Insert(fmt.Sprintf("new-%d", i), fmt.Sprintf("%d-%d", i, i), []string{fmt.Sprintf("%d %d", i, i+1)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		updateIdx = update.Index
	}

	if snapshot {
		if err := leader.ForceSnapshot(); err != nil {
			t.Fatal(err)
		}
	}

	newReplica, err := dndb.New(ndbPath, t.TempDir(), createConfig("node3", "localhost:3003", false))
	if err != nil {
		t.Fatalf("error creating new replica: %v", err)
	}

	sourcesBefore, err := newReplica.Sources()
	if err != nil {
		t.Fatal(err)
	}
	if len(sourcesBefore) != 1 {
		t.Fatal("new replica should only have initial source before joining cluster")
	}

	if err := leader.AddReplica(newReplica.ReplicaID(), newReplica.Addr()); err != nil {
		t.Fatalf("error adding new replica to cluster: %v", err)
	}

	waitForUpdate(t, 10*time.Second, newReplica, updateIdx)

	sourcesAfter, err := newReplica.Sources()
	if err != nil {
		t.Fatal(err)
	}
	if len(sourcesAfter) != 11 {
		t.Fatal("new replica should have all sources after joining cluster")
	}
}

func TestAddNewReplica(t *testing.T) {
	createClusterAndAddReplica(t, false)
}

func TestAddNewReplicaWithSnapshot(t *testing.T) {
	createClusterAndAddReplica(t, true)
}

func TestRemoveLeader(t *testing.T) {
	ndbPath := createSimpleNdb(t)

	cluster := createCluster(t, 3, ndbPath)

	leader1 := waitForLeader(t, 10*time.Second, cluster)

	update1, err := leader1.Insert("doc2", "id2", []string{"a b c", "x y z"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	waitForUpdate(t, 10*time.Second, leader1, update1.Index)

	if err := leader1.RemoveReplica(leader1.ReplicaID()); err != nil {
		t.Fatal(err)
	}

	leader2 := waitForLeader(t, 10*time.Second, cluster)

	if leader1.ReplicaID() == leader2.ReplicaID() {
		t.Fatal("leader should change")
	}

	update2, err := leader2.Insert("doc3", "id4", []string{"a b c d", "w x y z"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, node := range cluster {
		if node == leader1 {
			continue
		}

		waitForUpdate(t, 10*time.Second, node, update2.Index)

		results, err := node.Query("w x", 6, nil)
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 3 || results[0].Text != "w x y z" {
			t.Fatalf("incorrect results: %v", results)
		}
	}
}
