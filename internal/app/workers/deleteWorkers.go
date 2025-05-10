package workers

import (
	"context"
	"sync"

	"github.com/sinfirst/URL-Cutter/internal/app/app"
	"github.com/sinfirst/URL-Cutter/internal/app/storage/pg/postgresbd"
)

type Worker struct {
	deleteCh chan string
	db       *postgresbd.PGDB
	wg       sync.WaitGroup
	handler  *app.App
}

func NewDeleteWorker(ctx context.Context, db *postgresbd.PGDB, deleteCh chan string, handler *app.App) *Worker {
	worker := &Worker{
		db:       db,
		deleteCh: deleteCh,
		handler:  handler,
	}

	worker.wg.Add(2)
	go worker.UpdateDeleteWorker(ctx)
	go worker.DeleteWorker(ctx)

	return worker
}

func (w *Worker) UpdateDeleteWorker(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case urlID, ok := <-w.deleteCh:
			if !ok {
				return
			}
			w.db.UpdateDeleteParam(ctx, urlID)
			w.handler.AddToChan(urlID)
		}
	}
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
			w.db.Delete(ctx, urlID)
		}
	}
}

func (w *Worker) StopWorker() {
	w.wg.Wait()
}
