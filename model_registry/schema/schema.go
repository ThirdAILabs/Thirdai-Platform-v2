package schema

import (
	"gorm.io/gorm"
)

const (
	Public  = "public"
	Private = "private"
)

const (
	Pending  = "pending"
	Commited = "commited"
)

type Model struct {
	gorm.Model
	Name         string `gorm:"uniqueIndex"`
	Description  string
	ModelType    string
	ModelSubtype string
	Access       string
	Metadata     string
	Size         int64
	Status       string
	StorageType  string
	// TODO: add checksum
}

type AccessToken struct {
	gorm.Model

	AccessToken string `gorm:"uniqueIndex"`
	Name        string

	ModelID  uint
	ModelRef Model `gorm:"foreignKey:ModelID"`
}

type Admin struct {
	gorm.Model

	Email    string `gorm:"uniqueIndex"`
	Password string
}
