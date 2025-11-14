package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type PostgresDb struct {
	db     *sql.DB
	config *Config
}

func New(configPath string) (*PostgresDb, error) {
	conf, err := NewConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("init config %w", err)
	}
	pg := &PostgresDb{db: nil, config: conf}

	pg.db, err = sql.Open("postgres", pg.config.dataSourceName())

	if err != nil {
		return pg, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := pg.db.Ping(); err != nil {
		return pg, fmt.Errorf("failed to connect to database: %v", err)
	}

	return pg, nil
}
