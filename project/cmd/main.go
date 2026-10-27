package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"task-manager/internal/app"
	"task-manager/internal/presentations/task_presentation"
	"task-manager/internal/services/task_service"
	"task-manager/internal/storage/postgres_storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %s", err.Error())
	}
	defer pool.Close()

	pgStorage := postgres_storage.NewStorage(pool)

	taskService := task_service.NewService(pgStorage)

	taskPresentation := task_presentation.NewPresentation(taskService)

	e := app.BuildApp(taskPresentation)
	defer e.Close()

	if err := e.Start(":8080"); err != nil {
		panic(fmt.Errorf("can't start server: %w", err))
	}
}
