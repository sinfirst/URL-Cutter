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
	"go.uber.org/zap"
)

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

func (p *PGDB) GetURL(ctx context.Context, shortURL string) (string, error) {
	var origURL string
	row := p.db.QueryRow(ctx, `SELECT original_url FROM urls WHERE short_url = $1`, shortURL)
	row.Scan(&origURL)
	if origURL == "" {
		return "", fmt.Errorf("not found in storage")
	}
	return origURL, nil
}

func (p *PGDB) SetURL(ctx context.Context, shortURL, originalURL string, userID int) error {

	result, err := p.db.Exec(ctx, `INSERT INTO urls (short_url, original_url, user_id)
	 VALUES ($1, $2, $3) ON CONFLICT (short_url) DO NOTHING`, shortURL, originalURL, userID)

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

	rows, err := p.db.Query(ctx, `SELECT original_url, short_url FROM urls WHERE user_id = $1`, UserID)

	if err != nil {
		p.logger.Fatalw("Ошибка выполнения запроса %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&origURL, &shortURL)
		if err != nil {
			p.logger.Fatalw("Ошибка сканирования строки: %v", err)
			return nil, err
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
