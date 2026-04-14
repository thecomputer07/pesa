// Command pesa is the single binary serving the Italian carb calculator.
//
// Subcommands:
//   pesa serve              Run the HTTP server (default).
//   pesa migrate up         Apply pending migrations.
//   pesa migrate down       Revert the last migration.
//   pesa migrate status     Print migration status.
package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	stdhttp "net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/acme/autocert"

	"github.com/thecomputer07/pesa/internal/config"
	pesahttp "github.com/thecomputer07/pesa/internal/http"
	"github.com/thecomputer07/pesa/internal/repo"
	"github.com/thecomputer07/pesa/internal/service"
	"github.com/thecomputer07/pesa/internal/service/erasure"
	"github.com/thecomputer07/pesa/internal/service/export"
	"github.com/thecomputer07/pesa/internal/service/srt"
	pesaweb "github.com/thecomputer07/pesa/internal/web"
)

// Set by ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	args := os.Args[1:]
	cmd := "serve"
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	switch cmd {
	case "serve":
		runServe()
	case "migrate":
		runMigrate(args)
	case "scrape":
		runScrape(args)
	case "import":
		runImport(args)
	case "apply-tombstones":
		runApplyTombstones(args)
	case "staff":
		runStaff(args)
	case "version", "-v", "--version":
		fmt.Printf("pesa %s (commit %s, built %s)\n", version, commit, date)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		fmt.Fprintln(os.Stderr, "usage: pesa [serve|migrate|scrape|import|apply-tombstones|staff|version]")
		os.Exit(2)
	}
}

func mustConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	return cfg
}

func openDB(cfg *config.Config) *sql.DB {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		slog.Error("db open", "err", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db
}

func runMigrate(args []string) {
	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()

	sub := "up"
	if len(args) > 0 {
		sub = args[0]
	}
	var err error
	switch sub {
	case "up":
		err = repo.Migrate(db)
	case "down":
		err = repo.MigrateDown(db)
	case "status":
		err = repo.MigrateStatus(db)
	default:
		fmt.Fprintf(os.Stderr, "unknown migrate subcommand: %s\n", sub)
		os.Exit(2)
	}
	if err != nil {
		slog.Error("migrate", "sub", sub, "err", err)
		os.Exit(1)
	}
	slog.Info("migrate ok", "sub", sub)
}

func runServe() {
	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		slog.Warn("db ping failed at startup (continuing)", "err", err)
	}

	// Session manager.
	sm := scs.New()
	sm.Store = postgresstore.New(db)
	sm.Lifetime = 30 * 24 * time.Hour
	sm.IdleTimeout = 14 * 24 * time.Hour
	sm.Cookie.Name = "pesa_session"
	sm.Cookie.HttpOnly = true
	sm.Cookie.Secure = cfg.Env == "prod"
	sm.Cookie.SameSite = stdhttp.SameSiteLaxMode
	sm.Cookie.Path = "/"
	if cfg.CookieDomain != "" {
		sm.Cookie.Domain = cfg.CookieDomain
	}

	// Mailer.
	var mailer service.Mailer
	if cfg.Env == "prod" && cfg.SMTPHost != "" {
		mailer = &service.SMTPMailer{
			Host: cfg.SMTPHost, Port: cfg.SMTPPort,
			User: cfg.SMTPUser, Password: cfg.SMTPPassword, From: cfg.SMTPFrom,
		}
	} else {
		mailer = service.StdoutMailer{}
	}

	auth := &service.AuthService{
		DB:      db,
		Mailer:  mailer,
		Now:     time.Now,
		Domain:  cfg.CookieDomain,
		WebBase: cfg.WebBaseURL,
	}

	dist, err := fs.Sub(pesaweb.Dist, "dist")
	if err != nil {
		slog.Error("embed sub-fs", "err", err)
		os.Exit(1)
	}

	// Phase 6: privacy surface wiring.
	exportSecret := []byte(os.Getenv("PESA_EXPORT_TOKEN_SECRET"))
	if len(exportSecret) == 0 {
		exportSecret = []byte("dev-insecure-export-secret-change-me")
	}
	logSalt := []byte(os.Getenv("PESA_LOG_SALT"))
	if len(logSalt) == 0 {
		logSalt = []byte("dev-insecure-log-salt")
	}
	hashSalt := []byte(os.Getenv("PESA_ERASURE_HASH_SALT"))
	if len(hashSalt) == 0 {
		hashSalt = []byte("dev-insecure-erasure-salt")
	}
	cacheDir := os.Getenv("PESA_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = filepath.Join(".cache", "exports")
	}
	erasureSvc := erasure.New(db, hashSalt, logSalt, cacheDir)
	srtSvc := srt.New(db, hashSalt)
	srtNotifier := srt.NewNotifier(os.Getenv("PESA_LEGAL_TEMPLATES_DIR"), mailer)

	router := pesahttp.NewRouter(pesahttp.Deps{
		Cfg: cfg, DB: db, Session: sm, Auth: auth, Mailer: mailer, Dist: dist,
		Erasure: erasureSvc, SRT: srtSvc, Notifier: srtNotifier,
		ExportSecret: exportSecret,
	})

	// Start export worker (PRIV-01).
	exportWorker := export.NewWorker(db, cacheDir, exportSecret, logSalt)
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	go exportWorker.Run(workerCtx)

	srv := &stdhttp.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown wiring.
	idleConns := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down")
		_ = srv.Close()
		close(idleConns)
	}()

	if cfg.EnableTLS {
		runTLS(srv, cfg)
	} else {
		slog.Info("pesa listening", "addr", cfg.Addr, "version", version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			slog.Error("listen", "err", err)
			os.Exit(1)
		}
	}
	<-idleConns
}

func runTLS(srv *stdhttp.Server, cfg *config.Config) {
	hosts := strings.Split(cfg.AutocertHosts, ",")
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hosts...),
		Cache:      autocert.DirCache(cfg.AutocertCacheDir),
		Email:      cfg.AutocertEmail,
	}
	tlsCfg := m.TLSConfig()
	tlsCfg.MinVersion = tls.VersionTLS13
	tlsCfg.CipherSuites = []uint16{
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
	}
	tlsCfg.PreferServerCipherSuites = true

	// HTTP-01 + redirect on :80.
	go func() {
		hsrv := &stdhttp.Server{
			Addr:              ":80",
			Handler:           m.HTTPHandler(nil),
			ReadHeaderTimeout: 5 * time.Second,
		}
		_ = hsrv.ListenAndServe()
	}()

	srv.Addr = ":443"
	srv.TLSConfig = tlsCfg
	slog.Info("pesa listening TLS", "addr", srv.Addr, "hosts", cfg.AutocertHosts, "version", version)
	if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
		slog.Error("listen TLS", "err", err)
		os.Exit(1)
	}
}
