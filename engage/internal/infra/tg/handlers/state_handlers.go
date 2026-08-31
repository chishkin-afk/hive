package tghandlers

import (
	"context"

	statecache "github.com/braunkc/hive-engage/internal/infra/repository/cache/state"
	"github.com/mymmrac/telego"
)

type stepHandler func(ctx context.Context, upd telego.Update, state statecache.State) error

func (tgh *TgHandlers) prepareState(ctx context.Context, upd telego.Update, state statecache.State) error {
	handler := tgh.getTicketStepHandler(state.Step)

	if handler != nil {
		return handler(ctx, upd, state)
	}

	return nil
}
