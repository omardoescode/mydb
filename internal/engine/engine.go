package engine

import "log/slog"

type Engine struct {
}

func New() *Engine {
	return &Engine{}
}

func (eng Engine) Run() {
	slog.Info("engine started", "name", "mydb")
}
