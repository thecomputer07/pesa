// Package http wires the chi router with all middleware + handlers.
package http

import (
	"context"
	"database/sql"
	"io"
	"io/fs"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"

	"github.com/thecomputer07/pesa/internal/config"
	"github.com/thecomputer07/pesa/internal/http/handlers"
	"github.com/thecomputer07/pesa/internal/http/health"
	mw "github.com/thecomputer07/pesa/internal/http/middleware"
	"github.com/thecomputer07/pesa/internal/service"
	"github.com/thecomputer07/pesa/internal/service/erasure"
	"github.com/thecomputer07/pesa/internal/service/imagecache"
	"github.com/thecomputer07/pesa/internal/service/off"
	"github.com/thecomputer07/pesa/internal/service/srt"
	syncsvc "github.com/thecomputer07/pesa/internal/service/sync"
)

// Deps groups the runtime objects needed to build the router.
type Deps struct {
	Cfg      *config.Config
	DB       *sql.DB
	Session  *scs.SessionManager
	Auth     *service.AuthService
	Mailer   service.Mailer
	Dist     fs.FS
	Erasure  *erasure.Service
	SRT      *srt.Service
	Notifier *srt.Notifier
	// ExportSecret is the HMAC secret for download tokens (PESA_EXPORT_TOKEN_SECRET).
	ExportSecret []byte
}

// erasureAdapter adapts *erasure.Service to srt.EraseInvoker so the srt
// package can invoke erasure.Erase without importing erasure directly.
type erasureAdapter struct{ s *erasure.Service }

func (a erasureAdapter) Erase(ctx context.Context, userID int64, channel, notes string) (any, error) {
	return a.s.Erase(ctx, userID, channel, notes)
}

// NewRouter assembles middleware + routes.
func NewRouter(d Deps) stdhttp.Handler {
	r := chi.NewRouter()

	r.Use(mw.Recover)
	r.Use(mw.ReqID)
	r.Use(mw.Logger)
	r.Use(mw.SecurityHeaders)
	r.Use(mw.I18n)
	r.Use(mw.GlobalRateLimit)
	r.Use(d.Session.LoadAndSave)
	r.Use(mw.AuthContext(d.Session))
	r.Use(mw.CSRF)

	h := handlers.New(d.Auth, d.Session, d.Cfg.Env == "prod")
	foods := handlers.NewFoods(d.DB)
	sync := handlers.NewSync(syncsvc.New(d.DB))

	// Phase 5 wiring: image cache, OFF proxy, barcode handler, sources handler.
	var imgStore *imagecache.Store
	if dir := d.Cfg.ImageCacheDir; dir != "" {
		if s, err := imagecache.NewStore(dir); err == nil {
			imgStore = s
		}
	}
	var offPersister off.ImagePersister
	if imgStore != nil {
		offPersister = handlers.NewOFFImagePersister(imgStore)
	}
	offSvc := off.NewService(off.NewRepo(d.DB), off.NewClient("https://world.openfoodfacts.org"), offPersister)
	barcode := handlers.NewBarcode(offSvc)
	barcode.DB = d.DB
	sources := handlers.NewSources(d.DB)
	var images *handlers.Images
	if imgStore != nil {
		images = handlers.NewImages(imgStore)
	}

	// Health (unauthenticated).
	r.Get("/healthz", health.Liveness)
	r.Get("/readyz", health.Readiness(d.DB))

	// Auth surface.
	r.Get("/verify-email", h.VerifyEmail)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/logout", h.Logout)
		r.Post("/password-reset/request", h.RequestPasswordReset)
		r.Post("/password-reset/confirm", h.ConfirmPasswordReset)
	})

	// Foods surface (phase 2 + phase 5 barcode + custom foods).
	r.Route("/api/v1/foods", func(r chi.Router) {
		r.Get("/search", foods.Search)
		r.Get("/mine", foods.Mine)
		r.Post("/custom", foods.Create)
		r.Patch("/custom/{id}", foods.Update)
		r.Delete("/custom/{id}", foods.Delete)
		r.With(mw.PerIPRate(30, time.Minute)).Get("/barcode/{ean}", barcode.Lookup)
		r.Get("/{id}", foods.Detail)
	})

	// Sources attribution surface (phase 5 — plan-04).
	r.Get("/api/v1/sources", sources.List)

	// Cached third-party image surface (phase 5 — plan-02). Mounted before SPA fallback.
	if images != nil {
		r.Get("/static/images/{source}/{file}", images.Serve)
	}

	// Sync surface (phase 4). Auth enforced in handler via middleware.UserIDFrom.
	r.Route("/api/v1/sync", func(r chi.Router) {
		r.Post("/", sync.Post)
		r.Get("/", sync.Get)
	})

	// Phase 6 privacy surface — /api/v1/me/*.
	meExport := handlers.NewMeExport(d.DB, d.ExportSecret)
	var meDelete *handlers.MeDelete
	if d.Erasure != nil {
		meDelete = handlers.NewMeDelete(d.DB, d.Session, d.Erasure, d.Auth, d.Cfg.Env == "prod")
	}
	me := handlers.NewMe(d.DB, d.Auth, d.Mailer, d.Cfg.CookieDomain)
	r.Route("/api/v1/me", func(r chi.Router) {
		r.Get("/", me.Get)
		r.Patch("/", me.Patch)
		if meDelete != nil {
			r.Delete("/", meDelete.Handle)
		}
		r.Post("/export", meExport.Request)
		r.Get("/export/{job_id}", meExport.Status)
		r.Get("/export/download", meExport.Download)
	})

	// Phase 6 staff surface — /admin/subject-rights/tickets.
	if d.SRT != nil {
		srtH := handlers.NewSRT(d.SRT, d.Notifier, erasureAdapter{s: d.Erasure})
		r.Route("/admin", func(r chi.Router) {
			r.Use(mw.RequireStaff(d.DB))
			r.Route("/subject-rights/tickets", func(r chi.Router) {
				r.Post("/", srtH.Create)
				r.Get("/", srtH.List)
				r.Patch("/{id}", srtH.Patch)
				r.Post("/{id}/close", srtH.Close)
			})
		})
	}

	// Always-served well-known endpoints.
	r.Get("/.well-known/security.txt", func(w stdhttp.ResponseWriter, req *stdhttp.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		f, err := d.Dist.Open(".well-known/security.txt")
		if err != nil {
			stdhttp.NotFound(w, req)
			return
		}
		defer func() { _ = f.Close() }()
		_, _ = io.Copy(w, f)
	})

	// pprof gate (no-op unless PESA_ENABLE_PPROF=1).
	mw.MountPprof(r)

	// SPA fallback.
	r.NotFound(spaHandler(d.Dist))

	return r
}

func spaHandler(dist fs.FS) stdhttp.HandlerFunc {
	fileServer := stdhttp.FileServer(stdhttp.FS(dist))
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			fileServer.ServeHTTP(w, r)
			return
		}
		f, err := dist.Open(path)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// Fall back to index.html for client-side routing.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	}
}
