package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/lnardon/arete/internal/api/handlers"
	"github.com/lnardon/arete/internal/api/middleware"
	"github.com/lnardon/arete/internal/auth"
)

type RouterConfig struct {
	StaticDir         string
	HabitHandler      *handlers.HabitHandler
	CompletionHandler *handlers.CompletionHandler
	AuthHandler       *handlers.AuthHandler
	GoalHandler       *handlers.GoalHandler
	JournalHandler    *handlers.JournalHandler
	WhatsAppHandler   *handlers.WhatsAppHandler
	AuthService       *auth.Service
	AllowedOrigin     string
	RateLimiter       *middleware.RateLimiter
}

// Serves static files and falls back to index.html for SPA client-side routes.
func spaHandler(staticDir string) http.Handler {
	fs := http.FileServer(http.Dir(staticDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(staticDir, r.URL.Path)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}

		indexPath := filepath.Join(staticDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
		http.NotFound(w, r)
	})
}

func bodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
		next.ServeHTTP(w, r)
	})
}

func NewRouter(config RouterConfig) http.Handler {
	r := mux.NewRouter()
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS(config.AllowedOrigin))
	r.Use(middleware.Logging)
	r.Use(bodyLimit)

	jwtMiddleware := middleware.JWT(config.AuthService)
	rateLimitMiddleware := middleware.RateLimit(config.RateLimiter)

	// Public routes
	public := r.PathPrefix("/api/v1").Subrouter()
	public.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")
	public.Handle("/auth/register", rateLimitMiddleware(http.HandlerFunc(config.AuthHandler.Register))).Methods("POST")
	public.Handle("/auth/login", rateLimitMiddleware(http.HandlerFunc(config.AuthHandler.Login))).Methods("POST")
	public.HandleFunc("/auth/logout", config.AuthHandler.Logout).Methods("POST")
	// Evolution API can't send a JWT (or a custom header — see
	// WhatsAppHandler.Webhook) — this route is protected instead by a
	// shared-secret query param and, in production, by living on a private
	// network rather than being publicly routable.
	public.HandleFunc("/webhooks/whatsapp", config.WhatsAppHandler.Webhook).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Handle("/auth/me", jwtMiddleware(http.HandlerFunc(config.AuthHandler.Me))).Methods("GET")
	protected.Handle("/habits", jwtMiddleware(http.HandlerFunc(config.HabitHandler.ListHabits))).Methods("GET")
	protected.Handle("/habits", jwtMiddleware(http.HandlerFunc(config.HabitHandler.CreateHabit))).Methods("POST")
	protected.Handle("/habits/{id}", jwtMiddleware(http.HandlerFunc(config.HabitHandler.UpdateHabit))).Methods("PUT")
	protected.Handle("/habits/{id}", jwtMiddleware(http.HandlerFunc(config.HabitHandler.DeleteHabit))).Methods("DELETE")
	protected.Handle("/completions", jwtMiddleware(http.HandlerFunc(config.CompletionHandler.GetCompletions))).Methods("GET")
	protected.Handle("/completions/range", jwtMiddleware(http.HandlerFunc(config.CompletionHandler.GetCompletionsRange))).Methods("GET")
	protected.Handle("/completions/toggle", jwtMiddleware(http.HandlerFunc(config.CompletionHandler.ToggleCompletion))).Methods("POST")
	protected.Handle("/goals", jwtMiddleware(http.HandlerFunc(config.GoalHandler.ListGoals))).Methods("GET")
	protected.Handle("/goals", jwtMiddleware(http.HandlerFunc(config.GoalHandler.CreateGoal))).Methods("POST")
	protected.Handle("/goals/{id}/toggle", jwtMiddleware(http.HandlerFunc(config.GoalHandler.ToggleGoal))).Methods("PATCH")
	protected.Handle("/goals/{id}/progress", jwtMiddleware(http.HandlerFunc(config.GoalHandler.AddGoalProgress))).Methods("PATCH")
	protected.Handle("/goals/{id}", jwtMiddleware(http.HandlerFunc(config.GoalHandler.UpdateGoal))).Methods("PUT")
	protected.Handle("/goals/{id}", jwtMiddleware(http.HandlerFunc(config.GoalHandler.DeleteGoal))).Methods("DELETE")
	protected.Handle("/journal-entries", jwtMiddleware(http.HandlerFunc(config.JournalHandler.ListEntries))).Methods("GET")
	protected.Handle("/journal-entries/{date}", jwtMiddleware(http.HandlerFunc(config.JournalHandler.GetEntry))).Methods("GET")
	protected.Handle("/journal-entries/{date}", jwtMiddleware(http.HandlerFunc(config.JournalHandler.UpsertEntry))).Methods("PUT")
	protected.Handle("/journal-entries/{date}", jwtMiddleware(http.HandlerFunc(config.JournalHandler.DeleteEntry))).Methods("DELETE")
	protected.Handle("/whatsapp/link/code", jwtMiddleware(http.HandlerFunc(config.WhatsAppHandler.CreateLinkCode))).Methods("POST")
	protected.Handle("/whatsapp/status", jwtMiddleware(http.HandlerFunc(config.WhatsAppHandler.Status))).Methods("GET")
	protected.Handle("/whatsapp/link", jwtMiddleware(http.HandlerFunc(config.WhatsAppHandler.Unlink))).Methods("DELETE")

	// Serve frontend SPA
	if config.StaticDir != "" {
		r.PathPrefix("/").Handler(spaHandler(config.StaticDir)).Methods("GET", "OPTIONS")
	}

	return r
}
