package api

import (
	"database/sql"

	"github.com/peterintech/briefed/internal/database"
)

type apiConfig struct {
	DB   *database.Queries
	Conn *sql.DB
}

func New(db *database.Queries, conn *sql.DB) *apiConfig {
	return &apiConfig{DB: db, Conn: conn}
}
