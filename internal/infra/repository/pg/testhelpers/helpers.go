//go:build test

package testhelpers

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

// ResetDB truncates all tables to ensure clean state between tests
func ResetDB(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{
		"sessions",
		"user_providers",
		"user_profiles",
		"urls",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec("TRUNCATE TABLE " + table + " CASCADE")
		require.NoError(t, err, "failed to truncate table %s", table)
	}
}
