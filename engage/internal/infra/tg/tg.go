package tg

import (
	"log/slog"

	"github.com/braunkc/hive-engage/config"
	"github.com/mymmrac/telego"
)

func New(token string, env config.Env, log *slog.Logger) (*telego.Bot, error) {
	opt := telego.WithDefaultLogger(true, true)
	if env != config.Debug {
		opt = telego.WithDefaultLogger(false, false)
	}

	bot, err := telego.NewBot(token, opt)
	if err != nil {
		return nil, err
	}

	return bot, nil
}
