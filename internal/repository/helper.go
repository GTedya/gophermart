package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

func NewDB(dsn string) (*sql.DB, error) {
	dbConn, err := sql.Open(`postgres`, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to database: %w", err)
	}
	err = dbConn.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return dbConn, err
}
