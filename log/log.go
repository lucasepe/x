package log

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"

	"github.com/lucasepe/x/env"
)

func I(msg string, fields ...Field) {
	globalLogger().I(msg, fields...)
}

func D(msg string, fields ...Field) {
	globalLogger().D(msg, fields...)
}

func W(msg string, fields ...Field) {
	globalLogger().W(msg, fields...)
}

func E(msg string, fields ...Field) {
	globalLogger().E(msg, fields...)
}

// globalLogger returns the singleton Logger instance
func globalLogger() *Logger {
	once.Do(func() {
		instance = New(os.Stderr)
	})
	return instance
}

// internal singleton logic
var (
	once     sync.Once
	instance *Logger
)

type Level int

const (
	Debug Level = iota
	Info
	Warning
	Error
)

func (l Level) String() string {
	return [...]string{"D", "I", "W", "E"}[l]
}

type Logger struct {
	out        io.Writer
	minLevel   Level
	mu         sync.Mutex
	bufPool    sync.Pool
	pretty     bool
	withTime   bool
	timeFormat string
}

func New(out io.Writer) *Logger {
	l := &Logger{
		out:      out,
		minLevel: Info,
		bufPool: sync.Pool{
			New: func() any { return new(bytes.Buffer) },
		},
	}
	if env.True("DEBUG") {
		l.minLevel = Debug
	}
	if env.True("PRETTY") {
		l.pretty = true
	}

	if env.True("TIMESTAMP") {
		l.WithTimestamp(true, "02 Jan 2006 15:04")
	}
	return l
}

func (l *Logger) WithPretty(pretty bool) {
	l.pretty = pretty
}

func (l *Logger) WithTimestamp(enabled bool, layout string) {
	l.withTime = enabled
	l.timeFormat = layout
}

func (l *Logger) D(msg string, fields ...Field) { l.log(Debug, msg, fields...) }
func (l *Logger) I(msg string, fields ...Field) { l.log(Info, msg, fields...) }
func (l *Logger) W(msg string, fields ...Field) { l.log(Warning, msg, fields...) }
func (l *Logger) E(msg string, fields ...Field) { l.log(Error, msg, fields...) }

func (l *Logger) log(level Level, msg string, fields ...Field) {
	if level < l.minLevel {
		return
	}

	buf := l.bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	buf.WriteByte('[')
	buf.WriteString(level.String())
	buf.WriteString("] ")

	// Timestamp subito dopo il livello
	if l.withTime {
		ts := time.Now().Format(l.timeFormat)
		buf.WriteString(ts)
		buf.WriteByte(' ')
	}

	// Messaggio
	buf.WriteString(msg)
	buf.WriteByte('\n')

	if l.pretty && len(fields) > 0 {
		// Calcola padding massimo per chiavi
		maxLen := 0
		for _, f := range fields {
			if len(f.Key) > maxLen {
				maxLen = len(f.Key)
			}
		}

		// Stampa ogni campo con indentazione
		for i, f := range fields {
			prefix := "├──"
			if i == len(fields)-1 {
				prefix = "└──"
			}
			buf.WriteString("    ") // indent
			buf.WriteString(prefix)
			buf.WriteByte(' ')
			buf.WriteString(f.Key)
			buf.WriteString(": ")

			// Allinea con padding
			spaces := maxLen - len(f.Key)
			for s := 0; s < spaces; s++ {
				buf.WriteByte(' ')
			}

			f.Encoder.Encode(buf)
			buf.WriteByte('\n')
		}
	} else {
		// Fallback classico in-line
		for _, f := range fields {
			buf.WriteString("   ")
			buf.WriteString(f.Key)
			buf.WriteByte('=')
			f.Encoder.Encode(buf)
		}
		buf.WriteByte('\n')
	}

	l.mu.Lock()
	_, _ = l.out.Write(buf.Bytes())
	l.mu.Unlock()

	l.bufPool.Put(buf)
}
