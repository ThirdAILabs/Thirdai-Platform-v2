package distributed

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type InsertLogData struct {
	Document string
	DocId    string
	Chunks   []string
	Metadata []map[string]interface{}
}

type DeleteLocData struct {
	DocId      string
	KeepLatest bool
}

type UpvoteLogData struct {
	Query string
	Label uint64
}

type AssociateLogData struct {
	Source string
	Target string
}

type LogEntry struct {
	Insert    *InsertLogData
	Delete    *DeleteLocData
	Upvote    *UpvoteLogData
	Associate *AssociateLogData
}

func (log *LogEntry) SerializeLog() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(log); err != nil {
		return nil, fmt.Errorf("error serializing log: %w", err)
	}
	return buf.Bytes(), nil
}

func DeserializeLog(data []byte) (*LogEntry, error) {
	var log LogEntry
	if err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&log); err != nil {
		return nil, fmt.Errorf("error deserializing log: %w", err)
	}
	return &log, nil
}
