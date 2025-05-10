package postgresbd

import (
	"context"
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
	"github.com/sinfirst/URL-Cutter/internal/app/models"
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
}

func NewPGDB(config config.Config, logger zap.SugaredLogger) *PGDB {
	db, err := pgxpool.New(context.Background(), config.DatabaseDsn)

	if err != nil {
		logger.Errorw("Problem with connecting to db ", err)
		return nil
	}
	return &PGDB{logger: logger, db: db}
}

func (p *PGDB) Ping(ctx context.Context) error {
	err := p.db.Ping(ctx)

	if err != nil {
		p.logger.Errorw("Problem with ping to db: ", err)
		return err
	}

	return nil
}

func (p *PGDB) DeleteURL(ctx context.Context, shortURL string) {
	query := `UPDATE urls
				SET is_deleted = TRUE
				WHERE short_url = $1`

	_, err := p.db.Exec(ctx, query, shortURL)
	if err != nil {
		p.logger.Errorw("Update table error: ", err)
		return
	}
}

func (p *PGDB) GetURL(ctx context.Context, shortURL string) (string, error) {
	var origURL string
	var isDeleted bool

	query := `SELECT original_url, is_deleted FROM urls WHERE short_url = $1`
	row := p.db.QueryRow(ctx, query, shortURL)
	row.Scan(&origURL, &isDeleted)
	if origURL == "" || isDeleted {
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

func (p *PGDB) GetByUserID(ctx context.Context, userID int) ([]models.ShortenOrigURLs, error) {
	var origURL string
	var shortURL string
	var URLs []models.ShortenOrigURLs

	query := `SELECT original_url, short_url FROM urls WHERE user_id = $1`
	rows, err := p.db.Query(context.Background(), query, userID)

	if err != nil {
		p.logger.Fatalw("Ошибка выполнения запроса %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&origURL, &shortURL)
		if err != nil {
			p.logger.Fatalw("Ошибка сканирования строки: %v", err)
		}

		URLs = append(URLs, models.ShortenOrigURLs{OriginalURL: origURL, ShortURL: shortURL})
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
