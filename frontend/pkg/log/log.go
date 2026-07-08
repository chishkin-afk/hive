package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"log/slog"

	"github.com/devathh/hive/frontend/internal/infrastructure/config"
	"github.com/fatih/color"
)

type Handler struct {
	handler slog.Handler
	attrs   []slog.Attr
	groups  []string
}

func New(cfg *config.Config) *slog.Logger {
	var handler slog.Handler

	switch strings.ToLower(cfg.App.Env) {
	case "dev", "staging":
		handler = NewColorHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	case "prod":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return slog.New(handler)
}

func NewColorHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &Handler{
		handler: slog.NewTextHandler(w, opts),
	}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		handler: h.handler.WithAttrs(attrs),
		attrs:   append(h.attrs, attrs...),
		groups:  h.groups,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		handler: h.handler.WithGroup(name),
		attrs:   h.attrs,
		groups:  append(h.groups, name),
	}
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	// Colorize the level
	var levelColor *color.Color
	switch r.Level {
	case slog.LevelDebug:
		levelColor = color.New(color.FgBlue, color.Bold)
	case slog.LevelInfo:
		levelColor = color.New(color.FgGreen, color.Bold)
	case slog.LevelWarn:
		levelColor = color.New(color.FgYellow, color.Bold)
	case slog.LevelError:
		levelColor = color.New(color.FgRed, color.Bold)
	default:
		levelColor = color.New(color.Reset)
	}

	timeStr := r.Time.Format("15:04:05.000")
	levelStr := levelColor.Sprint(strings.ToUpper(r.Level.String()))
	msgStr := color.New(color.Bold).Sprint(r.Message)
	prefix := fmt.Sprintf("%s [%s] %s ", timeStr, levelStr, msgStr)

	fmt.Fprint(os.Stderr, prefix)

	r.Attrs(func(a slog.Attr) bool {
		keyColor := color.New(color.FgCyan)
		valColor := color.New(color.Reset)

		if a.Key == "err" {
			valColor = color.New(color.FgRed)
		}

		fmt.Fprintf(os.Stderr, "%s=%v ", keyColor.Sprint(a.Key), valColor.Sprint(a.Value))
		return true
	})

	fmt.Fprintln(os.Stderr)
	return nil
}
