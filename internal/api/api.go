package api

import "github.com/peterintech/briefed/internal/database"

type apiConfig struct {
	DB *database.Queries
}

func New(db *database.Queries) *apiConfig {
	return &apiConfig{
		DB: db,
	}
}
