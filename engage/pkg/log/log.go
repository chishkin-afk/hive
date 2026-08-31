package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type OutputType int

const (
	Console OutputType = 1
	File    OutputType = 2
	Both    OutputType = 3
)

type Config struct {
	Service     string
	OutputType  OutputType
	LogFilePath string     // default: root of project
	Level       slog.Level // default: slog.LevelInfo
	JSONFormat  bool       // default: false
}

type consoleHandler struct {
	handler slog.Handler
	cfg     *Config
}

type fileHandler struct {
	handler slog.Handler
	cfg     *Config
}

type CustomHandler struct {
	handlers []slog.Handler
	cfg      *Config
}

func NewHandler(cfg *Config) (*CustomHandler, error) {
	var handlers []slog.Handler

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     cfg.Level,
	}

	if cfg.OutputType&Console != 0 {
		var handler slog.Handler
		if cfg.JSONFormat {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = &consoleHandler{
				handler: slog.NewTextHandler(os.Stdout, opts),
				cfg:     cfg,
			}
		}

		handlers = append(handlers, handler)
	}

	if cfg.OutputType&File != 0 {
		if cfg.LogFilePath == "" {
			cfg.LogFilePath = fmt.Sprintf("%s.log", cfg.Service)
		}

		file, err := os.OpenFile(cfg.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}

		handler := &fileHandler{
			handler: slog.NewJSONHandler(file, opts),
			cfg:     cfg,
		}
		handlers = append(handlers, handler)
	}

	return &CustomHandler{
		handlers: handlers,
		cfg:      cfg,
	}, nil
}

func (h *consoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *consoleHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.cfg.JSONFormat {
		return h.handler.Handle(ctx, r)
	}

	return h.handlePretty(r)
}

func (h *consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &consoleHandler{
		handler: h.handler.WithAttrs(attrs),
		cfg:     h.cfg,
	}
}

func (h *consoleHandler) WithGroup(name string) slog.Handler {
	return &consoleHandler{
		handler: h.handler.WithGroup(name),
		cfg:     h.cfg,
	}
}

func (h *consoleHandler) handlePretty(r slog.Record) error {
	levelEmoji, levelColor := h.getLevelStyle(r.Level)
	timeStr := r.Time.UTC().Format("15:04:05.000")

	if _, err := fmt.Fprintf(os.Stdout, "%s %s%s %s",
		h.colorize(timeStr, "37"),
		levelEmoji,
		h.colorize(fmt.Sprintf("%6s", r.Level.String()), levelColor),
		h.colorize(r.Message, "1;36"),
	); err != nil {
		return err
	}

	attrs := make([]string, 0)

	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key == slog.SourceKey {
			if source, ok := attr.Value.Any().(*slog.Source); ok {
				shortFile := h.shortenFilePath(source.File)
				attrs = append(attrs, h.colorize(fmt.Sprintf("%s:%d", shortFile, source.Line), "35"))
			}
		} else {
			key := h.colorize(attr.Key+":", "37")
			value := h.colorize(fmt.Sprintf("%v", attr.Value.Any()), "35")
			attrs = append(attrs, key+value)
		}

		return true
	})

	if h.cfg.Service != "" {
		attrs = append(attrs, h.colorize("service:"+h.cfg.Service, "30"))
	}

	if len(attrs) > 0 {
		if _, err := fmt.Fprintf(os.Stdout, " %s", strings.Join(attrs, " ")); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(os.Stdout); err != nil {
		return err
	}

	return nil
}

func (h *consoleHandler) getLevelStyle(level slog.Level) (string, string) {
	switch level {
	case slog.LevelDebug:
		return "🐛", "37" // white
	case slog.LevelInfo:
		return "ℹ️", "32" // green
	case slog.LevelWarn:
		return "⚠️", "93" // yellow
	case slog.LevelError:
		return "❌", "31" // red
	default:
		return "🔍", "37"
	}
}

func (h *consoleHandler) colorize(text, colorCode string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", colorCode, text)
}

func (h *consoleHandler) shortenFilePath(path string) string {
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		return path[idx+1:]
	}
	return path
}

func (h *fileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *fileHandler) Handle(ctx context.Context, r slog.Record) error {
	utcRecord := r.Clone()
	utcRecord.Time = r.Time.UTC()

	return h.handler.Handle(ctx, utcRecord)
}

func (h *fileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &fileHandler{
		handler: h.handler.WithAttrs(attrs),
		cfg:     h.cfg,
	}
}

func (h *fileHandler) WithGroup(name string) slog.Handler {
	return &fileHandler{
		handler: h.handler.WithGroup(name),
		cfg:     h.cfg,
	}
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handlers[0].Enabled(ctx, level)
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))

	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}

	return &CustomHandler{
		handlers: newHandlers,
		cfg:      h.cfg,
	}
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))

	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}

	return &CustomHandler{
		handlers: newHandlers,
		cfg:      h.cfg,
	}
}
