package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/pkg/chargen"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Database struct {
	conn *sql.DB
	cfg  *config.Config
}

func NewDatabase(cfg *config.Config) *Database {
	db := &Database{
		cfg: cfg,
	}
	conn, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	db.conn = conn
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.createTable(ctx); err != nil {
		panic(err)
	}
	return db
}

func (d *Database) Ping(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if err := d.conn.PingContext(ctx); err != nil {
		return false
	}
	return true
}

func (d *Database) createTable(ctx context.Context) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}

	query := `CREATE TABLE IF NOT EXISTS "urls"(
    	id varchar(255) PRIMARY KEY,
    	full_url varchar(500) UNIQUE NOT NULL,
    	short_url varchar(255) UNIQUE NOT NULL
	);`

	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (d *Database) AddNewURL(ctx context.Context, fullURL string) (string, error) {
	tx, err := d.conn.Begin()
	if err != nil {
		return "", err
	}
	shortURL := chargen.CreateRandomCharSeq()
	id := uuid.NewString()

	query := `INSERT INTO urls (id, full_url, short_url) VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, query, id, fullURL, shortURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.Print("URL already exists")
			return "", err
		}
		return "", err
	}
	err = tx.Commit()
	if err != nil {
		log.Print(err)
		return "", err
	}
	return shortURL, nil
}

func (d *Database) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	var fullURL string
	query := `SELECT full_url FROM urls WHERE short_url=$1`
	err := d.conn.QueryRowContext(ctx, query, shortURL).Scan(&fullURL)
	if err != nil {
		return "", err
	}
	return fullURL, nil
}

func (d *Database) getShortURL(ctx context.Context, fullURL string) (string, error) {
	var shortURL string
	query := `SELECT short_url FROM urls WHERE full_url=$1`
	err := d.conn.QueryRowContext(ctx, query, fullURL).Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (d *Database) AddBatch(ctx context.Context, urls []BatchInput) ([]BatchOutput, error) {
	if len(urls) < 1 {
		return []BatchOutput{}, nil
	}

	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	var result []BatchOutput
	for _, v := range urls {
		if short, _ := d.getShortURL(ctx, v.OriginalURL); short != "" {
			result = append(result, BatchOutput{
				ShortURL:      d.cfg.ResultAddr + "/" + short,
				CorrelationID: v.CorrelationID,
			})
			continue
		}
		shortURL := chargen.CreateRandomCharSeq()
		id := uuid.NewString()
		query := `INSERT INTO urls (id, full_url, short_url) VALUES ($1, $2, $3)`
		_, err = tx.ExecContext(ctx, query, id, v.OriginalURL, shortURL)
		if err != nil {
			log.Print(err)
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				log.Print(rollbackErr)
				return nil, rollbackErr
			}
			return nil, err
		}
		result = append(result, BatchOutput{
			ShortURL:      d.cfg.ResultAddr + "/" + shortURL,
			CorrelationID: v.CorrelationID,
		})
	}
	err = tx.Commit()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return result, nil

}
