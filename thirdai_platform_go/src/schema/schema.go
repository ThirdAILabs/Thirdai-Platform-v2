package schema

import "time"

type Model struct {
	Id string `gorm:"primaryKey"`

	Name          string
	Type          string
	Subtype       string
	PublishedDate time.Time

	TrainStatus  string
	DeployStatus string

	Access            string
	DefaultPermission string

	Attributes   []ModelAttribute  `gorm:"constraint:OnDelete:CASCADE;"`
	Dependencies []ModelDependency `gorm:"foreignKey:ModelId;constraint:OnDelete:CASCADE;"`

	ParentId *string
	Parent   *Model `gorm:"constraint:OnDelete:SET NULL;"`

	UserId string
	User   *User

	Permissions []ModelPermission

	TeamId *string
	Team   *Team
}

type ModelAttribute struct {
	ModelId string `gorm:"primaryKey"`
	Key     string `gorm:"primaryKey"`
	Value   string
}

type ModelDependency struct {
	ModelId      string `gorm:"primaryKey"`
	DependencyId string `gorm:"primaryKey"`

	Model      *Model `gorm:"foreignKey:ModelId"`
	Dependency *Model `gorm:"foreignKey:DependencyId"`
}

type ModelPermission struct {
	UserId     string `gorm:"primaryKey;constraint:OnDelete:CASCADE;"`
	ModelId    string `gorm:"primaryKey;constraint:OnDelete:CASCADE;"`
	Permission string
}

type User struct {
	Id string `gorm:"primaryKey"`

	Username string `gorm:"uniqueIndex"`
	Email    string `gorm:"uniqueIndex"`
	Password []byte

	IsAdmin bool

	Models []Model
	Teams  []UserTeam
}

type Team struct {
	Id   string `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex"`
}

type UserTeam struct {
	UserId      string `gorm:"primaryKey;constraint:OnDelete:CASCADE;"`
	TeamId      string `gorm:"primaryKey;constraint:OnDelete:CASCADE;"`
	IsTeamAdmin bool

	User *User
	Team *Team
}
