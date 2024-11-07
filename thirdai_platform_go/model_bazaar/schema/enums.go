package schema

import "fmt"

const (
	NotStarted = "not_started"
	Starting   = "starting"
	InProgress = "in_progress"
	Stopped    = "stopped"
	Complete   = "Complete"
	Failed     = "failed"
)

const (
	Private   = "private"
	Protected = "protected"
	Public    = "public"
)

func CheckValidAccess(access string) error {
	if access == Public || access == Private || access == Protected {
		return nil
	}
	return fmt.Errorf("invalid access %v, must be 'public', 'private', or 'protected'", access)
}

const (
	ReadPerm  = "read"
	WritePerm = "write"
)

func IsValidPermission(permission string) bool {
	return permission == ReadPerm || permission == WritePerm
}

func CheckValidPermission(permission string) error {
	if permission == ReadPerm || permission == WritePerm {
		return nil
	}
	return fmt.Errorf("invalid permission %v, must be 'read' or 'write'", permission)
}

const (
	NdbModel      = "ndb"
	NlpTokenModel = "nlp-token"
	NlpTextModel  = "nlp-text"
)
