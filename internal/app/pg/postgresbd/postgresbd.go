package postgresbd

import (
	"context"
	"database/sql"
	"errors"
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
}

func NewPGDB(config config.Config, logger zap.SugaredLogger) *PGDB {
	db, err := pgxpool.New(context.Background(), config.DatabaseDsn)

	if err != nil {
		logger.Errorw("Problem with connecting to db ", err)
		return nil
	}
	return &PGDB{config: config, logger: logger, db: db}
}

func (p *PGDB) ConnectToDB() (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), p.config.DatabaseDsn) //sql.Open("pgx", p.config.DatabaseDsn)

	if err != nil {
		p.logger.Errorw("Problem with connecting to db ", err)
		return nil, err
	}

	err = db.Ping(context.Background())

	if err != nil {
		p.logger.Errorw("Problem with ping to db ", err)
		return nil, err
	}

	p.logger.Infow("Connecting and ping to db: OK")
	return db, nil
}

func (p *PGDB) Get(shortURL string) (string, bool) {
	var origURL string
	row := p.db.QueryRow(context.Background(), `SELECT original_url FROM urls WHERE short_url = $1`, shortURL)
	row.Scan(&origURL)
	if origURL == "" {
		return "", false
	}
	return origURL, true
}

func (p *PGDB) Set(shortURL, originalURL string) bool {

	result, err := p.db.Exec(context.Background(), `INSERT INTO urls (short_url, original_url) 
	VALUES ($1, $2) ON CONFLICT (short_url) DO NOTHING`, shortURL, originalURL)

	if rows := result.RowsAffected(); rows == 0 {
		return false
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return false
		}
	}

	return true
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
