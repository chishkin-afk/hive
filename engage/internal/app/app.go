package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/braunkc/hive-engage/config"
	ticketapplication "github.com/braunkc/hive-engage/internal/application/ticket"
	engagedomain "github.com/braunkc/hive-engage/internal/domain/engage"
	ticketdomain "github.com/braunkc/hive-engage/internal/domain/ticket"
	"github.com/braunkc/hive-engage/internal/infra/http"
	httpclient "github.com/braunkc/hive-engage/internal/infra/http/client"
	httphandlers "github.com/braunkc/hive-engage/internal/infra/http/handlers"
	httproutes "github.com/braunkc/hive-engage/internal/infra/http/routes"
	"github.com/braunkc/hive-engage/internal/infra/repository/cache"
	statecache "github.com/braunkc/hive-engage/internal/infra/repository/cache/state"
	ticketcache "github.com/braunkc/hive-engage/internal/infra/repository/cache/ticket"
	"github.com/braunkc/hive-engage/internal/infra/repository/database"
	ticketpersistence "github.com/braunkc/hive-engage/internal/infra/repository/persistence/ticket"
	"github.com/braunkc/hive-engage/internal/infra/tg"
	tgapi "github.com/braunkc/hive-engage/internal/infra/tg/api"
	tghandlers "github.com/braunkc/hive-engage/internal/infra/tg/handlers"
	"github.com/braunkc/hive-engage/pkg/log"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/redis/go-redis/v9"
)

type repositories struct {
	db    ticketdomain.TicketPersistenceRepository
	cache ticketdomain.TicketCacheRepository
	state *statecache.StateRepo
}

type services struct {
	ticket *ticketapplication.TicketService
}

type telegram struct {
	bot         *telego.Bot
	api         engagedomain.EngageAPI
	tgh         *tghandlers.TgHandlers
	updates     <-chan telego.Update
	webhookChan chan<- telego.Update
	bh          *telegohandler.BotHandler
}

type closers struct {
	pool   *pgxpool.Pool
	m      *migrate.Migrate
	client *redis.Client
}

type App struct {
	repos    *repositories
	services *services
	tg       *telegram
	client   *httpclient.HTTPClient
	server   *http.Server
	closers  *closers
	cfg      *config.Config
	log      *slog.Logger
}

func (a *App) Cfg() *config.Config {
	if a.cfg == nil {
		cfg, err := config.New(os.Getenv("CONFIG_PATH"))
		if err != nil {
			slog.Error("init config", slog.Any("err", err))
			os.Exit(1)
		}
		a.cfg = cfg
	}

	return a.cfg
}

func mapLogOutputType(outputType config.OutputType) log.OutputType {
	switch outputType {
	case config.OutputTypeFile:
		return log.File
	case config.OutputTypeBoth:
		return log.Both
	default:
		return log.Console
	}
}

func (a *App) Log() *slog.Logger {
	if a.log == nil {
		level := slog.LevelDebug
		if a.Cfg().Env == config.Prod {
			level = slog.LevelInfo
		}
		logCfg := log.Config{
			Service:     a.Cfg().Log.Service,
			OutputType:  mapLogOutputType(a.Cfg().Log.OutputType),
			LogFilePath: a.Cfg().Log.LogFilePath,
			Level:       level,
			JSONFormat:  a.Cfg().Log.JSON,
		}
		h, err := log.NewHandler(&logCfg)
		if err != nil {
			slog.Error("init logger", slog.Any("err", err))
			os.Exit(1)
		}
		a.log = slog.New(h)
	}

	return a.log
}

func (a *App) initTicketRepo(pool *pgxpool.Pool, log *slog.Logger) error {
	db, err := ticketpersistence.NewTicketRepo(pool, log)
	if err != nil {
		return fmt.Errorf("create ticket db repo: %w", err)
	}
	a.repos.db = db

	return nil
}

