package testutils

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

// DatabaseContainer holds a reference to a container used for integration
// testing.
type DatabaseContainer struct {
	Container  testcontainers.Container
	ConnString string
}

// GetCrdbDatabaseContainer creates a DatabaseContainer reference for
// testing with CockroachDB.
func GetCrdbDatabaseContainer(ctx context.Context) (*DatabaseContainer, error) {
	crdbContainer, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:v25.4.1",
		cockroachdb.WithInsecure(),
		cockroachdb.WithUser("root"),
		cockroachdb.WithDatabase("notifications"),
	)
	if err != nil {
		return nil, err
	}

	connString, err := crdbContainer.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	return &DatabaseContainer{
		Container:  crdbContainer,
		ConnString: connString,
	}, nil

}

// Migrate performs migration on a new database instance. Intended to be
// used after creating a new DatabaseContainer to match the production
// database.
func Migrate(ctx context.Context, databaseContainer *DatabaseContainer) error {
	migrationPath, err := getMigrationAbsolutePath("../../")
	if err != nil {
		return err
	}

	m, err := migrate.New(migrationPath, databaseContainer.ConnString)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		return err
	}

	return nil
}

func getMigrationAbsolutePath(relativePathToProjectRoot string) (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}

	migrationDir := filepath.Join(filepath.Dir(filename), relativePathToProjectRoot)

	absPath, err := filepath.Abs(migrationDir)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

func getSourceURL(relativePathToProjectRoot string) (string, error) {
	absPath, err := getMigrationAbsolutePath(relativePathToProjectRoot)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("file://%s", absPath), nil
}
