package schema

import (
	"gorm.io/gorm"
)

const (
	public  = "public"
	private = "private"
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

	Token string

	ModelID  uint
	ModelRef Model
}

type Admin struct {
	gorm.Model

	Email    string `gorm:"uniqueIndex"`
	Password string
}
