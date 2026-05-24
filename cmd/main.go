package main

import (
	"flag"
	"fmt"
	"log/slog"
	"mydb/internal/config"
	"mydb/internal/wal"
	"os"
	"path"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
	})))

	configPath := flag.String("config", "", "Config path")
	flag.Parse()

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "error: -config required")
		flag.Usage()
		os.Exit(2)
	}

	if err := config.LoadConfig(*configPath); err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}
	slog.Info("config loaded", "data_dir", config.CONFIG.DATA_DIR, "wal_dir", config.CONFIG.WAL_DIR)

	w, err := wal.New(path.Join(config.CONFIG.WAL_DIR, "test.bin"))
	if err != nil {
		slog.Error("wal new", "err", err)
		os.Exit(1)
	}

	data := "Hello, World"
	w.Append(&wal.Begin{XID: 1})
	w.Append(&wal.Update{XID: 1, PageID: 1, Offset: 1, After: []byte(data)})
	w.Append(&wal.Commit{XID: 1})
}
