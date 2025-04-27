package postgresbd

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/sinfirst/URL-Cutter/internal/app/config"
	"go.uber.org/zap"
)

type ShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenResponce struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type PGDB struct {
	logger zap.SugaredLogger
	db     *pgxpool.Pool
	config config.Config
}

func NewPGDB(config config.Config, logger zap.SugaredLogger) *PGDB {
	db, err := pgxpool.New(context.Background(), config.DatabaseDsn)

	if err != nil {
		logger.Errorw("Problem with connecting to db ", err)
		return nil
	}
	return &PGDB{logger: logger, db: db, config: config}
}

func (p *PGDB) Ping(ctx context.Context) error {
	err := p.db.Ping(ctx)

	if err != nil {
		p.logger.Errorw("Problem with ping to db: ", err)
		return err
	}

	return nil
}

func (p *PGDB) BatchShortenURL(ctx context.Context, requests []ShortenRequest) ([]ShortenResponce, error) {
	tx, err := p.db.Begin(ctx)

	if err != nil {
		p.logger.Errorw("Failed to start transaction", err)
		return nil, err
	}

	defer tx.Rollback(ctx)

	query := `INSERT INTO urls (short_url, original_url, user_id)
	 VALUES ($1, $2, $3) ON CONFLICT (short_url) DO NOTHING`

	var responces []ShortenResponce

	for _, req := range requests {
		shortURL := fmt.Sprintf("%x", md5.Sum([]byte(req.OriginalURL)))[:8]

		_, err := tx.Exec(ctx, query, shortURL, req.OriginalURL, 0)

		if err != nil {
			p.logger.Errorw("Failed to insert data", err)
			return nil, err
		}

		responces = append(responces, ShortenResponce{
			CorrelationID: req.CorrelationID,
			ShortURL:      p.config.Host + "/" + shortURL,
		})

	}

	if err := tx.Commit(ctx); err != nil {
		p.logger.Errorw("Failed to commit transaction", ctx)
		return nil, err
	}

	return responces, nil
}

func (p *PGDB) UpdateDeleteParam(ctx context.Context, shortURLs string) {
	query := `UPDATE urls
				SET is_deleted = TRUE
				WHERE short_url = $1`

	_, err := p.db.Exec(ctx, query, shortURLs)
	if err != nil {
		p.logger.Errorw("Update table error: ", err)
		return
	}
}

func (p *PGDB) Delete(ctx context.Context, shortURLs string) {
	query := `DELETE FROM urls
				WHERE short_url = $1`

	_, err := p.db.Exec(context.Background(), query, shortURLs)

	if err != nil {
		p.logger.Errorw("Problem with deleting from db: ", err)
		return
	}
}

func (p *PGDB) GetURL(ctx context.Context, shortURL string) (string, error) {
	var origURL string
	var isDelete bool

	query := `SELECT original_url, is_deleted FROM urls WHERE short_url = $1`
	row := p.db.QueryRow(context.Background(), query, shortURL)
	row.Scan(&origURL, &isDelete)
	if isDelete {
		return "", fmt.Errorf("deleted")
	}
	if origURL == "" {
		return "", fmt.Errorf("not found in storage")
	}

	return origURL, nil
}

func (p *PGDB) SetURL(ctx context.Context, shortURL, originalURL string, userID int) error {
	query := `INSERT INTO urls (short_url, original_url, user_id)
	 VALUES ($1, $2, $3) ON CONFLICT (short_url) DO NOTHING`

	result, err := p.db.Exec(ctx, query, shortURL, originalURL, userID)

	if rows := result.RowsAffected(); rows == 0 {
		return err
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return err
		}
	}

	return nil
}

func (p *PGDB) GetWithUserID(ctx context.Context, UserID int) (map[string]string, error) {
	var origURL string
	var shortURL string
	URLs := make(map[string]string)

	query := `SELECT original_url, short_url FROM urls WHERE user_id = $1`
	rows, err := p.db.Query(context.Background(), query, UserID)

	if err != nil {
		p.logger.Fatalw("Ошибка выполнения запроса %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&origURL, &shortURL)
		if err != nil {
			p.logger.Fatalw("Ошибка сканирования строки: %v", err)
		}

		URLs[shortURL] = origURL
	}

	if origURL == "" {
		return nil, fmt.Errorf("original url is empty")
	}
	return URLs, nil
}

func InitMigrations(conf config.Config, logger zap.SugaredLogger) {
	logger.Infow("Start migrations")
	db, err := sql.Open("pgx", conf.DatabaseDsn)

	if err != nil {
		logger.Fatalw("Error with connection to DB: ", err)
		return
	}

	defer db.Close()

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "migrations")

	err = goose.Up(db, migrationsPath)
	if err != nil {
		logger.Fatalw("Error with migrations: ", err)
		return
	}
}
