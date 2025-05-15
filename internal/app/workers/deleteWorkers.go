package workers

import (
	"context"
	"sync"

	"github.com/sinfirst/URL-Cutter/internal/app/storage/pg/postgresbd"
)

type Worker struct {
	deleteCh chan string
	db       *postgresbd.PGDB
	wg       sync.WaitGroup
}

func NewDeleteWorker(ctx context.Context, db *postgresbd.PGDB, deleteCh chan string) *Worker {
	worker := &Worker{
		db:       db,
		deleteCh: deleteCh,
	}

	worker.wg.Add(1)
	go worker.DeleteWorker(ctx)

	return worker
}

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

func (w *Worker) StopWorker() {
	w.wg.Wait()
}
