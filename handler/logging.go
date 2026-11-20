package handler

import (
	"Control/types"
	"io"
	"log/slog"
	"os"

	"github.com/robfig/cron/v3"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logCron *cron.Cron

func LoggingHandler() {
	cfg := types.GetConfig()
	willi := &lumberjack.Logger{
		Filename:   cfg.Log_Filename,
		MaxAge:     cfg.Log_Max_Age,
		MaxBackups: cfg.Log_Max_Backups,
		Compress:   cfg.Log_compress,
	}

	multiwriter := io.MultiWriter(willi, os.Stdout)

	logger := slog.New(slog.NewJSONHandler(multiwriter, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)

	logCron := cron.New()
	if _, err := logCron.AddFunc("0 0 * * *", func() {
		if err := willi.Rotate(); err != nil {
			logger.Error("log rotation error", "err", err)
			return
		}
		slog.Info("log rotated")
	}); err != nil {
		logger.Error("log rotation cron error", "err", err)
	}

	logCron.Start()
}

func StopLogging() {
	if logCron != nil {
		logCron.Stop()
	}
}
