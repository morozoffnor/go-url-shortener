package storage

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/morozoffnor/go-url-shortener/internal/auth"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/pkg/chargen"
	"github.com/morozoffnor/go-url-shortener/pkg/logger"
	"log"
	"sync"
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
	conn.Config().MaxConns = 20
	conn.Config().MinConns = 2
	db.conn = conn
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	doMigrations(cfg)
	//err = db.createTable(ctx)
	//if err != nil {
	//	panic(err)
	//}
	return db
}

func doMigrations(cfg *config.Config) {
	m, err := migrate.New("file://internal/storage/migrations", cfg.DatabaseDSN)

	if err != nil {
		panic(err)
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}
}

func (d *Database) Ping(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if err := d.conn.Ping(ctx); err != nil {
		return false
	}
	return true
}

// Оставлю тут на случай, если автотесты будут ругаться на создание таблицы
func (d *Database) createTable(ctx context.Context) error {
	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return err
	}

	query := `CREATE TABLE IF NOT EXISTS "urls"(
   	id varchar(255) PRIMARY KEY,
   	full_url varchar(500) UNIQUE NOT NULL,
   	short_url varchar(255) UNIQUE NOT NULL,
    user_id varchar(255) NOT NULL,
    is_deleted boolean NOT NULL 
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

	query := `INSERT INTO urls (id, full_url, short_url, user_id) VALUES ($1, $2, $3, $4)`
	_, err = tx.Exec(ctx, query, id, fullURL, shortURL, ctx.Value(auth.ContextUserID))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.Print("URL already exists")
			short, _ := d.getShortURL(ctx, fullURL)
			_ = tx.Commit(ctx)
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

func (d *Database) GetFullURL(ctx context.Context, shortURL string) (string, bool, error) {
	var fullURL string
	var isDeleted bool
	query := `SELECT full_url, is_deleted FROM urls WHERE short_url=$1`
	err := d.conn.QueryRow(ctx, query, shortURL).Scan(&fullURL, &isDeleted)
	if err != nil {
		return "", false, err
	}
	return fullURL, isDeleted, nil
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

		batch.Queue("INSERT INTO urls (id, full_url, short_url, user_id) VALUES ($1, $2, $3, $4)", id, v.OriginalURL, shortURL, ctx.Value(auth.ContextUserID))

		result = append(result, BatchOutput{
			ShortURL:      d.cfg.ResultAddr + "/" + shortURL,
			CorrelationID: v.CorrelationID,
		})
	}
	br := d.conn.SendBatch(ctx, batch)
	defer br.Close()
	return result, nil

}

func (d *Database) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]UserURLs, error) {
	if len(userID) == 0 {
		return nil, nil
	}

	var result []UserURLs
	rows, err := d.conn.Query(ctx, "SELECT short_url, full_url FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var row UserURLs
		err := rows.Scan(&row.ShortURL, &row.OriginalURL)
		if err != nil {
			return nil, err
		}
		row.ShortURL = d.cfg.ResultAddr + "/" + row.ShortURL

		result = append(result, row)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, err
}

func (d *Database) DeleteURLs(ctx context.Context, userID uuid.UUID, urls URLsForDeletion) {
	input := d.generator(ctx, userID, urls)
	out := d.fanOut(ctx, input)
	in := d.fanIn(ctx, out)
	d.softDeleteURLs(ctx, in)
}

func (d *Database) generator(ctx context.Context, userID uuid.UUID, urls URLsForDeletion) chan DeleteURLItem {
	inputCh := make(chan DeleteURLItem)

	// наполняем канал айтемами
	go func() {
		defer close(inputCh)
		for _, v := range urls {
			item := DeleteURLItem{
				UserID:   userID,
				ShortURL: v,
			}
			log.Print("gen", item)
			select {
			case <-ctx.Done():
				return
			case inputCh <- item:
			}
		}
	}()
	return inputCh
}

func (d *Database) fanOut(ctx context.Context, inputCh <-chan DeleteURLItem) chan string {
	outCh := make(chan string)

	// распределяем работу: ищем айдишники урлов в базе
	go func() {
		defer close(outCh)

		for item := range inputCh {
			var id string
			row := d.conn.QueryRow(ctx, "SELECT id FROM public.urls WHERE short_url = $1 AND user_id = $2", item.ShortURL, item.UserID)
			err := row.Scan(&id)
			if err != nil {
				logger.Logger.Error(err)
				continue
			}
			log.Print("fanOut", "sent id")
			select {
			case <-ctx.Done():
				return
			case outCh <- id:
			}
		}
	}()

	return outCh
}

func (d *Database) fanIn(ctx context.Context, ids ...chan string) chan string {
	delCh := make(chan string)

	var wg sync.WaitGroup

	// собираем полученные айдишники в один канал
	for _, ch := range ids {
		wg.Add(1)
		log.Print("fanIn", " collected id")
		go func() {
			defer wg.Done()

			for item := range ch {
				select {
				case <-ctx.Done():
					return
				case delCh <- item:
				}
			}
		}()
	}

	// ждём выполнения и закрываем канал
	go func() {
		wg.Wait()
		close(delCh)
	}()

	return delCh
}

func (d *Database) softDeleteURLs(ctx context.Context, delCh chan string) {
	var idsForDeletion []string
	for item := range delCh {
		idsForDeletion = append(idsForDeletion, item)
	}

	if len(idsForDeletion) == 0 {
		return
	}

	batch := &pgx.Batch{}
	for _, item := range idsForDeletion {
		log.Print("batch ", item)
		batch.Queue("UPDATE urls SET is_deleted = true WHERE id = $1", item)
	}

	br := d.conn.SendBatch(ctx, batch)
	defer br.Close()
}
