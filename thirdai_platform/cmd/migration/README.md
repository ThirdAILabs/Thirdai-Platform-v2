# Migrations


## Running Migrations

* To update the schema just run `go run cmd/migrations/main.go --db_uri <db uri>`. This will update the schema to the latest, including conversion from the old backend schema to the new schema.
* To undo a migration just run `go run cmd/migrations/main.go --db_uri <db uri> --rollback_last`. This will undo the last migration. 

## Adding New Migrations

Adding new migrations is simple, just add a new entry in the list in `main.go`. The migration library we are using has a good example here (https://github.com/go-gormigrate/gormigrate). 

The entry can contain 3 fields: 
* `ID`: used to identify the migration, this needs to be unique among the migrations. 
* `Migrate`: the function to apply the migration to update the schema.
* `Rollback`: the function to reverse the migration. This is optional, if not specified then the migration will just give an error if tried to run in reverse. 