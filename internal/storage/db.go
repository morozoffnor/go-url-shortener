package storage

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/pkg/chargen"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Database struct {
	conn *pgxpool.Pool
	cfg  *config.Config
}

func NewDatabase(cfg *config.Config, ctx context.Context) *Database {
	db := &Database{
		cfg: cfg,
	}
	conn, err := pgxpool.New(ctx, cfg.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	//conn, err := sql.Open("pgx", cfg.DatabaseDSN)
	//if err != nil {
	//	panic(err)
	//}
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
	if err := d.conn.Ping(ctx); err != nil {
		return false
	}
	return true
}

func (d *Database) createTable(ctx context.Context) error {
	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return err
	}

	query := `CREATE TABLE IF NOT EXISTS "urls"(
    	id varchar(255) PRIMARY KEY,
    	full_url varchar(500) UNIQUE NOT NULL,
    	short_url varchar(255) UNIQUE NOT NULL
	);`

	_, err = tx.Exec(ctx, query)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (d *Database) AddNewURL(ctx context.Context, fullURL string) (string, error) {
	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return "", err
	}
	shortURL := chargen.CreateRandomCharSeq()
	id := uuid.NewString()

	query := `INSERT INTO urls (id, full_url, short_url) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, query, id, fullURL, shortURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.Print("URL already exists")
			short, _ := d.getShortURL(ctx, fullURL)
			return short, pgErr
		}
		return "", err
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Print(err)
		return "", err
	}
	return shortURL, nil
}

func (d *Database) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	var fullURL string
	query := `SELECT full_url FROM urls WHERE short_url=$1`
	err := d.conn.QueryRow(ctx, query, shortURL).Scan(&fullURL)
	if err != nil {
		return "", err
	}
	return fullURL, nil
}

func (d *Database) getShortURL(ctx context.Context, fullURL string) (string, error) {
	var shortURL string
	query := `SELECT short_url FROM urls WHERE full_url=$1`
	err := d.conn.QueryRow(ctx, query, fullURL).Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (d *Database) AddBatch(ctx context.Context, urls []BatchInput) ([]BatchOutput, error) {
	if len(urls) < 1 {
		return []BatchOutput{}, nil
	}

	batch := &pgx.Batch{}
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

		batch.Queue("INSERT INTO urls (id, full_url, short_url) VALUES ($1, $2, $3)", id, v.OriginalURL, shortURL)

		result = append(result, BatchOutput{
			ShortURL:      d.cfg.ResultAddr + "/" + shortURL,
			CorrelationID: v.CorrelationID,
		})
	}
	br := d.conn.SendBatch(ctx, batch)
	defer br.Close()
	return result, nil

}
