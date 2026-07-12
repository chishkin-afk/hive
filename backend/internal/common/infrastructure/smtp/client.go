package smtpinfra

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
)

type SMTPClient struct {
	auth smtp.Auth
	addr string
	cfg  *config.Config
}

func New(cfg *config.Config) *SMTPClient {
	auth := smtp.PlainAuth("", cfg.SMTP.From, cfg.SMTP.Password, cfg.SMTP.Host)
	return &SMTPClient{
		auth: auth,
		addr: fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port),
		cfg:  cfg,
	}
}

func (sc *SMTPClient) Notify(ctx context.Context, body string, to []string) error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		msg := fmt.Sprintf(
			"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
			sc.cfg.SMTP.From, strings.Join(to, ","), sc.cfg.SMTP.Msg.Subject, body,
		)

		if err := smtp.SendMail(sc.addr, sc.auth, sc.cfg.SMTP.From, to, []byte(msg)); err != nil {
			errCh <- err
			return
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err, ok := <-errCh:
		if ok {
			return nil
		}

		return err
	}
}
