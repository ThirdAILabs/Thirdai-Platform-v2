package schema

import (
	"gorm.io/gorm"
)

const (
	Public  = "public"
	Private = "private"
)

type Model struct {
	gorm.Model
	Name         string `gorm:"uniqueIndex"`
	Path         string
	Description  string
	ModelType    string
	ModelSubtype string
	Access       string
	Metadata     string

	// TODO: add checksum
}

type AccessToken struct {
	gorm.Model

	AccessToken string `gorm:"uniqueIndex"`
	Name        string

	ModelID  uint
	ModelRef Model
}

type Admin struct {
	gorm.Model

	Email    string `gorm:"uniqueIndex"`
	Password string
}
