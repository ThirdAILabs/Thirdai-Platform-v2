package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Model struct {
	Id           string            `gorm:"primaryKey"`
	Attributes   []ModelAttribute  `gorm:"constraint:OnDelete:CASCADE;"`
	Dependencies []ModelDependency `gorm:"foreignKey:ModelId;constraint:OnDelete:CASCADE;"`

	ParentId *string
	Parent   *Model `gorm:"constraint:OnDelete:SET NULL;"`
}

type ModelAttribute struct {
	// ID  uint `gorm:"primaryKey"`
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

// Useful: https://gorm.io/docs/has_many.html
// Maybe userful: https://gorm.io/docs/associations.html#Finding-Associations

func main() {
	// dsn := "host=localhost user=postgres password=password dbname=test port=5432 sslmode=disable"
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Model{})
	db.AutoMigrate(&ModelAttribute{})
	db.AutoMigrate(&ModelDependency{})

	db.Create(&Model{Id: strconv.Itoa(0)})
	for i := 1; i < 3; i++ {
		p := strconv.Itoa(i - 1)
		db.Create(&Model{Id: strconv.Itoa(i), ParentId: &p})
	}

	db.Create(&ModelAttribute{ModelId: "0", Key: "a", Value: "a0"})
	db.Create(&ModelAttribute{ModelId: "0", Key: "b", Value: "b0"})
	db.Create(&ModelAttribute{ModelId: "1", Key: "a", Value: "a1"})

	db.Create(&ModelDependency{ModelId: "0", DependencyId: "1"})
	db.Create(&ModelDependency{ModelId: "0", DependencyId: "2"})
	db.Create(&ModelDependency{ModelId: "1", DependencyId: "2"})

	for i := 0; i < 3; i++ {
		var u Model
		db.Model(&Model{}).Preload("Parent").Preload("Attributes").Preload("Dependencies").Preload("Dependencies.Dependency").Find(&u, "Id = ?", strconv.Itoa(i))

		data, _ := json.MarshalIndent(u, "", "    ")
		fmt.Println(string(data))
	}

	// db.Delete(&Model{Id: "1"})

	// fmt.Println("-----------------")

	// var cs []ModelAttribute
	// db.Find(&cs)

	// data, _ := json.MarshalIndent(cs, "", "    ")
	// fmt.Println(string(data))

}
