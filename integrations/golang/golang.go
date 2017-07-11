package sgl_golang

import (
	"log"

	"github.com/Bplotka/sgl"
)

type errLoggerWriter struct {
	l sgl.Logger
}

func (w *errLoggerWriter) Write(p []byte) (n int, err error) {
	w.l.Error(string(p))
	return len(p), nil
}

func From(s sgl.Logger) *log.Logger {
	return log.New(&errLoggerWriter{l: s}, "", 0)
}
