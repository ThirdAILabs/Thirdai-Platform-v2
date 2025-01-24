package dndb

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type InsertOp struct {
	Document string
	DocId    string
	Chunks   []string
	Metadata []map[string]interface{}
}

type DeleteOp struct {
	DocId      string
	KeepLatest bool
}

type UpvoteOp struct {
	Query string
	Label uint64
}

type AssociateOp struct {
	Source string
	Target string
}

type UpdateOp struct {
	Insert    *InsertOp
	Delete    *DeleteOp
	Upvote    *UpvoteOp
	Associate *AssociateOp
}

func (op *UpdateOp) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(op); err != nil {
		return nil, fmt.Errorf("error serializing op: %w", err)
	}
	return buf.Bytes(), nil
}

func DeserializeOp(data []byte) (*UpdateOp, error) {
	var op UpdateOp
	if err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&op); err != nil {
		return nil, fmt.Errorf("error deserializing op: %w", err)
	}
	return &op, nil
}
