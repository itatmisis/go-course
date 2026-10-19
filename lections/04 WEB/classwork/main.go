package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
)

func main() {
	logg()
}

func logg() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: LevelCustom,
	}))

	logger.Debug("незначительное событие", "step", 1)
	logger.Info("важное событие", "user_id", 42)
	logger.Warn("необычное поведение", "attempts", 3)
	logger.Error("ошибка запроса", "status", 500)
	logger.Log(context.Background(), LevelCustom, "msg", "key", "val")

	err := errors.New("ошибка подключения")
	logger.Error("не удалось получить данные", "error", err)
}

const LevelCustom slog.Level = 10
