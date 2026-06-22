package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/ishchibormi/backend/config"
	"github.com/ishchibormi/backend/internal/admin"
	"github.com/ishchibormi/backend/internal/application"
	"github.com/ishchibormi/backend/internal/auth"
	"github.com/ishchibormi/backend/internal/category"
	"github.com/ishchibormi/backend/internal/chat"
	"github.com/ishchibormi/backend/internal/elon"
	"github.com/ishchibormi/backend/internal/finance"
	"github.com/ishchibormi/backend/internal/notification"
	"github.com/ishchibormi/backend/internal/report"
	"github.com/ishchibormi/backend/internal/review"
	"github.com/ishchibormi/backend/internal/upload"
	"github.com/ishchibormi/backend/internal/user"
	"github.com/ishchibormi/backend/pkg/db"
	"github.com/ishchibormi/backend/pkg/envfile"
	"github.com/ishchibormi/backend/pkg/httpx"
	"github.com/ishchibormi/backend/pkg/logger"
	"github.com/ishchibormi/backend/pkg/storage"
)

func main() {
	envfile.Load()
	log := logger.New()
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mdb, err := db.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Error("mongo connect failed", "err", err)
		os.Exit(1)
	}
	if err := db.EnsureIndexes(ctx, mdb); err != nil {
		log.Warn("ensure indexes", "err", err)
	}

	// services
	hub := chat.NewHub()
	notif := notification.New(mdb)
	notif.AttachPusher(hub)

	// S3 storage — optional. If creds aren't set, upload endpoints return 503.
	var s3svc *storage.Service
	if cfg.AWSS3Bucket != "" {
		s3svc, err = storage.New(ctx, storage.Config{
			Region: cfg.AWSRegion, AccessKeyID: cfg.AWSAccessKeyID,
			SecretAccessKey: cfg.AWSSecretAccessKey,
			Bucket: cfg.AWSS3Bucket, PublicBaseURL: cfg.AWSS3PublicBaseURL,
		})
		if err != nil {
			log.Warn("s3 init", "err", err)
		} else {
			log.Info("s3 ready", "bucket", cfg.AWSS3Bucket, "region", cfg.AWSRegion)
		}
	} else {
		log.Warn("s3 disabled: AWS_S3_BUCKET not set")
	}

	authH := auth.NewHandler(cfg, mdb)
	userH := user.NewHandler(mdb, s3svc)
	catH := category.NewHandler(mdb)
	elonH := elon.NewHandler(mdb, s3svc)
	appH := application.NewHandler(mdb, notif)
	revH := review.NewHandler(mdb, notif)
	finH := finance.NewHandler(mdb)
	repH := report.NewHandler(mdb)
	chatH := chat.NewHandler(mdb, hub, notif, cfg)
	uploadH := upload.NewHandler(s3svc)
	adminH := admin.NewHandler(cfg, mdb, notif, s3svc)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(httpx.Recover)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	otpLimiter := httpx.NewLimiter(10, 0.5)   // 10 burst, 1 / 2s
	chatLimiter := httpx.NewLimiter(60, 2)    // 60 burst, 2 / s
	applyLimiter := httpx.NewLimiter(20, 0.5) // 20 burst, slow refill

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { httpx.JSON(w, 200, map[string]string{"status": "ok"}) })

	r.Route("/api", func(r chi.Router) {
		// Public auth
		r.Group(func(r chi.Router) {
			r.Use(otpLimiter.Middleware("otp"))
			r.Post("/auth/otp/request", authH.RequestOTP)
			r.Post("/auth/otp/verify", authH.VerifyOTP)
			r.Get("/auth/otp/peek", authH.DevPeekOTP)
		})
		r.Post("/auth/refresh", authH.Refresh)

		// Public listing
		r.Get("/elons", elonH.Feed)
		r.Get("/elons/{id}", elonH.Get)
		r.Get("/users/{id}", userH.GetPublic)
		r.Get("/users", userH.Search)
		r.Get("/categories", catH.List)
		r.Get("/users/{id}/reviews", revH.ListForUser)

		// WS (public — token comes via ?token=)
		r.Get("/ws", chatH.WS)

		// Auth-protected
		r.Group(func(r chi.Router) {
			r.Use(httpx.UserAuth(cfg.JWTAccessSecret))

			r.Get("/me", userH.Me)
			r.Patch("/me", userH.UpdateMe)
			r.Post("/users/{id}/block", userH.Block)
			r.Delete("/users/{id}/block", userH.Unblock)

			r.Post("/categories", catH.Create)

			r.Post("/elons", elonH.Create)
			r.Patch("/elons/{id}", elonH.Update)
			r.Delete("/elons/{id}", elonH.Delete)
			r.Post("/elons/{id}/publish", elonH.Publish)
			r.Get("/my/elons", elonH.MyElons)

			r.Group(func(r chi.Router) {
				r.Use(applyLimiter.Middleware("apply"))
				r.Post("/elons/{id}/apply", appH.Apply)
			})
			r.Post("/applications/{id}/accept", appH.Accept)
			r.Post("/applications/{id}/reject", appH.Reject)
			r.Post("/applications/{id}/cancel", appH.Cancel)
			r.Post("/applications/{id}/confirm-done", appH.ConfirmDone)
			r.Post("/applications/{id}/review", revH.Create)

			r.Get("/my/applications", appH.MyApplications)
			r.Get("/my/elons/applications", appH.MyElonsApplications)
			r.Get("/me/history", appH.History)
			r.Get("/me/finance", finH.Me)

			r.Get("/conversations", chatH.ListConversations)
			r.Post("/conversations", chatH.StartConversation)
			r.Get("/conversations/{id}/messages", chatH.ListMessages)
			r.Group(func(r chi.Router) {
				r.Use(chatLimiter.Middleware("chat"))
				r.Post("/conversations/{id}/messages", chatH.SendMessage)
			})

			r.Get("/notifications", notif.List)
			r.Post("/notifications/read-all", notif.ReadAll)

			r.Post("/reports", repH.Create)

			// Uploads
			r.Post("/uploads", uploadH.Upload)
			r.Delete("/uploads", uploadH.Delete)
		})

		// Admin
		r.Post("/admin/login", adminH.Login)
		r.Route("/admin", func(r chi.Router) {
			r.Use(httpx.AdminAuth(cfg.JWTAccessSecret))
			r.Get("/dashboard", adminH.Dashboard)
			r.Get("/users", adminH.ListUsers)
			r.Post("/users/{id}/block", adminH.BlockUser)
			r.Delete("/users/{id}", adminH.DeleteUser)
			r.Get("/elons", adminH.ListElons)
			r.Delete("/elons/{id}", adminH.DeleteElon)
			r.Get("/categories", adminH.ListCategories)
			r.Patch("/categories/{id}/active", adminH.SetCategoryActive)
			r.Get("/reports", repH.ListAdmin)
			r.Patch("/reports/{id}/resolve", repH.Resolve)
			r.Post("/broadcast", adminH.Broadcast)
			r.Get("/audit", adminH.Audit)
		})
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("api listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Info("shutting down")
	shutdownCtx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	_ = srv.Shutdown(shutdownCtx)
}
