package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
	"thirdai_platform/cmd/migration/versions"
	"thirdai_platform/model_bazaar/schema"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func postgresDsn(uri string) string {
	if uri == "" {
		log.Fatalf("Missing --db_uri arg")
	}
	parts, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("error parsing db uri: %v", err)
	}
	pwd, _ := parts.User.Password()
	dbname := strings.TrimPrefix(parts.Path, "/")
	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v", parts.Hostname(), parts.User.Username(), pwd, dbname, parts.Port())
}

func main() {
	dbUri := flag.String("db_uri", "", "Database URI")
	flag.Parse()

	db, err := gorm.Open(postgres.Open(postgresDsn(*dbUri)), &gorm.Config{})
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	migration := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			// This is a placeholder to represent the state from the previous backend db schema.
			ID:      "0",
			Migrate: func(*gorm.DB) error { return nil },
		},
		{
			ID:      "1",
			Migrate: versions.Migration_1_initial_migration,
			// Rollback is not supported for this migration since the migration is more
			// complicated and not intended to be reversed
		},
	})

	migration.InitSchema(func(txn *gorm.DB) error {
		log.Println("clean database detected, running full schema initialization")

		return db.AutoMigrate(
			&schema.Model{}, &schema.ModelAttribute{}, &schema.ModelDependency{},
			&schema.User{}, &schema.Team{}, &schema.UserTeam{}, &schema.JobLog{},
		)
	})

	if err := migration.Migrate(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migration completed successfully")
}
