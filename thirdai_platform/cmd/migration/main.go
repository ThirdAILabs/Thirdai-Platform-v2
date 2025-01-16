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

func isFirstMigrationFromOldBackend(db *gorm.DB) bool {
	fromOldBackend := db.Migrator().HasTable("alembic_version")
	isFirstMigration := !db.Migrator().HasTable(gormigrate.DefaultOptions.TableName)
	return fromOldBackend && isFirstMigration
}

func main() {
	dbUri := flag.String("db_uri", "", "Database URI")
	printLatestVersion := flag.Bool("print_latest", false, "Just print out the latest version and return")
	rollbackTo := flag.String("rollback_to", "", "Instead of updating the schema, this flag indicates that it should be rolled back to the given version.")
	flag.Parse()

	migrations := []*gormigrate.Migration{
		{
			// This is a placeholder to represent the state from the previous backend db schema.
			// The reason for this is that gormigrate looks for a some migration entry in the
			// migrations table to indicate that it is not a clean DB. Having this placeholder
			// means that we can distinguish between having a clean DB, or a DB that has been
			// initialized with the old backend, and thus  needs to be migrated, rather than
			// initialized from scratch.
			ID: "PLACEHOLDER",
			Migrate: func(*gorm.DB) error {
				log.Println("running placeholder migration")
				return nil
			},
		},
		{
			ID:      "0",
			Migrate: versions.Migration_0_initial_migration,
			// Rollback is not supported for this migration since the migration is more
			// complicated and not intended to be reversed
		},
	}

	if *printLatestVersion {
		fmt.Println(migrations[len(migrations)-1].ID)
		return
	}

	db, err := gorm.Open(postgres.Open(postgresDsn(*dbUri)), &gorm.Config{})
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	migrator := gormigrate.New(db, gormigrate.DefaultOptions, migrations)

	if isFirstMigrationFromOldBackend(db) {
		// This needs to be done before specifying the InitSchema option, becuase otherwise
		// when gormigrate detects that no migration has ran, it will attemp to run the
		// InitSchema, which is incorrect here because the db is initialized, just not
		// by gormigrate.
		log.Println("migration is detected as the first migration from the old backend schema")
		if err := migrator.MigrateTo("PLACEHOLDER"); err != nil {
			log.Fatalf("unable to perform placeholder migration: %v", err)
		}
	}

	migrator.InitSchema(func(txn *gorm.DB) error {
		log.Println("clean database detected, running full schema initialization")

		return db.AutoMigrate(
			&schema.Model{}, &schema.ModelAttribute{}, &schema.ModelDependency{},
			&schema.User{}, &schema.Team{}, &schema.UserTeam{}, &schema.JobLog{},
			&schema.Upload{},
		)
	})

	if *rollbackTo != "" {
		if err := migrator.RollbackTo(*rollbackTo); err != nil {
			log.Fatalf("rollback failed: %v", err)
		}

		log.Println("rollback completed successfully")
	} else {
		if err := migrator.Migrate(); err != nil {
			log.Fatalf("migration failed: %v", err)
		}

		log.Println("migration completed successfully")
	}
}
