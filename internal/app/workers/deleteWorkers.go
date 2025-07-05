// Package workers пакет с описанием работы воркера для удаления данных из базы данных
package workers

import (
	"context"
	"sync"

	"github.com/sinfirst/URL-Cutter/internal/app/storage/pg/postgresbd"
)

// Worker структура воркера
type Worker struct {
	deleteCh chan string
	db       *postgresbd.PGDB
	wg       sync.WaitGroup
}

// NewDeleteWorker конструктор для Worker
func NewDeleteWorker(ctx context.Context, db *postgresbd.PGDB, deleteCh chan string) *Worker {
	worker := &Worker{
		db:       db,
		deleteCh: deleteCh,
	}

	worker.wg.Add(1)
	go worker.DeleteWorker(ctx)

	return worker
}

// DeleteWorker удаляет урл из бд, взятый из канала
func (w *Worker) DeleteWorker(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case urlID, ok := <-w.deleteCh:
			if !ok {
				return
			}
			err := w.db.DeleteURL(ctx, urlID)
			if err != nil {
				return
			}
		}
	}
}

// StopWorker останавливает воркеры
func (w *Worker) StopWorker() {
	w.wg.Wait()
}
