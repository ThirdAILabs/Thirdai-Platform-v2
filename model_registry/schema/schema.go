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
	ModelType    string
	ModelSubtype string
	Access       string
	Size         int64
	Description  string
	Status       string
	StorageType  string
	Checksum     string
}

type AccessCode struct {
	gorm.Model

	AccessCode string `gorm:"uniqueIndex"`
	Name       string

	ModelID  uint
	ModelRef Model `gorm:"foreignKey:ModelID"`
}

type Admin struct {
	gorm.Model

	Email    string `gorm:"uniqueIndex"`
	Password string
}
