package sgl

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Level uint8

const (
	// FatalLevel level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel = iota
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel

	delimiter = ' '
)

// A constant exposing all logging levels.
var AllLevels = []Level{
	FatalLevel,
	ErrorLevel,
	InfoLevel,
	DebugLevel,
}

func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	}

	return "unknown"
}

func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid Level: %q", lvl)
}

type PrintLogger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}

type Logger interface {
	PrintLogger

	SetLevel(level Level)
	Level() Level
	Out() io.Writer
	Fields() []Field

	WithErr(err error) PrintLogger
	WithField(key, val string) Logger
}

// Add hook for remote logging.
type SimpleLogger struct {
	out io.Writer // destination for output
	buf []byte    // for accumulating text to write

	// Atomic level updates & writes.
	mu sync.Mutex

	level  Level
	fields []Field

	timeNow func() time.Time
}

type Field struct {
	Key   string
	Value string
}

func New(out io.Writer) *SimpleLogger {
	l := &SimpleLogger{
		// TODO(bplotka): Support different log writers for err logs.
		out:     out,
		level:   InfoLevel,
		timeNow: time.Now,
	}
	return l
}

func (l *SimpleLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.level = level
}

func (l *SimpleLogger) Level() Level {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.level
}

func (l *SimpleLogger) Debug(format string, args ...interface{}) {
	if l.level >= DebugLevel {
		l.write(DebugLevel, format, args...)
	}
}

func (l *SimpleLogger) Info(format string, args ...interface{}) {
	if l.level >= InfoLevel {
		l.write(InfoLevel, format, args...)
	}
}

func (l *SimpleLogger) Error(format string, args ...interface{}) {
	if l.level >= ErrorLevel {
		l.write(ErrorLevel, format, args...)
	}
}

func (l *SimpleLogger) Fatal(format string, args ...interface{}) {
	l.write(FatalLevel, format, args...)
	os.Exit(1)
}

func (l *SimpleLogger) WithField(key, val string) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	return &SimpleLogger{
		out:     l.out,
		level:   l.level,
		fields:  append(l.fields, Field{Key: key, Value: val}),
		timeNow: l.timeNow,
	}
}

func (l *SimpleLogger) WithErr(err error) PrintLogger {
	return l.WithField("err", err.Error())
}

func (l *SimpleLogger) write(lvl Level, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	now := l.timeNow()
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Reset buffer.
	l.buf = l.buf[:0]
	// Time tag.
	l.buf = append(l.buf, 't', 'i', 'm', 'e', '=', '"')
	l.buf = append(l.buf, now.String()...)
	l.buf = append(l.buf, '"')
	// Level tag.
	l.buf = append(l.buf, delimiter, 'l', 'v', 'l', '=')
	l.buf = append(l.buf, lvl.String()...)
	// Msg tag.
	l.buf = append(l.buf, delimiter, 'm', 's', 'g', '=', '"')
	l.buf = append(l.buf, s...)
	l.buf = append(l.buf, '"')

	// Pack all fields.
	for _, f := range l.fields {
		l.buf = append(l.buf, delimiter)
		l.buf = append(l.buf, f.Key...)
		l.buf = append(l.buf, '=', '"')
		l.buf = append(l.buf, f.Value...)
		l.buf = append(l.buf, '"')
	}

	// File tag.
	l.buf = append(l.buf, delimiter, 'f', 'i', 'l', 'e', '=')
	l.buf = append(l.buf, short...)
	l.buf = append(l.buf, ':')
	itoa(&l.buf, line, -1)

	// Put newline always.
	l.buf = append(l.buf, '\n')
	_, err := l.out.Write(l.buf)
	if err != nil {
		println(err.Error())
	}
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *SimpleLogger) Fields() []Field {
	return l.fields
}

func (l *SimpleLogger) Out() io.Writer {
	return l.out
}