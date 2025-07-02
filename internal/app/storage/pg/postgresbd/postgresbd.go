// Package postgresbd пакет с описанием работы с базой данных
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

// PGDB структура для хранения переменных
type PGDB struct {
	logger zap.SugaredLogger
	db     *pgxpool.Pool
}

// NewPGDB конструктор для структуры
func NewPGDB(config config.Config, logger zap.SugaredLogger) *PGDB {
	db, err := pgxpool.New(context.Background(), config.DatabaseDsn)

	if err != nil {
		logger.Errorw("Problem with connecting to db ", err)
		return nil
	}
	return &PGDB{logger: logger, db: db}
}

// Ping проверка соеденения с бд
func (p *PGDB) Ping(ctx context.Context) error {
	err := p.db.Ping(ctx)

	if err != nil {
		p.logger.Errorw("Problem with ping to db: ", err)
		return err
	}

	return nil
}

// DeleteURL функция для удаления урлов из бд
func (p *PGDB) DeleteURL(ctx context.Context, shortURL string) error {
	query := `DELETE FROM urls
				WHERE short_url = $1`

	_, err := p.db.Exec(ctx, query, shortURL)

	if err != nil {
		p.logger.Errorw("Problem with deleting from db: ", err)
		return err
	}
	return nil
}

// GetURL получение данных из бд
func (p *PGDB) GetURL(ctx context.Context, shortURL string) (string, error) {
	var origURL string

	query := `SELECT original_url FROM urls WHERE short_url = $1`
	row := p.db.QueryRow(ctx, query, shortURL)
	err := row.Scan(&origURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("not found in storage")
		}
		p.logger.Infow("problem with scan", "error", err)
		return "", err
	}
	return origURL, nil
}

// SetURL сохранить урл в бд
func (p *PGDB) SetURL(ctx context.Context, shortURL, originalURL string, userID int) error {
	query := `INSERT INTO urls (short_url, original_url, user_id)
	 VALUES ($1, $2, $3)`

	_, err := p.db.Exec(ctx, query, shortURL, originalURL, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return err
		}
	}
	return nil
}

// GetByUserID получить урлы по userID
func (p *PGDB) GetByUserID(ctx context.Context, userID int) ([]models.ShortenOrigURLs, error) {
	var origURL string
	var shortURL string
	var urls []models.ShortenOrigURLs

	query := `SELECT original_url, short_url FROM urls WHERE user_id = $1`
	rows, err := p.db.Query(ctx, query, userID)

	if err != nil {
		p.logger.Errorw("Ошибка выполнения запроса %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&origURL, &shortURL)
		if err != nil {
			p.logger.Errorw("Ошибка сканирования строки: %v", err)
		}

		urls = append(urls, models.ShortenOrigURLs{OriginalURL: origURL, ShortURL: shortURL})
	}

	if origURL == "" {
		return nil, fmt.Errorf("original url is empty")
	}
	return urls, nil
}

// InitMigrations инициализация миграций
func InitMigrations(conf config.Config, logger zap.SugaredLogger) error {
	logger.Infow("Start migrations")
	db, err := sql.Open("pgx", conf.DatabaseDsn)

	if err != nil {
		logger.Errorw("Error with connection to DB: ", err)
		return err
	}

	defer db.Close()

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "migrations")

	err = goose.Up(db, migrationsPath)
	if err != nil {
		logger.Errorw("Error with migrations: ", err)
		return err
	}
	return nil
}
