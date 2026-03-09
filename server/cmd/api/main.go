package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/lnardon/arete/internal/api"
	"github.com/lnardon/arete/internal/api/handlers"
	"github.com/lnardon/arete/internal/api/middleware"
	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/config"
	"github.com/lnardon/arete/internal/database"
	"github.com/lnardon/arete/internal/repository"
)

func resolveStaticDir(staticDir string) string {
	if staticDir == "" {
		return ""
	}

	relPath := filepath.FromSlash(staticDir)
	if filepath.IsAbs(staticDir) {
		if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
			return staticDir
		}
		return ""
	}

	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for i := 0; i < 10; i++ {
			try := filepath.Join(dir, relPath)
			if info, err := os.Stat(try); err == nil && info.IsDir() {
				return try
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	execPath, err := os.Executable()
	if err != nil {
		return ""
	}

	dir := filepath.Dir(execPath)
	for i := 0; i < 10; i++ {
		try := filepath.Join(dir, relPath)
		if info, err := os.Stat(try); err == nil && info.IsDir() {
			return try
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	staticDir := resolveStaticDir(cfg.Server.StaticDir)
	if cfg.Server.StaticDir != "" && staticDir == "" {
		log.Printf("WARNING: static dir %q not found; frontend will 404", cfg.Server.StaticDir)
	} else if staticDir != "" {
		log.Printf("Serving frontend from %s", staticDir)
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	habitRepo := repository.NewHabitRepository(db)
	userRepo := repository.NewUserRepository(db)
	authSvc := auth.NewService(cfg.JWT, cfg.App.CookieSecure)

	habitHandler := handlers.NewHabitHandler(habitRepo)
	completionHandler := handlers.NewCompletionHandler(habitRepo)
	authHandler := handlers.NewAuthHandler(userRepo, authSvc)

	rateLimiter := middleware.NewRateLimiter(10, time.Minute)

	router := api.NewRouter(api.RouterConfig{
		StaticDir:         staticDir,
		HabitHandler:      habitHandler,
		CompletionHandler: completionHandler,
		AuthHandler:       authHandler,
		AuthService:       authSvc,
		AllowedOrigin:     cfg.App.AppDomain,
		RateLimiter:       rateLimiter,
	})

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited successfully!")
}
