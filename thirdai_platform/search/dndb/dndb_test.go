package dndb_test

import (
	"encoding/csv"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"testing"
	"thirdai_platform/search/dndb"
	"thirdai_platform/search/ndb"
	"time"

	"github.com/hashicorp/raft"
)

func init() {
	const licensePath = "../../../.test_license/thirdai.license"
	if err := ndb.SetLicensePath(licensePath); err != nil {
		panic(err)
	}
}

func createConfig(id string, transport raft.Transport, bootstrap bool) dndb.RaftConfig {
	return dndb.RaftConfig{
		ReplicaId:     id,
		BindAddr:      id,
		Transport:     transport,
		SnapshotStore: raft.NewInmemSnapshotStore(),
		LogStore:      raft.NewInmemStore(),
		StableStore:   raft.NewInmemStore(),
		Bootstrap:     bootstrap,
	}
}

type testCluster []*dndb.DNDB

func waitForLeader(t *testing.T, timeout time.Duration, cluster testCluster) *dndb.DNDB {
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

func waitForUpdate(t *testing.T, timeout time.Duration, node *dndb.DNDB, index uint64) {
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

func createTransports(nNodes int) []*raft.InmemTransport {
	transports := make([]*raft.InmemTransport, 0, nNodes)
	for i := 0; i < nNodes; i++ {
		_, t := raft.NewInmemTransport(raft.ServerAddress(fmt.Sprintf("node%d", i)))
		transports = append(transports, t)
	}

	for _, t0 := range transports {
		for _, t1 := range transports {
			t0.Connect(t1.LocalAddr(), t1)
		}
	}

	return transports
}

func createCluster(t *testing.T, transports []*raft.InmemTransport, ndbPath string) []*dndb.DNDB {
	leader, err := dndb.New(ndbPath, t.TempDir(), createConfig("node0", transports[0], true))
	if err != nil {
		t.Fatal(err)
	}

	cluster := testCluster{leader}

	waitForLeader(t, 10*time.Second, cluster)

	for i := 1; i < len(transports); i++ {
		id := fmt.Sprintf("node%d", i)
		follower, err := dndb.New(ndbPath, t.TempDir(), createConfig(id, transports[i], false))
		if err != nil {
			t.Fatalf("error creating follower: %v", err)
		}

		if err := leader.AddReplica(follower.ReplicaID(), follower.Addr()); err != nil {
			t.Fatalf("error adding follower to cluster: %v", err)
		}

		cluster = append(cluster, follower)
	}

	t.Cleanup(func() {
		for _, node := range cluster {
			if err := node.Shutdown(); err != nil {
				t.Fatalf("shutdown error: %v", err)
			}
		}
	})

	return cluster
}

func createSimpleNdb(t *testing.T, empty bool) string {
	ndbPath := t.TempDir()
	ndb, err := ndb.New(ndbPath)
	if err != nil {
		t.Fatal(err)
	}

	defer ndb.Free()

	if !empty {
		if err := ndb.Insert("doc1", "id1", []string{"a b", "x y"}, nil, nil); err != nil {
			t.Fatal(err)
		}
	}

	return ndbPath
}

func TestBasicReplication(t *testing.T) {
	ndbPath := createSimpleNdb(t, false)

	transports := createTransports(3)
	cluster := createCluster(t, transports, ndbPath)

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
	ndbPath := createSimpleNdb(t, false)

	transports := createTransports(4)
	cluster := createCluster(t, transports[:3], ndbPath)

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
		stop := make(chan bool)
		failed := 0
		go func() { // Runs queries in background while snapshot is called.
			for {
				select {
				case <-stop:
					return
				default:
					if _, err := leader.Query("some random query", 5, nil); err != nil {
						failed++
					}
				}
			}
		}()

		time.Sleep(10 * time.Millisecond) // Allow queries to start

		if err := leader.ForceSnapshot(); err != nil {
			t.Fatal(err)
		}

		close(stop)

		if failed > 0 {
			t.Fatal("background queries failed")
		}
	}

	newReplica, err := dndb.New(ndbPath, t.TempDir(), createConfig("node3", transports[3], false))
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
	ndbPath := createSimpleNdb(t, false)

	transports := createTransports(3)
	cluster := createCluster(t, transports, ndbPath)

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

func subsampleQuery(query string) string {
	tokens := strings.Split(query, " ")
	rand.Shuffle(len(tokens), func(i, j int) {
		tokens[i], tokens[j] = tokens[j], tokens[i]
	})
	return strings.Join(tokens[:int(float64(len(tokens))*0.7)], " ")
}

type sample struct {
	text string
	id   int
}

func getQueryAccuracy(t *testing.T, dndb *dndb.DNDB, samples []sample) float64 {
	correct := 0
	for _, s := range samples {
		results, err := dndb.Query(subsampleQuery(s.text), 5, nil)
		if err != nil {
			t.Fatalf("query error: %v", err)
		}

		if len(results) > 0 && int(results[0].Id) == s.id {
			correct++
		}
	}

	return float64(correct) / float64(len(samples))
}

func checkBasicQueryAccuracy(t *testing.T, dndb *dndb.DNDB, samples []sample) {
	for _, sample := range samples {
		results, err := dndb.Query(subsampleQuery(sample.text), 5, nil)
		if err != nil {
			t.Fatalf("query error: %v", err)
		}
		if len(results) < 1 || int(results[0].Id) != sample.id {
			t.Fatalf("incorrect query results, expected %d, got %d", results[0].Id, sample.id)
		}
	}
}

func chooseN(words []string, n int) []string {
	selection := make([]string, 0)
	for i := 0; i < n; i++ {
		selection = append(selection, words[rand.IntN(len(words))])
	}
	return selection
}

func createAcronymQueries(samples []sample) []sample {
	randomWords := make([]string, 0)
	for _, sample := range samples {
		words := strings.Split(sample.text, " ")
		randomWords = append(randomWords, chooseN(words, 4)...)
	}

	newSamples := make([]sample, 0, len(samples))
	for _, s := range samples {
		acronym := make([]byte, 0)
		for _, word := range strings.Split(s.text, " ") {
			if len(word) > 0 {
				acronym = append(acronym, word[0])
			}
		}

		newQuery := string(acronym) + " " + strings.Join(chooseN(randomWords, 5), " ")

		newSamples = append(newSamples, sample{text: newQuery, id: s.id})
	}

	return newSamples
}

func createAssociations(samples []sample) ([]sample, []string) {
	newSamples := make([]sample, 0, len(samples))
	targets := make([]string, 0, len(samples))
	for i, s := range samples {
		newSamples = append(newSamples, sample{
			text: fmt.Sprintf("%d %d %d", i, i+len(samples), i+2*len(samples)),
			id:   s.id,
		})
		targets = append(targets,
			strings.Join(chooseN(strings.Split(samples[i].text, " "), 10), " "),
		)
	}

	return newSamples, targets
}

func loadSamples(t *testing.T) []sample {
	file, err := os.Open("../../integration_tests/data/articles.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	samples := make([]sample, 0)

	for _, row := range rows[1:] {
		id, err := strconv.Atoi(row[1])
		if err != nil {
			t.Fatal(err)
		}

		samples = append(samples, sample{row[0], id})
	}

	return samples
}

func TestComplicatedUpdates(t *testing.T) {
	ndbPath := createSimpleNdb(t, true)

	transports := createTransports(5)
	cluster := createCluster(t, transports, ndbPath)

	leader := waitForLeader(t, 10*time.Second, cluster)

	documents := loadSamples(t)
	acronyms := createAcronymQueries(documents)
	associationSamples, targets := createAssociations(documents)

	for i := 0; i < len(documents); i += 5 {
		doc := fmt.Sprintf("doc-%d", i)
		chunks := []string{}
		for _, s := range documents[i:min(i+5, len(documents))] {
			chunks = append(chunks, s.text)
		}
		_, err := leader.Insert(doc, doc, chunks, nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	var updateIdx uint64
	{
		update, err := leader.Insert("random-doc", "random-doc", []string{"? ? ? ?"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		updateIdx = update.Index
	}

	for _, node := range cluster {
		waitForUpdate(t, 30*time.Second, node, updateIdx)

		checkBasicQueryAccuracy(t, leader, documents)

		if acc := getQueryAccuracy(t, node, acronyms); acc >= 0.5 {
			t.Fatalf("accuracy should be < 0.5 before upvote: %f", acc)
		}

		if acc := getQueryAccuracy(t, node, associationSamples); acc >= 0.5 {
			t.Fatalf("accuracy should be < 0.5 before associate: %f", acc)
		}

		sources, err := node.Sources()
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, source := range sources {
			if source.DocId == "random-doc" {
				found = true
			}
		}

		if !found {
			t.Fatal("random doc should be present")
		}
	}

	for _, acronymn := range acronyms {
		_, err := leader.Upvote(acronymn.text, uint64(acronymn.id))
		if err != nil {
			t.Fatal(err)
		}
	}

	strength := 1
	for i, association := range associationSamples {
		_, err := leader.Associate(association.text, targets[i], uint32(strength))
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		update, err := leader.Delete("random-doc", false)
		if err != nil {
			t.Fatal(err)
		}
		updateIdx = update.Index
	}

	for _, node := range cluster {
		waitForUpdate(t, 30*time.Second, node, updateIdx)

		checkBasicQueryAccuracy(t, leader, documents)

		if acc := getQueryAccuracy(t, node, acronyms); acc < 0.9 {
			t.Fatalf("accuracy should be >= 0.9 after upvote: %f", acc)
		}

		// TODO(Nicholas): Make pass strength so this works
		if acc := getQueryAccuracy(t, node, associationSamples); acc < 0.9 {
			t.Fatalf("accuracy should be >= 0.9 after associate: %f", acc)
		}

		sources, err := node.Sources()
		if err != nil {
			t.Fatal(err)
		}

		for _, source := range sources {
			if source.DocId == "random-doc" {
				t.Fatal("doc should be deleted")
			}
		}
	}
}