func (a *App) initDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.New(ctx, a.Cfg().DB)
	if err != nil {
		return fmt.Errorf("init database: %w", err)
	}
	a.closers.pool = pool

	if err = a.initTicketRepo(pool, a.Log()); err != nil {
		return err
	}

	m, err := migrate.New("file://migrations", fmt.Sprintf("postgresql://%s&x-migrations-table=engage_migrations", a.Cfg().DB.URL()))
	if err != nil {
		return fmt.Errorf("create migration: %w", err)
	}
	a.closers.m = m

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("up migration: %w", err)
	}

	a.Log().Info("database inited")
	return nil
}

func (a *App) initTicketCache(client *redis.Client) error {
	cache, err := ticketcache.NewTicketRepo(client, a.Log())
	if err != nil {
		return fmt.Errorf("create ticket cache repo: %w", err)
	}
	a.repos.cache = cache

	return nil
}

func (a *App) initStateCache(client *redis.Client) error {
	state, err := statecache.NewStateRepo(client, a.Log())
	if err != nil {
		return fmt.Errorf("create state cache repo: %w", err)
	}
	a.repos.state = state

	return nil
}

func (a *App) initCache() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := cache.New(ctx, a.Cfg().Cache)
	if err != nil {
		return fmt.Errorf("create cache client: %w", err)
	}
	a.closers.client = client

	if err = a.initTicketCache(client); err != nil {
		return err
	}
	if err = a.initStateCache(client); err != nil {
		return err
	}

	a.Log().Info("cache inited")
	return nil
}

func (a *App) Repos() *repositories {
	if a.repos == nil {
		a.repos, a.closers = &repositories{}, &closers{}
		if err := a.initDB(); err != nil {
			a.Log().Error("init db", slog.Any("err", err))
			os.Exit(1)
		}
		if err := a.initCache(); err != nil {
			a.Log().Error("init cache", slog.Any("err", err))
			os.Exit(1)
		}
	}

	return a.repos
}

func (s *services) initTicketService(
	db ticketdomain.TicketPersistenceRepository,
	cache ticketdomain.TicketCacheRepository,
	log *slog.Logger,
) {
	s.ticket = ticketapplication.NewTicketService(db, cache, log)
}

func (a *App) Services() *services {
	if a.services == nil {
		services := &services{}
		services.initTicketService(a.Repos().db, a.repos.cache, a.Log())
		a.services = services
		a.Log().Info("services inited")
	}

	return a.services
}

func (t *telegram) initAPI(log *slog.Logger) error {
	api, err := tgapi.NewTgAPI(t.bot, log)
	if err != nil {
		return fmt.Errorf("init telegram api: %w", err)
	}
	t.api = api

	return nil
}

func (t *telegram) initUpdates(env config.Env, cfg config.TgConfig) error {
	var updates <-chan telego.Update
	var webhookChan chan<- telego.Update
	if env == config.Prod {
		params := telego.SetWebhookParams{
			URL:         cfg.Webhook.URL,
			SecretToken: cfg.Webhook.Secret,
		}
		if err := t.bot.SetWebhook(context.Background(), &params); err != nil {
			return fmt.Errorf("set telegram webhook: %w", err)
		}

		ch := make(chan telego.Update, 128)
		updates, webhookChan = ch, ch
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := t.bot.DeleteWebhook(ctx, nil)
		if err != nil {
			return fmt.Errorf("delete telegram webhook: %w", err)
		}
		updates, err = t.bot.UpdatesViaLongPolling(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("get updates via long polling: %w", err)
		}
	}
	t.updates = updates
	t.webhookChan = webhookChan

	return nil
}

func (a *App) HTTPClient() *httpclient.HTTPClient {
	if a.client == nil {
		a.client = httpclient.New(a.Cfg().HTTP.Auth, a.Log())
		a.Log().Info("http client inited")
	}

	return a.client
}

