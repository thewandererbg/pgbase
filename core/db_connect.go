package core

import (
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pocketbase/dbx"
)

func DefaultDBConnect(dbPath string) (*dbx.DB, error) {
	godotenv.Load()

	var dbURI string
	if strings.Contains(dbPath, "data.db") {
		dbURI = os.Getenv("PB_DATA_URI")
	} else {
		dbURI = os.Getenv("PB_AUX_URI")
	}

	db, err := dbx.Open("pgx", dbURI)
	if err != nil {
		return nil, err
	}

	return db, nil
}
