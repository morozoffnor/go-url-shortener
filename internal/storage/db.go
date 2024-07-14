package storage

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Database struct {
	connString string
}

func TestConnection(ps string) bool {
	//ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable port=%s",
	//	`localhost`, `url`, `134562`, `url`, `5433`)

	db, err := sql.Open("pgx", ps)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return false
	}
	return true
}
