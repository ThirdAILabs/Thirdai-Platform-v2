package schema

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Model struct {
	Id uuid.UUID `gorm:"type:uuid;primaryKey"`

	Name string `gorm:"size:100;not null"`
	Type string `gorm:"size:100;not null"`

	PublishedDate time.Time

	TrainStatus  string `gorm:"size:100;not null"`
	DeployStatus string `gorm:"size:100;not null"`

	Access            string `gorm:"size:100;not null;default:'private'"`
	DefaultPermission string `gorm:"size:100;not null;default:'read'"`

	Attributes   []ModelAttribute  `gorm:"constraint:OnDelete:CASCADE"`
	Dependencies []ModelDependency `gorm:"foreignKey:ModelId;constraint:OnDelete:CASCADE"`

	BaseModelId *uuid.UUID `gorm:"type:uuid"`
	BaseModel   *Model     `gorm:"constraint:OnDelete:SET NULL"`

	UserId uuid.UUID `gorm:"type:uuid;not null"`
	User   *User

	TeamId *uuid.UUID `gorm:"type:uuid"`
	Team   *Team      `gorm:"constraint:OnDelete:SET NULL"`

	UserAPIKeys []UserAPIKey `gorm:"many2many:user_api_key_models;"`
}

func (m *Model) GetAttributes() map[string]string {
	attrs := make(map[string]string)
	for _, attr := range m.Attributes {
		attrs[attr.Key] = attr.Value
	}
	return attrs
}

type ModelAttribute struct {
	ModelId uuid.UUID `gorm:"type:uuid;primaryKey"`
	Key     string    `gorm:"primaryKey"`
	Value   string
}

type ModelDependency struct {
	ModelId      uuid.UUID `gorm:"type:uuid;primaryKey"`
	DependencyId uuid.UUID `gorm:"type:uuid;primaryKey"`

	Model      *Model `gorm:"foreignKey:ModelId"`
	Dependency *Model `gorm:"foreignKey:DependencyId"`
}

type User struct {
	Id uuid.UUID `gorm:"type:uuid;primaryKey"`

	Username string `gorm:"unique;size:50;not null"`
	Email    string `gorm:"unique;size:254;not null"`
	Password []byte

	IsAdmin bool `gorm:"not null;default:false"`

	Models []Model
	Teams  []UserTeam `gorm:"constraint:OnDelete:CASCADE"`
}

type UserAPIKey struct {
	Id uuid.UUID `gorm:"type:uuid;primaryKey"`

	HashKey string  `gorm:"column:hashkey;unique;size:500;not null;index"` // Added `index` as we use this field for comparing
	Name    string  `gorm:"size:500;not null"`
	Models  []Model `gorm:"many2many:user_api_key_models;constraint:OnDelete:CASCADE;"`

	AllModels bool `gorm:"default:false;not null"`

	GeneratedTime time.Time
	ExpiryTime    time.Time `gorm:"not null"`

	CreatedBy uuid.UUID `gorm:"type:uuid;not null"`
	User      User      `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE;"`
}

type Team struct {
	Id   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string    `gorm:"unique;size:100;not null"`
}

type UserTeam struct {
	UserId      uuid.UUID `gorm:"type:uuid;primaryKey"`
	TeamId      uuid.UUID `gorm:"type:uuid;primaryKey"`
	IsTeamAdmin bool      `gorm:"not null;default:false"`

	User *User `gorm:"constraint:OnDelete:CASCADE"`
	Team *Team `gorm:"constraint:OnDelete:CASCADE"`
}

type JobLog struct {
	Id      uuid.UUID `gorm:"type:uuid;primaryKey"`
	ModelId uuid.UUID `gorm:"type:uuid;index"`
	Job     string    `gorm:"size:50;not null"`
	Level   string    `gorm:"size:50;not null"`
	Message string
}

type Upload struct {
	Id     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId uuid.UUID `gorm:"type:uuid"`

	UploadDate time.Time
	Files      string

	User *User `gorm:"constraint:OnDelete:CASCADE"`
}

func (m *Model) TrainJobName() string {
	return fmt.Sprintf("train-%v-%v", m.Type, m.Id)
}

func (m *Model) DeployJobName() string {
	return fmt.Sprintf("deploy-%v-%v", m.Type, m.Id)
}
