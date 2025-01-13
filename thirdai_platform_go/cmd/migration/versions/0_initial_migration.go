package versions

import (
	"log"
	"thirdai_platform/model_bazaar/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

/*
 * In general, gorm and sqlalchemy have different names for indexes/constraints.
 * For simplicity, these migrations just delete the old indexes/constraints and
 * let gorm recreate them.
 */
func dropIndexes(model interface{}, txn *gorm.DB, indexes ...string) error {
	for _, idx := range indexes {
		if err := txn.Migrator().DropIndex(model, idx); err != nil {
			return err
		}
	}
	return nil
}

func dropConstraints(model interface{}, txn *gorm.DB, constraints ...string) error {
	for _, constraint := range constraints {
		if err := txn.Migrator().DropConstraint(model, constraint); err != nil {
			return err
		}
	}
	return nil
}

type modelTypeConversion struct {
	oldType    string
	oldSubType string
	newType    string
}

func (conv *modelTypeConversion) query() (interface{}, []interface{}) {
	if conv.oldSubType != "" {
		return "type = ? AND sub_type = ?", []interface{}{conv.oldType, conv.oldSubType}
	} else {
		return "type = ?", []interface{}{conv.oldType}
	}
}

func migrateModelTypes(txn *gorm.DB) error {
	type Model struct {
		NewType string `gorm:"size:100"`
	}

	if err := txn.Migrator().AddColumn(&Model{}, "NewType"); err != nil {
		return err
	}

	typeConversions := []modelTypeConversion{
		{oldType: "ndb", oldSubType: "", newType: schema.NdbModel},
		{oldType: "udt", oldSubType: "token", newType: schema.NlpTokenModel},
		{oldType: "udt", oldSubType: "text", newType: schema.NlpTextModel},
		{oldType: "knowledge-extraction", oldSubType: "", newType: schema.KnowledgeExtraction},
		{oldType: "enterprise-search", oldSubType: "", newType: schema.EnterpriseSearch},
	}

	for _, conv := range typeConversions {
		query, args := conv.query()
		err := txn.Model(&Model{}).Where(query, args...).Update("new_type", conv.newType).Error
		if err != nil {
			return err
		}
	}

	if err := txn.Migrator().DropColumn(&Model{}, "type"); err != nil {
		return err
	}

	if err := txn.Migrator().DropColumn(&Model{}, "sub_type"); err != nil {
		return err
	}

	if err := txn.Migrator().RenameColumn(&Model{}, "new_type", "type"); err != nil {
		return err
	}

	return nil
}

func dropModelColumns(txn *gorm.DB) error {
	type Model struct{}

	if err := txn.Migrator().DropColumn(&Model{}, "downloads"); err != nil {
		return err
	}

	if err := txn.Migrator().DropColumn(&Model{}, "domain"); err != nil {
		return err
	}

	return nil
}

func renameModelColumns(txn *gorm.DB) error {
	type Model struct{}

	if err := txn.Migrator().RenameColumn(&Model{}, "parent_id", "base_model_id"); err != nil {
		return err
	}

	if err := txn.Migrator().RenameColumn(&Model{}, "access_level", "access"); err != nil {
		return err
	}

	return nil
}

func updateModelColumnTypes(txn *gorm.DB) error {
	type Model struct {
		TrainStatus       string `gorm:"size:100;not null"`
		DeployStatus      string `gorm:"size:100;not null"`
		Access            string `gorm:"size:100;not null;default:'private'"`
		DefaultPermission string `gorm:"size:100;not null;default:'read'"`
	}

	for _, col := range []string{"train_status", "deploy_status", "access", "default_permission"} {
		if err := txn.Migrator().AlterColumn(&Model{}, col); err != nil {
			return err
		}
	}

	return nil
}

func migrateMetadataToAttribute(txn *gorm.DB) error {
	type Metadata struct {
		ModelId uuid.UUID `gorm:"primaryKey"`
		Train   string
	}

	if err := txn.Migrator().AlterColumn(&Metadata{}, "train"); err != nil {
		return err
	}

	var metadatas []Metadata
	if err := txn.Find(&metadatas).Error; err != nil {
		return err
	}

	var attrs []schema.ModelAttribute
	for _, metadata := range metadatas {
		if len(metadata.Train) > 0 {
			attrs = append(attrs, schema.ModelAttribute{ModelId: metadata.ModelId, Key: "metadata", Value: metadata.Train})
		}
	}

	if len(attrs) > 0 {
		if err := txn.Save(&attrs).Error; err != nil {
			return err
		}
	}

	return nil
}

func migrateModel(txn *gorm.DB) error {
	log.Println("migrating table 'models'")

	if err := migrateModelTypes(txn); err != nil {
		return err
	}

	if err := dropModelColumns(txn); err != nil {
		return err
	}

	if err := renameModelColumns(txn); err != nil {
		return err
	}

	if err := updateModelColumnTypes(txn); err != nil {
		return err
	}

	if err := migrateMetadataToAttribute(txn); err != nil {
		return err
	}

	type Model struct{}

	if err := dropIndexes(&Model{}, txn, "model_identifier_index", "train_status_index"); err != nil {
		return err
	}

	if err := dropConstraints(&Model{}, txn, "models_user_id_name_key", "models_parent_id_fkey", "models_team_id_fkey", "models_user_id_fkey"); err != nil {
		return err
	}

	log.Println("table 'models' migration complete")

	return nil
}

func migrateModelAttribute(txn *gorm.DB) error {
	log.Println("migrating table 'model_attributes'")

	type ModelAttribute struct{}

	if err := dropIndexes(&ModelAttribute{}, txn, "model_attribute"); err != nil {
		return err
	}

	if err := dropConstraints(&ModelAttribute{}, txn, "model_attributes_model_id_fkey"); err != nil {
		return err
	}

	log.Println("table 'model_attributes' migration complete")

	return nil
}

func migrateModelDependency(txn *gorm.DB) error {
	log.Println("migrating table 'model_dependencies'")

	type ModelDependency struct{}

	// In this table we had a unique constraint that was redundant with the primary key constraint,
	// but it caused the name of the primary key index to be different.
	if err := txn.Migrator().RenameIndex(&ModelDependency{}, "unique_model_dependency", "model_dependencies_pkey"); err != nil {
		return err
	}

	if err := dropIndexes(&ModelDependency{}, txn, "dependency_model_index", "model_dependency_index"); err != nil {
		return err
	}

	if err := dropConstraints(&ModelDependency{}, txn, "model_dependencies_dependency_id_fkey", "model_dependencies_model_id_fkey"); err != nil {
		return err
	}

	log.Println("table 'model_dependencies' migration complete")

	return nil
}

func migrateUser(txn *gorm.DB) error {
	log.Println("migrating table 'users'")

	type User struct {
		Password []byte
	}

	if err := txn.Migrator().RenameColumn(&User{}, "global_admin", "is_admin"); err != nil {
		return err
	}

	if err := txn.Migrator().DropColumn(&User{}, "verification_token"); err != nil {
		return err
	}

	if err := txn.Migrator().DropColumn(&User{}, "verified"); err != nil {
		return err
	}

	if err := txn.Migrator().RenameColumn(&User{}, "password_hash", "password"); err != nil {
		return err
	}

	// Update data type from string to bytes
	if err := txn.Migrator().AlterColumn(&User{}, "password"); err != nil {
		return err
	}

	if err := dropConstraints(&User{}, txn, "users_email_key", "users_username_key"); err != nil {
		return err
	}

	log.Println("table 'users' migration complete")

	return nil
}

func migrateTeam(txn *gorm.DB) error {
	log.Println("migrating table 'teams'")

	type Team struct{}

	if err := dropConstraints(&Team{}, txn, "teams_name_key"); err != nil {
		return err
	}

	log.Println("table 'teams' migration complete")

	return nil
}

func migrateUserTeam(txn *gorm.DB) error {
	log.Println("migrating table 'user_teams'")

	type UserTeam struct {
		IsTeamAdmin bool
	}

	if err := txn.Migrator().AddColumn(&UserTeam{}, "IsTeamAdmin"); err != nil {
		return err
	}

	err := txn.Model(&UserTeam{}).Where("role = ?", "team_admin").Update("is_team_admin", true).Error
	if err != nil {
		return err
	}

	if err := txn.Migrator().DropColumn(&UserTeam{}, "role"); err != nil {
		return err
	}

	if err := dropConstraints(&UserTeam{}, txn, "user_teams_team_id_fkey", "user_teams_user_id_fkey"); err != nil {
		return err
	}

	log.Println("table 'user_teams' migration complete")

	return nil
}

func dropUnusedTables(txn *gorm.DB) error {
	tables := []string{"password_resets", "model_permissions", "metadata", "catalog", "job_messages"}
	for _, table := range tables {
		err := txn.Migrator().DropTable(table)
		if err != nil {
			return err
		}
	}

	return nil
}

func Migration_0_initial_migration(txn *gorm.DB) error {
	log.Println("performing inital migration to new backend schema")

	if err := migrateTeam(txn); err != nil {
		return err
	}

	if err := migrateUser(txn); err != nil {
		return err
	}

	if err := migrateUserTeam(txn); err != nil {
		return err
	}

	if err := migrateModelAttribute(txn); err != nil {
		return err
	}

	if err := migrateModelDependency(txn); err != nil {
		return err
	}

	if err := migrateModel(txn); err != nil {
		return err
	}

	if err := dropUnusedTables(txn); err != nil {
		return err
	}

	// TODO(Nicholas): replace with final structs
	err := txn.Migrator().AutoMigrate(
		&schema.Model{}, &schema.ModelAttribute{}, &schema.ModelDependency{},
		&schema.User{}, &schema.Team{}, &schema.UserTeam{}, &schema.JobLog{},
		&schema.Upload{},
	)
	if err != nil {
		return err
	}

	log.Println("initial migration to new backend schema complete")

	return nil

}