func (t *telegram) initHandlers(
	client *httpclient.HTTPClient,
	ticketService *ticketapplication.TicketService,
	stateRepo *statecache.StateRepo,
	cfg config.TgConfig,
	log *slog.Logger,
) error {
	tgh, err := tghandlers.NewTgHandlers(t.api, client, ticketService, stateRepo, cfg, log)
	if err != nil {
		return fmt.Errorf("init telegram handlers: %w", err)
	}
	t.tgh = tgh

	bh, err := t.tgh.SetupHandlers(t.bot, t.updates)
	if err != nil {
		return fmt.Errorf("setup telegram handlers: %w", err)
	}
	t.bh = bh

	return nil
}

func (a *App) Telegram() *telegram {
	if a.tg == nil {
		telegram := &telegram{}

		bot, err := tg.New(a.Cfg().Tg.Token, a.Cfg().Env, a.Log())
		if err != nil {
			a.Log().Error("init telegram bot", slog.Any("err", err))
			os.Exit(1)
		}
		telegram.bot = bot

		if err = telegram.initAPI(a.Log()); err != nil {
			a.Log().Error("init telegram api", slog.Any("err", err))
			os.Exit(1)
		}
		if err := telegram.initUpdates(a.Cfg().Env, a.Cfg().Tg); err != nil {
			a.Log().Error("init telegram updates", slog.Any("err", err))
			os.Exit(1)
		}
		if err = telegram.initHandlers(a.HTTPClient(), a.Services().ticket, a.Repos().state, a.Cfg().Tg, a.Log()); err != nil {
			a.Log().Error("init telegram handlers", slog.Any("err", err))
			os.Exit(1)
		}

		a.tg = telegram
		a.Log().Info("telegram inited")
	}

	return a.tg
}

func (a *App) HTTPServer() *http.Server {
	if a.server == nil {
		server := http.New(a.Cfg().Env, a.Cfg().HTTP, a.Log())
		handlers := httphandlers.New(a.Telegram().api, a.Telegram().bh, a.Log())
		httproutes.Setup(server.Router, handlers, a.Telegram().webhookChan, a.Cfg().Tg.Webhook)

		a.server = server

		a.Log().Info("http server inited")
	}

	return a.server
}

func New() *App {
	a := &App{}
	a.HTTPServer()

	a.Log().Info("app created")
	return a
}

func (a *App) Run() {
	errChan := make(chan error, 2)

	go func() {
		a.Log().Info("starting bot handlers...")
		if err := a.Telegram().bh.Start(); err != nil {
			a.Log().Error("failed to start handling", slog.Any("err", err))
			errChan <- err
		}
	}()

	go func() {
		if err := a.HTTPServer().Run(); err != nil {
			a.Log().Error("failed to run http server", slog.Any("err", err))
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var exitCode int
	select {
	case err := <-errChan:
		a.Log().Error("application error, shutting down...", slog.Any("err", err))
		exitCode = 1
	case <-quit:
		a.Log().Info("received signal, shutting down...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.Cfg().ShutdownTimeout)
	defer cancel()

	a.Shutdown(ctx)
	os.Exit(exitCode)
}

func (a *App) Shutdown(ctx context.Context) {
	done := make(chan struct{})

	go func() {
		defer close(done)

		if err := a.HTTPServer().Stop(ctx); err != nil {
			a.Log().Error("failed to stop http server", slog.Any("err", err))
		}

		a.Log().Info("stopping bot handlers...")
		if err := a.Telegram().bh.Stop(); err != nil {
			a.Log().Error("failed to stop bh", slog.Any("err", err))
		}

		a.Log().Info("closing cache client...")
		if err := a.closers.client.Close(); err != nil {
			a.Log().Error("failed to close cache client", slog.Any("err", err))
		}

		a.Log().Info("closing database connection...")
		a.closers.pool.Close()
		sourceErr, dbErr := a.closers.m.Close()
		if sourceErr != nil {
			a.Log().Error("source error", slog.Any("err", sourceErr))
		}
		if dbErr != nil {
			a.Log().Error("db error", slog.Any("err", dbErr))
		}
	}()

	select {
	case <-done:
		a.Log().Info("gracefully stopped")
	case <-ctx.Done():
		a.Log().Error("shutdown timeout exceeded, forcing exit", slog.Any("err", ctx.Err()))
	}
}
