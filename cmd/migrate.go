package migrate

import (
	sql "database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"

	"vilmasoftware.com/colablists/pkg/infra"
)

const migrationsFolder = "migrations"

type MigrationResult struct {
	RanMigrations  []string
	Error          error
	MigrationError string
}

func NewMigrationResultSuccess(migrations []string) MigrationResult {
	return MigrationResult{
		RanMigrations: migrations,
	}
}

func NewMigrationResultError(err error, filename string) MigrationResult {
	return MigrationResult{
		Error:          err,
		MigrationError: filename,
	}
}

func IsFirstMigration() bool {
	info, err := os.Stat(infra.GetDatabaseUrl())
	if os.IsNotExist(err) {
		return true
	}
	return info.Size() == 0
}
func MigrateDb() MigrationResult {
	var appliedMigrations []string
	if IsFirstMigration() {
		appliedMigrations = make([]string, 0)
	} else {
		appliedMigrations = listAppliedMigrations()
	}
	migrations := listMigrations()
	slices.Sort(appliedMigrations)
	log.Printf("Applied migrations: %v", appliedMigrations)
	log.Printf("Migrations: %v", migrations)
	slices.Sort(migrations)
	migrationsIdx, appliedMigrationsIdx := 0, 0
	toApply := make([]string, 0)
	for migrationsIdx < len(migrations) && appliedMigrationsIdx < len(appliedMigrations) {
		if migrations[migrationsIdx] == appliedMigrations[appliedMigrationsIdx] {
			migrationsIdx++
			appliedMigrationsIdx++
		} else {
			if appliedMigrationsIdx < len(appliedMigrations) {
				return MigrationResult{
					Error:          fmt.Errorf("migration history does not match migration files. Missing %v", appliedMigrations[appliedMigrationsIdx]),
					MigrationError: appliedMigrations[appliedMigrationsIdx],
				}
			}
			toApply = append(toApply, migrations[migrationsIdx])
			migrationsIdx++
		}
	}
	if len(appliedMigrations) != 0 && appliedMigrationsIdx < len(appliedMigrations) {
		return MigrationResult{
			Error:          fmt.Errorf("folder %v is missing applied migration %v", migrationsFolder, appliedMigrations[appliedMigrationsIdx]),
			MigrationError: appliedMigrations[appliedMigrationsIdx],
		}
	}
	for migrationsIdx < len(migrations) {
		toApply = append(toApply, migrations[migrationsIdx])
		migrationsIdx++
	}
	sql, err := infra.CreateConnection()
	if err != nil {
		return MigrationResult{
			Error: err,
		}
	}
	defer sql.Close()
	tx, err := sql.Begin()
	if err != nil {
		return MigrationResult{
			Error: err,
		}
	}
	defer tx.Rollback()
	fmt.Printf("Applying migrations: %v\n", toApply)
	for _, migration := range toApply {
		err := executeMigration(tx, migration)
		if err != nil {
			return MigrationResult{
				Error: err,
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return MigrationResult{
			Error: err,
		}
	}
	return MigrationResult{RanMigrations: toApply}
}

// Executes a migration file adding it to migrations history.
func executeMigration(tx *sql.Tx, filename string) error {
	sqlBytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	_, err = tx.Exec(string(sqlBytes))
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO migrations (filename) VALUES (?)", filename)
	if err != nil {
		return err
	}
	return nil
}

func listMigrations() []string {
	files, err := os.ReadDir(migrationsFolder)
	if err != nil {
		panic(err)
	}
	migrations := make([]string, 0)
	for _, file := range files {
		migrations = append(migrations, filepath.Join(migrationsFolder, file.Name()))
	}
	return migrations
}

func listAppliedMigrations() []string {
	conn, err := infra.CreateConnection()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	rows, err := conn.Query("SELECT filename FROM migrations")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	migrations := make([]string, 0)
	for rows.Next() {
		var migration string
		err := rows.Scan(&migration)
		if err != nil {
			panic(err)
		}
		migrations = append(migrations, migration)
	}
	return migrations
}
