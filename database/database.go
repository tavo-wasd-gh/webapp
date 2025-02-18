package database

import (
	"database/sql"

	"github.com/tavo-wasd-gh/webtoolkit/logger"
	_ "github.com/mattn/go-sqlite3"
)

func Init(connDvr, connStr string) (*sql.DB, error) {
	if connDvr == "" {
		connDvr = "sqlite3"
	}

	if connStr == "" {
		connStr = "./db.db"
	}

	db, err := sql.Open(connDvr, connStr)
	if err != nil {
		return nil, logger.Errorf("error opening connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, logger.Errorf("error pinging database: %v", err)
	}

	return db, nil
}
