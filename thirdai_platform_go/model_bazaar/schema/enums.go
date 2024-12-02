package schema

import "fmt"

const (
	NotStarted = "not_started"
	Starting   = "starting"
	InProgress = "in_progress"
	Stopped    = "stopped"
	Complete   = "complete"
	Failed     = "failed"
)

const (
	Private   = "private"
	Protected = "protected"
	Public    = "public"
)

func CheckValidStatus(status string) error {
	if status == NotStarted || status == Starting || status == InProgress || status == Stopped || status == Complete || status == Failed {
		return nil
	}
	return fmt.Errorf("invalid status '%v'", status)
}

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

func CheckValidPermission(permission string) error {
	if permission == ReadPerm || permission == WritePerm {
		return nil
	}
	return fmt.Errorf("invalid permission %v, must be 'read' or 'write'", permission)
}

const (
	NdbModel         = "ndb"
	NlpTokenModel    = "nlp-token"
	NlpTextModel     = "nlp-text"
	EnterpriseSearch = "enterprise-search"
)

func CheckValidModelType(modelType string) error {
	if modelType == NdbModel || modelType == NlpTokenModel || modelType == NlpTextModel {
		return nil
	}
	return fmt.Errorf("invalid model type '%v'", modelType)
}
