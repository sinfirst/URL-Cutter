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
	config config.Config
	logger zap.SugaredLogger
	db     *pgxpool.Pool
	ctx    context.Context
}

func NewPGDB(config config.Config, logger zap.SugaredLogger) *PGDB {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, config.DatabaseDsn)

	if err != nil {
		logger.Errorw("Problem with connecting to db ", err)
		return nil
	}
	return &PGDB{config: config, logger: logger, db: db, ctx: ctx}
}

func (p *PGDB) ConnectToDB() (*pgxpool.Pool, error) {
	db, err := pgxpool.New(p.ctx, p.config.DatabaseDsn) //sql.Open("pgx", p.config.DatabaseDsn)

	if err != nil {
		p.logger.Errorw("Problem with connecting to db ", err)
		return nil, err
	}

	err = db.Ping(p.ctx)

	if err != nil {
		p.logger.Errorw("Problem with ping to db ", err)
		return nil, err
	}

	p.logger.Infow("Connecting and ping to db: OK")
	return db, nil
}

func (p *PGDB) GetFromStorage(shortURL string) (string, error) {
	var origURL string
	row := p.db.QueryRow(p.ctx, `SELECT original_url FROM urls WHERE short_url = $1`, shortURL)
	row.Scan(&origURL)
	if origURL == "" {
		return "", fmt.Errorf("not found in storage")
	}
	return origURL, nil
}

func (p *PGDB) SetInStorage(shortURL, originalURL string) error {

	result, err := p.db.Exec(p.ctx, `INSERT INTO urls (short_url, original_url) 
	VALUES ($1, $2) ON CONFLICT (short_url) DO NOTHING`, shortURL, originalURL)

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

func InitMigrations(conf config.Config, logger zap.SugaredLogger) {
	logger.Infow("Start migrations")
	db, err := sql.Open("pgx", conf.DatabaseDsn)

	if err != nil {
		logger.Errorw("Error with connection to DB: ", err)
	}

	defer db.Close()

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "migrations")

	err = goose.Up(db, migrationsPath)
	if err != nil {
		logger.Errorw("Error with migrations: ", err)
	}
}
